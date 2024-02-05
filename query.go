package densify

import (
	"errors"
	"fmt"
	"strings"
)

type DensifyAPIQuery struct {
	AnalysisTechnology string // aws, azure, gcp, k8s
	AccountName        string // account name to look for
	AccountNumber      string // account number to look for
	SystemName         string // the entity name to pull recommendations for
	SkipErrors         bool   // skip/ignore errors

	K8sCluster        string // the k8s cluster to look for
	K8sNamespace      string // the k8s namespace to look for
	K8sPodName        string // the k8s pod name to look for
	K8sContainerName  string // the k8s container name to look for (optional)
	K8sControllerType string // the controller type used; ex. Deployment

	FallbackInstance   string // the fallback instance type in case there is no recommendation yet
	FallbackCPURequest string // the fallback CPU Request in case there is no recommendation yet
	FallbackMemRequest string // the fallback CPU Limit in case there is no recommendation yet
	FallbackCPULimit   string // the fallback Memory Request in case there is no recommendation yet
	FallbackMemLimit   string // the fallback Memory Limit in case there is no recommendation yet
}

func (q *DensifyAPIQuery) setValuesToLowercase() {
	q.AnalysisTechnology = strings.ToLower(q.AnalysisTechnology)
	q.AccountName = strings.ToLower(q.AccountName)
	q.AccountNumber = strings.ToLower(q.AccountNumber)
	q.SystemName = strings.ToLower(q.SystemName)
	q.K8sCluster = strings.ToLower(q.K8sCluster)
	q.K8sNamespace = strings.ToLower(q.K8sNamespace)
	q.K8sPodName = strings.ToLower(q.K8sPodName)
	q.K8sControllerType = strings.ToLower(q.K8sControllerType)
}

// check if the query is for Kubernetes/containers
func (q *DensifyAPIQuery) isKubernetesRequest() bool {
	switch q.AnalysisTechnology {
	case "k8s":
		return true
	case "kubernetes":
		return true
	default:
		return false
	}
}

func (q *DensifyAPIQuery) validate() error {
	_, err := q.getURIPath()
	if err != nil {
		return err
	}
	// validate the query parameters passed are sufficient
	if q.isKubernetesRequest() {
		// k8s validation
		if q.K8sCluster == "" || q.K8sNamespace == "" || q.K8sControllerType == "" || q.K8sPodName == "" {
			return fmt.Errorf("query must have required k8s fields: cluster, namespace, controllerType, podName, containerName")
		}
		if !q.isValidControllerType() {
			return fmt.Errorf("query controller type must be valid: pod, deployment, replicaset, daemonset, statefulset, cronjob, job")
		}
	} else {
		// cloud validation
		if q.SystemName == "" {
			return fmt.Errorf("query must have System Name")
		}
		if q.AccountNumber == "" && q.AccountName == "" {
			return fmt.Errorf("query must have Account Name or Account Number")
		}
	}
	// no errors means it's a valid looking query
	return nil
}

func (q *DensifyAPIQuery) isValidControllerType() bool {
	// check the controller types
	switch strings.ToLower(q.K8sControllerType) {
	case "deployment":
		return true
	case "":
		return true
	case "daemonset":
		return true
	case "replicaset":
		return true
	case "statefulset":
		return true
	case "pod":
		return true
	case "cronjob":
		return true
	case "job":
		return true
	default:
		return false
	}
}

// returns the Densify API analysis path based on the technology platform used, ex. aws, azure, gcp, kubernetes
func (q *DensifyAPIQuery) getURIPath() (string, error) {
	resp := ""
	switch q.AnalysisTechnology {
	case "aws":
		resp = "/analysis/cloud/aws"
	case "azure":
		resp = "/analysis/cloud/azure"
	case "gcp":
		resp = "/analysis/cloud/gcp"
	case "k8s":
		resp = "/analysis/containers/kubernetes"
	case "kubernetes":
		resp = "/analysis/containers/kubernetes"
	default:
		return "", errors.New("invalid tech value provided; must be one of the following: aws, azure, gcp, kubernetes, k8s")
	}
	return resp, nil
}
