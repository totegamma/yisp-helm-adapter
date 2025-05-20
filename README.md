# yisp-helm-adapter


## example
```yaml
!yisp
- import
- ["x", "https://raw.githubusercontent.com/totegamma/yisp-helm-adapter/refs/heads/main/helm-chart.yaml"]

---
!yisp
- *x.helm-chart
- repo: "https://charts.concrnt.net/"
  release: "concrnt"
  version: "0.7.13"
  values:
    meilisearch:
      enabled: true
```

