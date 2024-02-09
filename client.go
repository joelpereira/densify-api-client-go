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

type DensifyClient struct {
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

type AuthResponse struct {
	ApiToken string
	Expires  int64
	Status   int
	Message  string
}
type AuthError struct {
	/* variables */
}

const (
	apiEndpoint   = "/api/v2"
	apiAuthorize  = "/authorize"
	apiContainers = "containers"
	apiCloud      = "cloud"
)

// New Densify API Client
func NewDensifyClient(instanceURL *string, username *string, password *string, timeout_seconds int) (*DensifyClient, error) {
	if instanceURL == nil || username == nil || password == nil {
		return nil, fmt.Errorf(`instanceURL, username, password cannot be empty`)
	}

	pre := ""
	if !strings.HasPrefix(strings.ToLower(*instanceURL), "http") {
		pre = `https://`
	}

	c := DensifyClient{
		// HTTPClient: &http.Client{Timeout: time.Duration(timeout_seconds) * time.Second},
		HTTPClient: &http.Client{Timeout: 200},
	}

	c.BaseURL = fmt.Sprintf("%s%s%s", pre, strings.ToLower(*instanceURL), apiEndpoint)
	c.ApiUserName = *username
	c.ApiPassword = *password

	_, err := c.GetNewAuthToken()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *DensifyClient) ConfigureQuery(query *DensifyAPIQuery) error {
	// validate the query has all the required values
	if query == nil {
		return fmt.Errorf("query cannot be empty/nil")
	}
	// let's lowercase all the values first
	query.setValuesToLowercase()

	// validate the query is valid
	err := query.validate()
	if err != nil {
		return err
	}

	c.Query = query
	// reset other fields
	c.AnalysisIds = []string{}

	return nil // no error
}

func (c *DensifyClient) GetNewAuthToken() (*AuthResponse, error) {
	urlAuth := fmt.Sprintf("%s%s", c.BaseURL, apiAuthorize)

	postBody, _ := json.Marshal(map[string]string{
		"userName": c.ApiUserName,
		"pwd":      c.ApiPassword,
	})
	request, error := http.NewRequest("POST", urlAuth, bytes.NewBuffer(postBody))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if error != nil {
		return nil, error
	}
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

func (c *DensifyClient) GetAccountOrCluster() (*[]DensifyAnalysis, error) {
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

	// create a unique list of analyses name/number; this is only to be used for error output to list unique account name/numbers (since the Densify API can have duplicate analyses)
	var uniqueListOfAccounts UniqueList
	uniqueListOfAccounts.Initialize()

	errMsgParameter := ""
	for i := 0; i < len(analyses); i++ {
		// if it's a kubernetes/container request, look at analysis name, and check if it contains the cluster string instead
		if isKubernetesRequest {
			uniqueListOfAccounts.Add(analyses[i].AnalysisName)
			errMsgParameter = "cluster"
			if strings.Contains(strings.ToLower(analyses[i].AnalysisName), clusterName) {
				retAnalyses = append(retAnalyses, analyses[i])
				found = true
			}
		} else { // else, look at cloud account
			// search by account number
			if accountNumber != "" {
				uniqueListOfAccounts.Add(analyses[i].AccountId)
				errMsgParameter = "account number"
				if strings.Contains(strings.ToLower(analyses[i].AccountId), accountNumber) {
					retAnalyses = append(retAnalyses, analyses[i])
					found = true
				}
			} else if accountName != "" {
				uniqueListOfAccounts.Add(analyses[i].AccountName)
				errMsgParameter = "account name"
				if strings.Contains(strings.ToLower(analyses[i].AccountName), accountName) {
					retAnalyses = append(retAnalyses, analyses[i])
					found = true
				}
			}
		}
	}
	// if nothing was found, throw an error message with the list of account numbers/names/clusters
	if !found {
		qn := c.Query.AccountNumber
		if c.Query.AccountName != "" {
			qn = c.Query.AccountName
		} else if isKubernetesRequest {
			qn = c.Query.K8sCluster
		}

		retErr = fmt.Sprintf("no %s found named '%s'. Existing %ss are:\n", errMsgParameter, qn, errMsgParameter)
		// output the list of unique densify analyses (account name/number); this is to avoid duplicate values for ease of reading an error message and not for CSV machine processing.
		retErr += uniqueListOfAccounts.CsvStrWithNewLine()
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
func (c *DensifyClient) GetDensifyRecommendation() (*DensifyRecommendation, error) {
	emptyObj := c.returnEmptyRecommendationWithFallback()
	// make sure a query has been defined
	if c.Query == nil {
		if c.Query.SkipErrors {
			return emptyObj, nil
		}
		return emptyObj, fmt.Errorf("you must specify a query first")
	}
	err := c.Query.validate()
	if err != nil {
		if c.Query.SkipErrors {
			return emptyObj, nil
		}
		return emptyObj, err
	}

	isKubernetesRequest := c.Query.isKubernetesRequest()
	recos, err := c.GetDensifyRecommendations()
	if err != nil {
		if c.Query.SkipErrors {
			return emptyObj, nil
		}
		return emptyObj, err
	}
	// go through the list of recommendations and look for the entity name provided
	count := len(*recos)
	var reco DensifyRecommendation
	for i := 0; i < count; i++ {
		if isKubernetesRequest { // kubernetes recommendation
			// check the namespace and pod name as well
			recoName := strings.ToLower((*recos)[i].Container)
			recoNamespace := strings.ToLower((*recos)[i].Namespace)
			recoPodName := strings.ToLower((*recos)[i].PodService)
			recoControllerType := strings.ToLower((*recos)[i].ControllerType)
			if recoNamespace == c.Query.K8sNamespace && recoControllerType == c.Query.K8sControllerType && recoPodName == c.Query.K8sPodName {
				if c.Query.K8sContainerName != "" {
					// if a container name was provided, only return that one container, rather than the whole pod (which could have multiple containers)
					if recoName == c.Query.K8sContainerName {
						reco = (*recos)[i]
						// also manually add the container recommendation(s) to the pod list of containers
						reco.AddContainerToPod(&(*recos)[i])
						return &reco, nil
					}
				} else {
					// no container_name was provided in the query, so let's add to the pod list of containers
					if reco.isEmpty() {
						reco = (*recos)[i]
					}
					// also manually add the container recommendation(s) to the internal list
					reco.AddContainerToPod(&(*recos)[i])
					// if there are multiple containers within the pod, let's clear out the container name
					if len(reco.Containers) > 1 {
						reco.Name = reco.PodService // change the (container) name > pod name
						reco.Container = ""         // clear the container name
					}
				}
			}
		} else { // cloud instance recommendation
			recoName := strings.ToLower((*recos)[i].Name)
			if recoName == c.Query.SystemName {
				reco = (*recos)[i]
				// check if the ApprovedType needs a fallback
				if reco.ApprovedType == "" {
					reco.ApprovedType = c.Query.FallbackInstance
				}
				return &reco, nil
			}
		}
	}
	// return the recommendation if it exists
	if isKubernetesRequest && reco.Namespace != "" {
		return &reco, nil
	}

	if c.Query.SkipErrors {
		return emptyObj, nil
	}
	// return a different error msg if it's a cloud vs k8s query
	if isKubernetesRequest {
		return emptyObj, fmt.Errorf(`could not find a Densify recommendation for pod (%s) in namespace (%s), controller (%s), container name (%s)`, c.Query.K8sPodName, c.Query.K8sNamespace, c.Query.K8sControllerType, c.Query.K8sContainerName)
	} else {
		return emptyObj, fmt.Errorf("could not find a Densify recommendation named: %s", c.Query.SystemName)
	}
}

// pull a list of recommendations from the Densify API
func (c *DensifyClient) GetDensifyRecommendations() (*[]DensifyRecommendation, error) {
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
				recos[i].AnalysisType = apiContainers
			} else {
				recos[i].AnalysisType = apiCloud
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

// Pull a list of recommendations from the Densify API; spendTolerance "1.2" means anything more than 120% of optimal would move from "OK" to "Outside Spend Tolerance." Zero (0) means don't set spend tolerance.
func (c *DensifyClient) LoadDensifyInstanceGovernanceAllInstances(reco *DensifyRecommendation, spendTolerance float32) error {
	// make sure a query has been defined
	if c.Query == nil {
		return fmt.Errorf("you must specify a query first")
	}
	// check if we have a recommendation and EntityId
	if reco == nil || reco.EntityId == "" {
		return fmt.Errorf(`no Densify recommendation with an EntityId found; make sure you call GetRecommendation() first`)
	}

	// url: baseurl + /systems/entityid/analysis-details?target=all_instances
	url := fmt.Sprintf("%s/systems/%s/analysis-details?target=all_instances", c.BaseURL, reco.EntityId)
	// add spend tolerance
	if spendTolerance > 0 {
		url = fmt.Sprintf("%s&spendTolerance=%f", url, spendTolerance)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle error
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiToken))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "application/json")

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	//Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var instGov DensifyInstanceGovernance
	err = json.Unmarshal(body, &instGov)
	// Check for errors
	if err != nil {
		return errors.New("JSON decode error: " + err.Error())
	}
	// check if we received something else from the api
	if instGov.Message != "" {
		return fmt.Errorf("encountered an error: %s", instGov.Message)
	}

	// add it to the current recommendation
	reco.InstanceGovernance = instGov
	return nil
}

func (c *DensifyClient) IsTokenExpired() bool {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	return now >= c.ApiTokenExpiry
}

func (c *DensifyClient) ConvertRecommendationsToTF(recommendations *[]DensifyRecommendation) string {
	return c.ConvertRecommendationsToTFWithVarName(recommendations, "densify_recommendations")
}

func (c *DensifyClient) ConvertRecommendationsToTFWithVarName(recommendations *[]DensifyRecommendation, tfVarName string) string {
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

func (c *DensifyClient) returnEmptyRecommendationWithFallback() *DensifyRecommendation {
	const emptyRecoType = "Client Error - using fallback values"
	containers := []DensifyContainerRecommendation{{
		Container:          c.Query.K8sContainerName,
		FallbackCpuRequest: c.Query.FallbackCPURequest,
		FallbackCpuLimit:   c.Query.FallbackCPULimit,
		FallbackMemRequest: c.Query.FallbackMemRequest,
		FallbackMemLimit:   c.Query.FallbackMemLimit,
		RecommendationType: emptyRecoType,
	}}
	return &DensifyRecommendation{
		RecommendedType: c.Query.FallbackInstance,
		Containers:      containers,
	}
}
