package densify

import "fmt"

type FloatType float64
type Currency float32

func ConvertFloatToStr(value interface{}) string {
	return fmt.Sprintf("%f", value)
}

type DensifyAnalysis struct {
	AccountId       string `json:"accountId"`
	AccountName     string `json:"accountName"`
	AnalysisId      string `json:"analysisId"`
	AnalysisName    string `json:"analysisName"`
	Href            string `json:"href"`
	AnalysisStatus  string `json:"analysisStatus"`
	AnalysisResults string `json:"analysisResults"`
}

type DensifyRecommendations struct {
	// Will be manually added from the analysis

	AnalysisType       string // Values: cloud, containers
	AnalysisTechnology string // Values: aws, gcp, azure, k8s
	AccountId          string `json:"accountId"`
	AnalysisId         string `json:"analysisId"`
	AccountName        string `json:"accountName"`
	// AnalysisName string `json:"analysisName"`

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
	TotalHoursRunning       int       `json:"totalHoursRunning"`
	TotalHours              int       `json:"totalHours"`
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
	RecommSeenCount         int       `json:"recommSeenCount"`
	AuditInfo               AuditInfo `json:"auditInfo"`

	// ASG specific values
	MinGroupCurrent             string    `json:"minGroupCurrent"`
	MinGroupRecommended         string    `json:"minGroupRecommended"`
	MaxGroupCurrent             string    `json:"maxGroupCurrent"`
	MaxGroupRecommended         string    `json:"maxGroupRecommended"`
	CurrentDesiredCapacity      string    `json:"currentDesiredCapacity"`
	AvgInstanceCountRecommended FloatType `json:"avgInstanceCountRecommended"`
	AvgInstanceCountCurrent     FloatType `json:"avgInstanceCountCurrent"`

	// Containers
	Container             string    `json:"container"`
	Cluster               string    `json:"cluster"`
	HostName              string    `json:"hostName"`
	EstimatedSavings      FloatType `json:"estimatedSavings"`
	TotalNetSavings       FloatType `json:"totalNetSavings"`
	DisplayName           string    `json:"displayName"`
	PodService            string    `json:"podService"`
	CurrentCount          int       `json:"currentCount"`
	CurrentCpuRequest     int       `json:"currentCpuRequest"`
	CurrentCpuLimit       int       `json:"currentCpuLimit"`
	CurrentMemRequest     int       `json:"currentMemRequest"`
	CurrentMemLimit       int       `json:"currentMemLimit"`
	RecommendedCpuRequest int       `json:"recommendedCpuRequest"`
	RecommendedCpuLimit   int       `json:"recommendedCpuLimit"`
	RecommendedMemRequest int       `json:"recommendedMemRequest"`
	RecommendedMemLimit   int       `json:"recommendedMemLimit"`
	RunningHours          int       `json:"runningHours"`
	ControllerType        string    `json:"controllerType"`
	Namespace             string    `json:"namespace"`
}

type AuditInfo struct {
	DataCollection AuditInfoDataCollection     `json:"dataCollection"`
	WorkloadData   AuditInfoWorkloadDataLast30 `json:"workloadDataLast30"`
}

type AuditInfoDataCollection struct {
	DateFirstAudited int64 `json:"dateFirstAudited"`
	DateLastAudited  int64 `json:"dateLastAudited"`
	AuditCount       int   `json:"auditCount"`
}

type AuditInfoWorkloadDataLast30 struct {
	FirstDate int64 `json:"firstDate"`
	LastDate  int64 `json:"lastDate"`
	TotalDays int   `json:"totalDays"`
	SeenDays  int   `json:"seenDays"`
}
