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

	// Densify Query
	Query *DensifyAPIQuery

	// other values to store in-between API calls

	AnalysisIds []string // store the analysis ids that make up the account or cluster (which can be separated across multiple analyses)
}

type DensifyAPIQuery struct {
	AnalysisTechnology string // aws, azure, gcp, k8s
	AccountName        string // account name to look for
	AccountNumber      string // account number to look for
	SystemName         string // the entity name to pull recommendations for

	K8sCluster        string // the k8s cluster to look for
	K8sNamespace      string // the k8s namespace to look for
	K8sPodName        string // the k8s pod name to look for
	K8sContainerName  string // the k8s container name to look for
	K8sControllerType string // the controller type used; ex. Deployment

	FallbackInstance   string
	FallbackCPURequest string
	FallbackMemRequest string
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
	if instanceURL == nil || username == nil || password == nil {
		return nil, fmt.Errorf(`instanceURL, username, password cannot be empty`)
	}

	pre := ""
	if !strings.HasPrefix(strings.ToLower(*instanceURL), "http") {
		pre = `https://`
	}

	c := Client{
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}

	c.BaseURL = fmt.Sprintf("%s%s%s", pre, strings.ToLower(*instanceURL), "/api/v2")
	c.ApiUserName = *username
	c.ApiPassword = *password

	_, err := c.GetNewAuthToken()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) ConfigureQuery(query *DensifyAPIQuery) error {
	// validate the query has all the required values
	if query == nil {
		return fmt.Errorf("query cannot be empty/nil")
	}
	// validate the technology is valid
	err := query.validate()
	if err != nil {
		return err
	}

	// query looks valid; let's lowercase all the values first
	query.setValuesToLowercase()

	c.Query = query
	// reset other fields
	c.AnalysisIds = []string{}

	return nil // no error
}

// func (c *Client) getToken(instanceURL string, username string, password string) (string, error) {
func (c *Client) GetNewAuthToken() (*AuthResponse, error) {
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

func (c *Client) GetAccountOrCluster() (*[]DensifyAnalysis, error) {
	// make sure a query has been defined
	if c.Query == nil {
		return nil, fmt.Errorf("you must specify a query first")
	}

	urlAnalyses, err := c.Query.getURIPath()
	if err != nil {
		return nil, err
	}

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

	analyses := []DensifyAnalysis{}
	err = json.Unmarshal(body, &analyses)
	// Check for errors
	if err != nil {
		return nil, errors.New("JSON decode error: " + err.Error())
	}
	retAnalyses := []DensifyAnalysis{}
	retErr := ""
	accountName := strings.ToLower(c.Query.AccountName)
	accountNumber := strings.ToLower(c.Query.AccountNumber)
	clusterName := strings.ToLower(c.Query.K8sCluster)
	found := false
	isKubernetesRequest := c.Query.isKubernetesRequest()
	for i := 0; i < len(analyses); i++ {
		// if it's a kubernetes/container request, look at analysis name, and check if it contains the cluster string instead
		if isKubernetesRequest {
			if strings.Contains(strings.ToLower(analyses[i].AnalysisName), clusterName) {
				retAnalyses = append(retAnalyses, analyses[i])
				found = true
			}
		} else { // else, look at cloud account
			// search by account number
			if strings.Contains(strings.ToLower(analyses[i].AccountId), accountNumber) || strings.Contains(strings.ToLower(analyses[i].AccountName), accountName) {
				retAnalyses = append(retAnalyses, analyses[i])
				found = true
			}
		}
	}
	// if nothing was found, throw an error message with the list of analyses names
	if !found {
		qn := c.Query.AccountNumber
		if isKubernetesRequest {
			qn = c.Query.K8sCluster
		}

		retErr = fmt.Sprintf(`no account or cluster found with the name '%s'. Existing names are:\n`, qn)
		for i := 0; i < len(analyses); i++ {
			if isKubernetesRequest {
				retErr = fmt.Sprintf("%s\"%s\"\n", retErr, analyses[i].AnalysisId)
			} else {
				retErr = fmt.Sprintf("%s\"%s\"\n", retErr, analyses[i].AccountId)
			}
		}
		return nil, errors.New(retErr)
	}
	// set the analysis ids as well
	for i := 0; i < len(retAnalyses); i++ {
		c.AnalysisIds = append(c.AnalysisIds, retAnalyses[i].AnalysisId)
	}
	// c.AnalysisIds = retAnalysis.AnalysisId
	return &retAnalyses, nil
}

// pull the recommendations and look for a specific entity in the list
func (c *Client) GetDensifyRecommendation() (*DensifyRecommendation, error) {
	// make sure a query has been defined
	if c.Query == nil {
		return nil, fmt.Errorf("you must specify a query first")
	}
	err := c.Query.validate()
	if err != nil {
		return nil, err
	}

	isKubernetesRequest := c.Query.isKubernetesRequest()
	recos, err := c.GetDensifyRecommendations()
	if err != nil {
		return nil, err
	}
	// go through the list of recommendations and look for the entity name provided
	count := len(*recos)
	for i := 0; i < count; i++ {
		if isKubernetesRequest {
			// check the namespace and pod name as well
			recoName := strings.ToLower((*recos)[i].Container)
			recoNamespace := strings.ToLower((*recos)[i].Namespace)
			recoPodName := strings.ToLower((*recos)[i].PodService)
			recoControllerType := strings.ToLower((*recos)[i].ControllerType)
			if recoNamespace == c.Query.K8sNamespace && recoControllerType == c.Query.K8sControllerType && recoPodName == c.Query.K8sPodName && recoName == c.Query.K8sContainerName {
				reco := (*recos)[i]
				return &reco, nil
			}
		} else {
			recoName := strings.ToLower((*recos)[i].Name)
			if recoName == c.Query.SystemName {
				reco := (*recos)[i]
				// check if the ApprovedType needs a fallback
				if reco.ApprovedType == "" {
					reco.ApprovedType = c.Query.FallbackInstance
				}
				return &reco, nil
			}
		}
	}

	// return a different error msg if it's a cloud vs k8s query
	if isKubernetesRequest {
		return nil, fmt.Errorf(`could not find a Densify recommendation for container (%s) in namespace (%s), controller (%s), pod name (%s)`, c.Query.K8sContainerName, c.Query.K8sNamespace, c.Query.K8sControllerType, c.Query.K8sPodName)
	} else {
		return nil, fmt.Errorf("could not find a Densify recommendation named: %s", c.Query.SystemName)
	}
}

// func (c *Client) GetRecommendations(tech string, analysisId string) (*[]DensifyRecommendations, error) {
func (c *Client) GetDensifyRecommendations() (*[]DensifyRecommendation, error) {
	// make sure a query has been defined
	if c.Query == nil {
		return nil, fmt.Errorf("you must specify a query first")
	}
	// check if we have an AnalysisId
	if c.AnalysisIds == nil || len(c.AnalysisIds) == 0 {
		return nil, fmt.Errorf(`no Densify analyses found; make sure you call GetAccountOrCluster() first`)
	}

	// check that output is either json/terraform
	techUrl, err := c.Query.getURIPath()
	if err != nil {
		return nil, err
	}

	// pull recommendations for each of the analyses
	var retRecos []DensifyRecommendation
	for x := 0; x < len(c.AnalysisIds); x++ {
		url := fmt.Sprintf("%s%s/%s/results", c.BaseURL, techUrl, c.AnalysisIds[x])
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			// handle error
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiToken))
		req.Header.Set("Cache-Control", "no-cache")
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

		var recos []DensifyRecommendation
		err = json.Unmarshal(body, &recos)
		// Check for errors
		if err != nil {
			return nil, errors.New("JSON decode error: " + err.Error())
		}

		// add some additional parameters that are not returned in the API call
		count := len(recos)
		for i := 0; i < count; i++ {
			if c.Query.isKubernetesRequest() {
				recos[i].AnalysisType = "containers"
			} else {
				recos[i].AnalysisType = "cloud"
			}
			recos[i].AnalysisTechnology = c.Query.AnalysisTechnology
			recos[i].AccountId = c.Query.AccountNumber
			recos[i].AccountName = c.Query.AccountName
			recos[i].ApprovedType = recos[i].RecommendedType
		}
		// now we copy the recommendations into the retRecos slice
		retRecos = append(retRecos, recos...)
	}
	return &retRecos, nil
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
		if q.K8sCluster == "" || q.K8sNamespace == "" || q.K8sControllerType == "" || q.K8sPodName == "" || q.K8sContainerName == "" {
			return fmt.Errorf("query must have required k8s fields: cluster, namespace, controllerType, podName, containerName")
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

func (c *Client) IsTokenExpired() bool {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	return now >= c.ApiTokenExpiry
}

func (c *Client) ConvertRecommendationsToTF(recommendations *[]DensifyRecommendation) string {
	return c.ConvertRecommendationsToTFWithVarName(recommendations, "densify_recommendations")
}

func (c *Client) ConvertRecommendationsToTFWithVarName(recommendations *[]DensifyRecommendation, tfVarName string) string {
	var sb strings.Builder
	sb.WriteString(tfVarName + " = {")
	count := len(*recommendations)
	newline := "\n"
	for i := 0; i < count; i++ {
		reco := (*recommendations)[i]
		// if reco.AnalysisType == "cloud" {
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
		// } else
		if reco.AnalysisType == "containers" {
			sb.WriteString(fmt.Sprintf(`  "%s" {%s`, reco.Name, newline))
			sb.WriteString(fmt.Sprintf(`    cluster="%s"%s`, reco.Cluster, newline))
			sb.WriteString(fmt.Sprintf(`    container="%s"%s`, reco.Container, newline))
			sb.WriteString(fmt.Sprintf(`    controllerType="%s"%s`, reco.ControllerType, newline))
			sb.WriteString(fmt.Sprintf(`    namespace="%s"%s`, reco.Namespace, newline))
			sb.WriteString(fmt.Sprintf(`    podService="%s"%s`, reco.PodService, newline))
			sb.WriteString(fmt.Sprintf(`    estimatedSavings="%s"%s`, ConvertFloatToStr(reco.EstimatedSavings), newline))
			sb.WriteString(fmt.Sprintf(`    totalNetSavings="%s"%s`, ConvertFloatToStr(reco.TotalNetSavings), newline))
			sb.WriteString(fmt.Sprintf(`    displayName="%s"%s`, reco.DisplayName, newline))
			sb.WriteString(fmt.Sprintf(`    currentCount="%d"%s`, reco.CurrentCount, newline))
			sb.WriteString(fmt.Sprintf(`    currentCpuRequest="%d"%s`, reco.CurrentCpuRequest, newline))
			sb.WriteString(fmt.Sprintf(`    currentCpuLimit="%d"%s`, reco.CurrentCpuLimit, newline))
			sb.WriteString(fmt.Sprintf(`    currentMemRequest="%d"%s`, reco.CurrentMemRequest, newline))
			sb.WriteString(fmt.Sprintf(`    currentMemLimit="%d"%s`, reco.CurrentMemLimit, newline))
			sb.WriteString(fmt.Sprintf(`    recommendedCpuRequest="%d"%s`, reco.RecommendedCpuRequest, newline))
			sb.WriteString(fmt.Sprintf(`    recommendedCpuLimit="%d"%s`, reco.RecommendedCpuLimit, newline))
			sb.WriteString(fmt.Sprintf(`    recommendedMemRequest="%d"%s`, reco.RecommendedMemRequest, newline))
			sb.WriteString(fmt.Sprintf(`    recommendedMemLimit="%d"%s`, reco.RecommendedMemLimit, newline))
			sb.WriteString(fmt.Sprintf(`    runningHours="%d"%s`, reco.RunningHours, newline))
		}
		sb.WriteString(`  }` + newline)
	}
	sb.WriteString("}")

	return sb.String()
}
