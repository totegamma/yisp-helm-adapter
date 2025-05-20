package cmd

import (
	"strings"
    "os"
    "os/exec"
	"io"
    "bytes"
    "fmt"
    "regexp"
    "path/filepath"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var rootCmd = &cobra.Command{
	Use:   "yisp-helm-adapter",
	Short: "yisp-helm-adapter is a command line tool for YISP",
	Args: cobra.ExactArgs(3),
	Run:   Run,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}


func Run(cmd *cobra.Command, args []string) {

	repo := args[0]
	release := args[1]
	version := args[2]

    values := ""

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(fmt.Sprintf("Error reading from stdin: %s", err))
		}
		values = string(bytes)
	}

    workDir, err := os.MkdirTemp("", "helm-")
    if err != nil {
        panic(fmt.Sprintf("Error creating temp dir: %s", err))
    }

    defer func() {
        err := os.RemoveAll(workDir)
        if err != nil {
            panic(fmt.Sprintf("Error deleting workDir: %s", err))
        }
    }()

    cacheDir := chartDir
    if strings.HasPrefix(cacheDir, "~/") {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            panic(fmt.Sprintf("Error getting home dir: %s", err))
        }
        cacheDir = filepath.Join(homeDir, strings.TrimPrefix(cacheDir, "~/"))
    }

    _, err = getHelmVersion()
    if err != nil {
        panic(fmt.Sprintf("Error getting helm version: %s", err))
    }

    cacheFile, err := findHelmCache(release, version, cacheDir)
    if err != nil {
        fmt.Printf("No helm cache found. pulling chart...\n")

        mkdirErr := os.MkdirAll(cacheDir, 0755)
        if mkdirErr != nil {
            panic(fmt.Sprintf("Error creating cacheDir: %s", mkdirErr))
        }

        _, err = runHelmCommand(pullCommand(repo, release, version, cacheDir))
        if err != nil {
            panic(fmt.Sprintf("Error pulling helm chart: %s", err))
        }

        cacheFile, err = findHelmCache(release, version, cacheDir)
        if err != nil {
            panic(fmt.Sprintf("Error finding helm cache: %s", err))
        }
    }

    // untar chart to workDir
    fmt.Printf("Untarring chart %s to %s\n", cacheFile, workDir)
    _, err = exec.Command("tar", "xf", cacheFile, "-C", workDir).Output()
    if err != nil {
        panic(fmt.Sprintf("Error untarring helm chart: %s", err))
    }


    generated, err := runHelmCommand(templateCommand(release, values, workDir))
    if err != nil {
        panic(fmt.Sprintf("Error generating helm template: %s", err))
    }

    fmt.Printf("Generated helm template:\n%s\n", string(generated))
}


// ---


const (
    helmCmd = "helm"
    chartDir = "~/.cache/helmcharts"
)

func runHelmCommand(args []string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(helmCmd, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	errorOutput := stderr.String()
	if err != nil {
        panic(fmt.Sprintf("Error running helm command: %s\n%s", err, errorOutput))
	}
	return stdout.Bytes(), err
}

func getHelmVersion() (string, error) {
	stdout, err := runHelmCommand([]string{"version", "-c", "--short"})
	if err != nil {
		return "", err
	}
	r, err := regexp.Compile(`v?\d+(\.\d+)+`)
	if err != nil {
		return "", err
	}
	v := r.FindString(string(stdout))
	if v == "" {
		return "", fmt.Errorf("cannot find version string in %s", string(stdout))
	}

	return v, nil
}

func pullCommand(repo, name, version, cacheDir string) []string {
	args := []string{
		"pull",
        "--repo", repo, name,
        "-d", cacheDir,
	}

	if version != "" {
		args = append(args, "--version", version)
	}

	return args
}

func templateCommand(name, values, workDir string) []string {

    args := []string{
        "template",
        filepath.Join(workDir, name),
    }

    if values != "" {
        valuesFile := filepath.Join(workDir, "values.yaml")
        f, err := os.Create(valuesFile)
        if err != nil {
            panic(fmt.Sprintf("Error creating values file: %s", err))
        }
        defer f.Close()
        _, err = f.WriteString(values)
        if err != nil {
            panic(fmt.Sprintf("Error writing values file: %s", err))
        }
        args = append(args, "-f", valuesFile)
    }

    return args
}

func findHelmCache(name string, version string, basePath string) (string, error) {
    cacheName := name + "-"
    if version != "" {
        cacheName += version + "*"
    } else {
        cacheName += "*"
    }

    fmt.Printf("Cache name: %s\n", cacheName)

    // find cache by glob
    files, err := filepath.Glob(filepath.Join(basePath, cacheName))
    if err != nil {
        panic(fmt.Sprintf("Error finding helm cache: %s", err))
    }

    if len(files) == 0 {
        return "", fmt.Errorf("No helm cache found for %s", name)
    }

    return files[0], nil
}


