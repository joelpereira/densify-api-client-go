package main

import (
	"fmt"

	"github.com/joelpereira/densify-api-cient-go"
)

func main() {
	instanceURL := `https://instance.densify.com:443`
	username := `email@org.com`
	password := `password`

	fmt.Printf("Logging in to: %s...\n", instanceURL)
	client, err := densify.NewClient(&instanceURL, &username, &password)
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

	// set values
	densifyAPIQuery := densify.DensifyAPIQuery{
		AnalysisTechnology:   "aws/azure/gcp/k8s",
		AccountOrClusterName: "account-name",
		EntityName:           "system-name",
		// if it's a kubernetes resource:
		K8sNamespace:      "namespace",
		K8sPodName:        "podname",
		K8sControllerType: "deployment/daemonset/statefulset",
	}
	err = client.ConfigureQuery(&densifyAPIQuery)
	if err != nil {
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

	// recommendations, err := client.GetRecommendations()
	// if err != nil {
	// 	fmt.Printf("GET ANALYSIS: ERROR: '%v'\n\n", err.Error())
	// 	return
	// }

	// fmt.Println("Recommendations:::")
	// fmt.Println(recommendations)

	// tf := client.ConvertRecommendationsToTF(recommendations)
	// fmt.Println("TF format:::")
	// fmt.Println(tf)

	// check if token is expired
	// fmt.Println(client.IsTokenExpired())
}
