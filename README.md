# yisp-helm-adapter

## example
```yaml
!yisp
- import
- ["helmchart", "https://raw.githubusercontent.com/totegamma/yisp-helm-adapter/refs/heads/main/helm-chart.yisp"]

---
!yisp
- *helmchart.template
- repo: "https://charts.concrnt.net/"
  release: "concrnt"
  version: "0.7.13"
  values:
    meilisearch:
      enabled: true
```

