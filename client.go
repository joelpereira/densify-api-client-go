package densify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// type config struct {
// 	baseURL string
// 	userName string
// 	password string
// }

// type connection struct {
// 	Token string
// 	TokenExpiry int
// }

// var client = &http.Client{Timeout: 60 * time.Second}
// var baseURL string
// var apiUserName string
// var apiPassword string
// var apiToken string
// var apiTokenExpiry int64

type Client struct {
	HTTPClient     *http.Client
	BaseURL        string
	ApiUserName    string
	ApiPassword    string
	ApiToken       string
	ApiTokenExpiry int64
}

type AuthResponse struct {
	ApiToken string
	Expires  int64
	Status   int
	Message  string
}
type AuthError struct {
	/* variables */
}

// NewClient -
func NewClient(instanceURL, username, password *string) (*Client, error) {
	pre := ""
	if !strings.HasPrefix(strings.ToLower(*instanceURL), "http") {
		pre = `https://`
	}

	c := Client{
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}

	// return c.getToken()

	if instanceURL != nil {
		c.BaseURL = fmt.Sprintf("%s%s%s", pre, strings.ToLower(*instanceURL), "/api/v2")
	}

	// If username or password not provided, return empty client
	if username == nil || password == nil {
		return &c, nil
	}

	c.ApiUserName = *username
	c.ApiPassword = *password

	_, err := c.GetNewToken()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// func (c *Client) getToken(instanceURL string, username string, password string) (string, error) {
func (c *Client) GetNewToken() (*AuthResponse, error) {
	urlAuth := fmt.Sprintf("%s%s", c.BaseURL, "/authorize")

	postBody, _ := json.Marshal(map[string]string{
		"userName": c.ApiUserName,
		"pwd":      c.ApiPassword,
	})
	request, error := http.NewRequest("POST", urlAuth, bytes.NewBuffer(postBody))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if error != nil {
		return nil, error
	}
	// client := &http.Client{}
	// client = http.Client{Timeout: timeout}
	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	// check if the http call was successful (200)
	if response.StatusCode != 200 {
		return nil, fmt.Errorf(`auth request received error: %s`, response.Status)
	}
	defer response.Body.Close()

	//Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		// log.Fatalln(err)
		return nil, err
	}

	var authResponse AuthResponse
	err = json.Unmarshal(body, &authResponse)
	// Check for errors
	if err != nil {
		return nil, errors.New("JSON decode error: " + err.Error())
	}

	c.ApiToken = authResponse.ApiToken
	c.ApiTokenExpiry = authResponse.Expires

	retMsg := ""
	if authResponse.Message != "" {
		retMsg = fmt.Sprintf("%v - %v", authResponse.Status, authResponse.Message)
	}
	fmt.Println(retMsg)

	return &authResponse, nil
}

// func (c *Client) Authenticate(instanceURL string, username string, password string) (string, error) {
// 	pre := ""
// 	if !strings.HasPrefix(strings.ToLower(instanceURL), "http") {
// 		pre = `https://`
// 	}
// 	c.BaseURL = fmt.Sprintf("%s%s%s", pre, strings.ToLower(instanceURL), "/api/v2")
// 	c.ApiUserName = username
// 	c.ApiPassword = password
// 	return c.getToken()
// }

// func (c *Client) RefreshToken() (string, error) {
// 	return c.getToken()
// }

func (c *Client) GetAnalysis(tech string, analysisName string) (*DensifyAnalysis, error) {
	// retVal := models.ResponseAnalysis{}
	urlAnalyses, err := c.validateTech(tech)
	if err != nil {
		return nil, err
	}
	// switch tech {
	// case "aws":
	// 	urlAnalyses = "/analysis/cloud/aws"
	// case "azure":
	// 	urlAnalyses = "/analysis/cloud/azure"
	// case "gcp":
	// 	urlAnalyses = "/analysis/cloud/gcp"
	// case "k8s":
	// 	urlAnalyses = "/analysis/containers/kubernetes"
	// default:
	// 	return nil, "Invalid tech value provided; must be one of the following: aws, azure, gcp, k8s"
	// }

	url := fmt.Sprintf("%s%s", c.BaseURL, urlAnalyses)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle error
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiToken))
	req.Header.Set("Accept", "application/json")

	// resp, err := http.DefaultClient.Do(req)
	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	//Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var analyses []DensifyAnalysis
	err = json.Unmarshal(body, &analyses)
	// Check for errors
	if err != nil {
		return nil, errors.New("JSON decode error: " + err.Error())
	}
	var retAnalysis DensifyAnalysis
	retErr := ""
	analysisName = strings.ToLower(analysisName)
	analysisFound := false
	for i := 0; i < len(analyses); i++ {
		if strings.ToLower(analyses[i].AnalysisName) == analysisName {
			retAnalysis = analyses[i]
			i = len(analyses)
			analysisFound = true
		}
	}
	// if nothing was found, throw an error message with the list of analyses names
	if !analysisFound {
		retErr = "no analysis found with that name. Existing analysis names are:\n"
		for i := 0; i < len(analyses); i++ {
			retErr = fmt.Sprintf("%s%s\n", retErr, analyses[i].AnalysisName)
		}
		return nil, errors.New(retErr)
	}
	return &retAnalysis, nil
}

func (c *Client) GetRecommendations(tech string, analysisId string) (*[]DensifyRecommendations, error) {
	// check that output is either json/terraform
	techUrl, err := c.validateTech(tech)
	if err != nil {
		return nil, err
	}
	// outputFormat, err := validateOutputFormat(output)
	// if err != nil {
	// 	return nil, err
	// }

	url := fmt.Sprintf("%s%s/%s/results", c.BaseURL, techUrl, analysisId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle error
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiToken))
	req.Header.Set("Cache-Control", "no-cache")
	// req.Header.Set("Accept", outputFormat)
	req.Header.Set("Accept", "application/json")

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	//Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var recos []DensifyRecommendations
	err = json.Unmarshal(body, &recos)
	// Check for errors
	if err != nil {
		return nil, errors.New("JSON decode error: " + err.Error())
	}
	// specify the type, cloud/container, within each obj/reco
	count := len(recos)
	for i := 0; i < count; i++ {
		if tech == "k8s" || tech == "kubernetes" {
			recos[i].AnalysisType = "containers"
		} else {
			recos[i].AnalysisType = "cloud"
		}
		recos[i].AnalysisTechnology = tech
	}
	return &recos, nil
}

func (c *Client) validateTech(tech string) (string, error) {
	resp := ""
	switch tech {
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
		return "", errors.New("invalid tech value provided; must be one of the following: aws, azure, gcp, k8s")
	}
	return resp, nil
}

// json or terraform
// func validateOutputFormat(output string) (string, string) {
// 	resp := ""
// 	err := ""
// 	switch output {
// 	case "json":
// 		resp = "application/json"
// 	case "terraform":
// 		resp = "application/terraform-map"
// 	default:
// 		err = "Invalid output value provided; must be one of the following: json, terraform"
// 	}
// 	return resp, err
// }

func (c *Client) ConvertRecommendationsToTF(recommendations *[]DensifyRecommendations) string {
	return c.ConvertRecommendationsToTFWithVarName(recommendations, "densify_recommendations")
}

func (c *Client) ConvertRecommendationsToTFWithVarName(recommendations *[]DensifyRecommendations, tfVarName string) string {
	var sb strings.Builder
	sb.WriteString(tfVarName + " = {")
	count := len(*recommendations)
	newline := "\n"
	for i := 0; i < count; i++ {
		reco := (*recommendations)[i]
		if reco.AnalysisType == "cloud" {
			sb.WriteString(fmt.Sprintf(`  "%s" {%s`, reco.Name, newline))
			sb.WriteString(fmt.Sprintf(`    analysisType="%s"%s`, reco.AnalysisType, newline))
			sb.WriteString(fmt.Sprintf(`    analysisTechnology="%s"%s`, reco.AnalysisTechnology, newline))
			sb.WriteString(fmt.Sprintf(`    accountIdRef="%s"%s`, reco.AccountIdRef, newline))
			sb.WriteString(fmt.Sprintf(`    region="%s"%s`, reco.Region, newline))
			sb.WriteString(fmt.Sprintf(`    serviceType="%s"%s`, reco.ServiceType, newline))
			sb.WriteString(fmt.Sprintf(`    currentType="%s"%s`, reco.CurrentType, newline))
			sb.WriteString(fmt.Sprintf(`    recommendationType="%s"%s`, reco.RecommendationType, newline))
			sb.WriteString(fmt.Sprintf(`    currentType="%s"%s`, reco.CurrentType, newline))
			sb.WriteString(fmt.Sprintf(`    recommendedType="%s"%s`, reco.RecommendedType, newline))
			sb.WriteString(fmt.Sprintf(`    powerState="%s"%s`, reco.PowerState, newline))
			sb.WriteString(fmt.Sprintf(`    predictedUptime="%s"%s`, ConvertFloatToStr(reco.PredictedUptime), newline))
			sb.WriteString(fmt.Sprintf(`    implementationMethod="%s"%s`, reco.ImplementationMethod, newline))
			sb.WriteString(fmt.Sprintf(`    approvalTypecurrentType="%s"%s`, reco.ApprovalType, newline))
			sb.WriteString(fmt.Sprintf(`    savingsEstimate="%s"%s`, ConvertFloatToStr(reco.SavingsEstimate), newline))
			sb.WriteString(fmt.Sprintf(`    effortEstimate="%s"%s`, reco.EffortEstimate, newline))
			sb.WriteString(fmt.Sprintf(`    densifyPolicy="%s"%s`, reco.DensifyPolicy, newline))
		} else if reco.AnalysisType == "containers" {
			// TODO::: Complete the container recos
			sb.WriteString(fmt.Sprintf(`  "%s" {%s`, reco.Name, newline))
			sb.WriteString(fmt.Sprintf(`    analysisType="%s"%s`, reco.AnalysisType, newline))
		}
		sb.WriteString(`  }` + newline)
	}
	sb.WriteString("}")

	return sb.String()
}