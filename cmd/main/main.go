package main

import (
	"fmt"

	densifyClient "github.com/joelpereira/densify-api-cient-go/densifyClient"
)

func main() {
	baseURL := `https://instance.densify.com:8443`
	username := `email@org.com`
	password := `password`

	fmt.Printf("Logging in to: %s...\n", baseURL)
	response, err := densifyClient.Authenticate(baseURL, username, password)
	fmt.Printf("AUTHENTICATE: Response: %v, Error: '%v'\n\n", response, err)
	if err != nil {
		return
	}

	// response, err = client.RefreshToken()
	// fmt.Printf("REFRESH TOKEN: Response: %v, Error: '%v'\n\n", response, err)

	tech := "aws"
	analysisName := "analysis_name"

	analysis, err := densifyClient.GetAnalysis(tech, analysisName)
	if err != nil {
		fmt.Printf("GET ANALYSIS: ERROR: '%v'\n\n", err.Error())
		return
	}
	fmt.Printf("GET ANALYSIS: Response: AnalysisId: %s, the rest: %v\n\n", analysis.AnalysisId, analysis)

	recommendations, err := densifyClient.GetRecommendations(tech, analysis.AnalysisId)
	if err != nil {
		fmt.Printf("GET ANALYSIS: ERROR: '%v'\n\n", err.Error())
		return
	}
	// fmt.Println("Recommendations:::")
	// fmt.Println(recommendations)

	tf := densifyClient.ConvertRecommendationsToTF(recommendations)
	fmt.Println("TF format:::")
	fmt.Println(tf)
}
