package densify

import "fmt"

type FloatType float64
type Currency float32

func ConvertFloatToStr(value interface{}) string {
	return fmt.Sprintf("%f", value)
}

type DensifyAnalysis struct {
	AccountId        string `json:"accountId"`
	AccountName      string `json:"accountName"`
	AnalysisId       string `json:"analysisId"`
	AnalysisName     string `json:"analysisName"`
	Href             string `json:"href"`
	AnalysisStatus   string `json:"analysisStatus"`
	AnalysisResults  string `json:"analysisResults"`
	PolicyName       string `json:"policyName"`
	PolicyInstanceId string `json:"policyInstanceId"`
}

type DensifyRecommendation struct {
	// Will be manually added from the analysis

	AnalysisType       string // Values: cloud, containers
	AnalysisTechnology string // Values: aws, gcp, azure, k8s
	AccountId          string `json:"accountId"`
	AnalysisId         string `json:"analysisId"`
	AccountName        string `json:"accountName"`
	// AnalysisName string `json:"analysisName"`
	ApprovedType string `json:"approvedType"` // this checks if a change has been approved (by looking at the ApprovalType) and returns the RecommendedType, otherwise it will return the CurrentType.

	// returned by Densify API
	// Cloud

	EntityId                string    `json:"entityId"`
	ResourceId              string    `json:"resourceId"`
	AccountIdRef            string    `json:"accountIdRef"`
	Region                  string    `json:"region"`
	CurrentType             string    `json:"currentType"`
	RecommendationType      string    `json:"recommendationType"`
	RecommendedType         string    `json:"recommendedType"`
	ImplementationMethod    string    `json:"implementationMethod"`
	PredictedUptime         FloatType `json:"predictedUptime"`
	TotalHoursRunning       int64     `json:"totalHoursRunning"`
	TotalHours              int64     `json:"totalHours"`
	Name                    string    `json:"name"`
	RptHref                 string    `json:"rptHref"`
	ApprovalType            string    `json:"approvalType"`
	DensifyPolicy           string    `json:"densifyPolicy"`
	SavingsEstimate         Currency  `json:"savingsEstimate"`
	EffortEstimate          string    `json:"effortEstimate"`
	PowerState              string    `json:"powerState"`
	RecommendedHostEntityId string    `json:"recommendedHostEntityId"`
	CurrentCost             Currency  `json:"currentCost"`
	RecommendedCost         Currency  `json:"recommendedCost"`
	ServiceType             string    `json:"serviceType"`
	CurrentHourlyRate       FloatType `json:"currentHourlyRate"`
	RecommendedHourlyRate   FloatType `json:"recommendedHourlyRate"`
	RecommFirstSeen         int64     `json:"recommFirstSeen"`
	RecommLastSeen          int64     `json:"recommLastSeen"`
	RecommSeenCount         int64     `json:"recommSeenCount"`
	AuditInfo               AuditInfo `json:"auditInfo"`

	// ASG specific values
	MinGroupCurrent             string    `json:"minGroupCurrent"`
	MinGroupRecommended         string    `json:"minGroupRecommended"`
	MaxGroupCurrent             string    `json:"maxGroupCurrent"`
	MaxGroupRecommended         string    `json:"maxGroupRecommended"`
	CurrentDesiredCapacity      string    `json:"currentDesiredCapacity"`
	AvgInstanceCountRecommended FloatType `json:"avgInstanceCountRecommended"`
	AvgInstanceCountCurrent     FloatType `json:"avgInstanceCountCurrent"`

	// Container values
	Container             string    `json:"container"`
	Cluster               string    `json:"cluster"`
	HostName              string    `json:"hostName"`
	EstimatedSavings      FloatType `json:"estimatedSavings"`
	TotalNetSavings       FloatType `json:"totalNetSavings"`
	DisplayName           string    `json:"displayName"`
	PodService            string    `json:"podService"`
	CurrentCount          int64     `json:"currentCount"`
	CurrentCpuRequest     int64     `json:"currentCpuRequest"`
	CurrentCpuLimit       int64     `json:"currentCpuLimit"`
	CurrentMemRequest     int64     `json:"currentMemRequest"`
	CurrentMemLimit       int64     `json:"currentMemLimit"`
	RecommendedCpuRequest int64     `json:"recommendedCpuRequest"`
	RecommendedCpuLimit   int64     `json:"recommendedCpuLimit"`
	RecommendedMemRequest int64     `json:"recommendedMemRequest"`
	RecommendedMemLimit   int64     `json:"recommendedMemLimit"`
	RunningHours          int64     `json:"runningHours"`
	ControllerType        string    `json:"controllerType"`
	Namespace             string    `json:"namespace"`

	Containers         []DensifyContainerRecommendation `json:"containers"`
	InstanceGovernance DensifyInstanceGovernance        `json:"instanceGovernance"`
}

type AuditInfo struct {
	DataCollection AuditInfoDataCollection     `json:"dataCollection"`
	WorkloadData   AuditInfoWorkloadDataLast30 `json:"workloadDataLast30"`
}

type AuditInfoDataCollection struct {
	DateFirstAudited int64 `json:"dateFirstAudited"`
	DateLastAudited  int64 `json:"dateLastAudited"`
	AuditCount       int64 `json:"auditCount"`
}

type AuditInfoWorkloadDataLast30 struct {
	FirstDate int64 `json:"firstDate"`
	LastDate  int64 `json:"lastDate"`
	TotalDays int64 `json:"totalDays"`
	SeenDays  int64 `json:"seenDays"`
}

type DensifyContainerRecommendation struct {
	Container             string    `json:"container"`
	Cluster               string    `json:"cluster"`
	EntityId              string    `json:"entityId"`
	EstimatedSavings      FloatType `json:"estimatedSavings"`
	TotalNetSavings       FloatType `json:"totalNetSavings"`
	DisplayName           string    `json:"displayName"`
	PodService            string    `json:"podService"`
	CurrentCount          int64     `json:"currentCount"`
	CurrentCpuRequest     int64     `json:"currentCpuRequest"`
	CurrentCpuLimit       int64     `json:"currentCpuLimit"`
	CurrentMemRequest     int64     `json:"currentMemRequest"`
	CurrentMemLimit       int64     `json:"currentMemLimit"`
	RecommendedCpuRequest int64     `json:"recommendedCpuRequest"`
	RecommendedCpuLimit   int64     `json:"recommendedCpuLimit"`
	RecommendedMemRequest int64     `json:"recommendedMemRequest"`
	RecommendedMemLimit   int64     `json:"recommendedMemLimit"`
	FallbackCpuRequest    string    `json:"fallbackCpuRequest"`
	FallbackCpuLimit      string    `json:"fallbackCpuLimit"`
	FallbackMemRequest    string    `json:"fallbackMemRequest"`
	FallbackMemLimit      string    `json:"fallbackMemLimit"`
	RunningHours          int64     `json:"runningHours"`
	ControllerType        string    `json:"controllerType"`
	Namespace             string    `json:"namespace"`
	RecommendationType    string    `json:"recommendationType"`
	ApprovalType          string    `json:"approvalType"`
	ApprovedType          string    `json:"approvedType"`
	DaysRecoUnchanged     int64     `json:"recommSeenCount"`
}

type DensifyInstanceGovernance struct {
	CurrentInstance DensifyInstanceGovernanceCurrent  `json:"current"`
	OptimalInstance DensifyInstanceGovernanceOptimal  `json:"optimal"`
	Targets         []DensifyInstanceGovernanceTarget `json:"targets"`

	Status  int    `json:"status"`  // if there's an error, this will be populated
	Message string `json:"message"` // if there's an error, this will be populated
}

type DensifyInstanceGovernanceCurrent struct {
	EntityId      string `json:"entityId"`
	DisplayName   string `json:"displayName"`
	ResourceId    string `json:"resourceId"`
	ResourceGroup string `json:"resourceGroup"`
	InstanceType  string `json:"instanceType"`
	BlendedScore  int    `json:"blendedScore"`
	Compatability string `json:"compatability"`
}

type DensifyInstanceGovernanceOptimal struct {
	InstanceType       string `json:"instanceType"`
	BlendedScore       int    `json:"blendedScore"`
	Compatability      string `json:"compatability"`
	RecommendationType string `json:"recommendationType"`
}

type DensifyInstanceGovernanceTarget struct {
	InstanceType          string   `json:"instance_type"`
	BlendedScore          int      `json:"blended_score"`
	Compatability         string   `json:"compatability"` // Values are: OK or Incompatible
	IncompatibilityReason []string `json:"incompatibilityReason"`
}

// this checks if a change has been approved (by looking at the ApprovalType) and returns the RecommendedType, otherwise it will return the CurrentType.
func (r *DensifyRecommendation) GetApprovedType() string {
	// basic check(s) first
	if r == nil {
		return ""
	}

	switch r.ApprovalType {
	case "na":
		// not approved; use CurrentType
		return r.CurrentType
	case "all":
		// all/any recommendation is approved
		return r.RecommendedType
	case "any":
		// all/any recommendation is approved
		return r.RecommendedType
	default:
		// specific recommendation is approved and specified in ApprovalType
		return r.ApprovalType
	}
}

// adds the container recommendation to the list of containers
func (pod *DensifyRecommendation) AddContainerToPod(reco *DensifyRecommendation) {
	c := DensifyContainerRecommendation{
		Container:             reco.Container,
		DisplayName:           reco.DisplayName,
		Cluster:               reco.Cluster,
		Namespace:             reco.Namespace,
		PodService:            reco.PodService,
		ControllerType:        reco.ControllerType,
		CurrentCpuRequest:     reco.CurrentCpuRequest,
		CurrentCpuLimit:       reco.CurrentCpuLimit,
		CurrentMemRequest:     reco.CurrentMemRequest,
		CurrentMemLimit:       reco.CurrentMemLimit,
		RecommendedCpuRequest: reco.RecommendedCpuRequest,
		RecommendedCpuLimit:   reco.RecommendedCpuLimit,
		RecommendedMemRequest: reco.RecommendedMemRequest,
		RecommendedMemLimit:   reco.RecommendedMemLimit,
		ApprovalType:          reco.ApprovalType,
		ApprovedType:          reco.ApprovedType,
		EntityId:              reco.EntityId,
		RecommendationType:    reco.RecommendationType,
		DaysRecoUnchanged:     reco.RecommSeenCount,
	}
	pod.Containers = append(pod.Containers, c)
}

// returns true if the object is empty/nil
func (r DensifyRecommendation) isEmpty() bool {
	if r.Name == "" && r.Container == "" && r.Namespace == "" {
		return true
	}
	return false
}
