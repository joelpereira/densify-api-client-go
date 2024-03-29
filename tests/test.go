package main

import (
	"fmt"
	"os"

	"github.com/joelpereira/densify-api-client-go"
)

func main() {
	instanceURL := `https://instance.densify.com:443`
	username := `user@xyz.com`
	password := `password`

	// query
	densifyAPIQuery := densify.DensifyAPIQuery{
		AnalysisTechnology: "aws/azure/gcp/kubernetes/k8s",
		AccountNumber:      "account-num",
		// or:
		AccountName: "account-name",
		SystemName:  "system-name",
		// FallbackInstance: "m6i.large",

		// if it's a kubernetes resource:
		// K8sCluster:        "cluster",
		// K8sNamespace:      "namespace",
		// K8sPodName:        "podname",
		// K8sControllerType: "deployment/daemonset/statefulset/cronjob",
	}

	numArgs := len(os.Args)
	if numArgs < 7 {
		fmt.Println("Insufficient args provided: ", numArgs)
		return
	}
	instanceURL = os.Args[1]
	username = os.Args[2]
	password = os.Args[3]
	tech := os.Args[4]
	// cloud
	account := os.Args[5]
	system := os.Args[6]
	// container
	cluster := ""
	namespace := ""
	controller := ""
	pod := ""
	container := ""
	if numArgs > 8 {
		cluster = os.Args[5]
		namespace = os.Args[6]
		controller = os.Args[7]
		pod = os.Args[8]
		if numArgs > 9 {
			container = os.Args[9]
		}
	}

	fmt.Printf("Logging in to: %s...\n", instanceURL)
	client, err := densify.NewDensifyClient(&instanceURL, &username, &password, 0)
	if err != nil {
		fmt.Printf("ERROR: '%v'\n\n", err)
		return
	}
	fmt.Printf("NEW CLIENT: Response: %v, Error: '%v'\n\n", client.ApiToken, err)
	if err != nil {
		return
	}

	// response, err = client.RefreshToken()
	// fmt.Printf("REFRESH TOKEN: Response: %v, Error: '%v'\n\n", response, err)

	// governance
	densifyAPIQuery = densify.DensifyAPIQuery{
		AnalysisTechnology: tech,
		// AccountName:        "Mobile Services (Pay-Go)",
		AccountNumber:     account,
		SystemName:        system,
		K8sCluster:        cluster,
		K8sNamespace:      namespace,
		K8sControllerType: controller,
		K8sPodName:        pod,
		K8sContainerName:  container,

		SkipErrors: true,
		// FallbackInstance: "m6i.large",
		FallbackCPURequest: "123m",
		FallbackCPULimit:   "321m",
		FallbackMemRequest: "234Mi",
		FallbackMemLimit:   "432Mi",
	}

	err = client.ConfigureQuery(&densifyAPIQuery)
	if err != nil {
		fmt.Printf("Configure Query: ERROR: '%v'\n\n", err.Error())
		return
	}

	analyses, err := client.GetAccountOrCluster()
	if err != nil {
		fmt.Printf("GET ACCOUNT(S): ERROR: '%v'\n\n", err.Error())
		return
	}
	fmt.Printf("GET ACCOUNT(S): Response: Count: %d\nAccount(s): %v\n\n", len(*analyses), analyses)

	recommendation, err := client.GetDensifyRecommendation()
	if err != nil {
		fmt.Printf("GET RECOMMENDATION: ERROR: '%v'\n\n", err.Error())
		return
	}
	fmt.Printf("GET RECOMMENDATION: '%v'\n\n", recommendation)
	if len(recommendation.Containers) >= 1 {
		fmt.Printf("Recommendation value: '%v'\n\n", recommendation.Containers[0].FallbackCpuRequest)
		fmt.Printf("Recommendation value: '%v'\n\n", recommendation.Containers[0].FallbackCpuLimit)
		fmt.Printf("Recommendation value: '%v'\n\n", recommendation.Containers[0].FallbackMemRequest)
		fmt.Printf("Recommendation value: '%v'\n\n", recommendation.Containers[0].FallbackMemLimit)
	}

	// INSTANCE GOVERNANCE
	err = client.LoadDensifyGuardrailsAllInstances(recommendation, 1.2)
	if err != nil {
		fmt.Printf("GET INSTANCE GOVERNANCE: ERROR: '%v'\n\n", err.Error())
		return
	}
	fmt.Printf("GET INSTANCE GOVERNANCE (GUARDRAILS): '%v'\n\n", recommendation.Guardrails)

	l, err := recommendation.GetGuardrailsOutsideSpendTolerance()
	if err != nil {
		fmt.Printf("InstanceGovernance ERROR: '%v'\n\n", err.Error())
		return
	}
	fmt.Printf("GET INSTANCE GOVERNANCE: '%v'\n\n", l)
	fmt.Printf("GET INSTANCE GOVERNANCE Total Length: '%d'\n\n", l.TotalLength())
	fmt.Printf("GET INSTANCE GOVERNANCE Sorted Keys: %v\n\n", l.GetSortedScoreList())
	fmt.Printf("GET INSTANCE GOVERNANCE Lowest Score: '%d'\n\n", l.GetMinScore())
	fmt.Printf("GET INSTANCE GOVERNANCE Highest Score: '%d'\n\n", l.GetMaxScore())

	// tf := client.ConvertRecommendationsToTF(recommendations)
	// fmt.Println("Terraform Format:")
	// fmt.Println(tf)

	// check if token is expired
	// fmt.Println(client.IsTokenExpired())
}
