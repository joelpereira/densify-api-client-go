package main

import (
	"fmt"

	"github.com/joelpereira/densify-api-cient-go"
)

func main() {
	instanceURL := `https://partner1.densify.com:443`
	username := `jpereira@densify.com`
	password := `Jp3r31r@Jp3r31r@6`
	// instanceURL := `https://instance.densify.com:443`
	// username := `email@org.com`
	// password := `password`

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

	tech := "aws"
	// analysisName := "analysis_name"
	analysisName := "922390019409 (Mobile_Prod)"

	// set values
	client.SetQuery(tech, analysisName)

	analysis, err := client.GetAnalysis()
	if err != nil {
		fmt.Printf("GET ANALYSIS: ERROR: '%v'\n\n", err.Error())
		return
	}
	fmt.Printf("GET ANALYSIS: Response: AnalysisId: %s, the rest: %v\n\n", analysis.AnalysisId, analysis)

	recommendations, err := client.GetRecommendations()
	if err != nil {
		fmt.Printf("GET ANALYSIS: ERROR: '%v'\n\n", err.Error())
		return
	}

	// fmt.Println("Recommendations:::")
	// fmt.Println(recommendations)

	tf := client.ConvertRecommendationsToTF(recommendations)
	fmt.Println("TF format:::")
	fmt.Println(tf)

	// check if token is expired
	fmt.Println(client.IsTokenExpired())
}
