# Densify API Client Module for Go (golang)

## How to Use
You can import the Densify Module using the following sample code below.

### Import
```go
import (
	"github.com/joelpereira/densify-api-cient-go"
)
```

### Authenticate
```go
client, err := client.NewClient(baseURL, username, password)
if err != nil {
    return
}
```

### Configure Query
```go
densifyAPIQuery := densify.DensifyAPIQuery{
    AnalysisTechnology:   "aws/azure/gcp/k8s",
    AccountOrClusterName: "account-name",
    EntityName:           "system-name",
    // if it's a kubernetes resource:
    K8sNamespace:         "namespace",
    K8sPodName:           "podname",
    K8sControllerType:    "deployment/daemonset/statefulset",
}
err = client.ConfigureQuery(&densifyAPIQuery)
if err != nil {
    return
}
```

### Pull Analysis
```go
analysis, err := client.GetAccountOrCluster()
if err != nil {
    return
}
```

### Pull the single Recommendation based on the query
```go
recommendations, err := client.GetDensifyRecommendation()
if err != nil {
    return
}
```
