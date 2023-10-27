# Densify API Client Module for Go (golang)

## How to Use
You can import the Densify Module using the following sample code below.

### Import
```go
import (
	"densify.com/api/client"
)
```

### Authenticate
```go
response, err := client.Authenticate(baseURL, username, password)
if err != nil {
    return
}
```

### Pull Analysis
```go
analysis, err := client.GetAnalysis(tech, analysisName)
if err != nil {
    return
}
```

### Pull the Recommendations within the specified analysis
```go
recommendations, err := client.GetRecommendations(tech, analysis.AnalysisId)
if err != nil {
    return
}
```
