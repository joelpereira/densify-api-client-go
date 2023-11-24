package densify

import (
	"fmt"
	"sort"
	"strings"
)

type DensifyGovernanceInstanceList struct {
	Compatability string                                           `json:"compatability"`
	InstanceList  map[int]map[string]DensifyGovernanceInstanceNode `json:"nodeList"`
}
type DensifyGovernanceInstanceNode struct {
	InstanceType string `json:"instance_type"`
	BlendedScore int    `json:"blended_score"`
}

func (l *DensifyGovernanceInstanceList) Length() int {
	return len(l.InstanceList)
}
func (l *DensifyGovernanceInstanceList) LengthInKey(score int) int {
	return len(l.InstanceList[score])
}
func (l *DensifyGovernanceInstanceList) TotalLength() int {
	if l.InstanceList == nil || l.Length() == 0 {
		return 0
	}
	keys := l.GetSortedScoreList() // return first one in the sorted list
	length := 0
	for i := 0; i < len(keys); i++ {
		length += len(l.InstanceList[keys[i]])
	}
	return length
}

func (l *DensifyGovernanceInstanceList) AddNode(instance string, score int) {
	// check that the main list was instantiated
	if l.InstanceList == nil {
		l.InstanceList = map[int]map[string]DensifyGovernanceInstanceNode{}
	}
	// check that the sub list was instantiated
	if l.InstanceList[score] == nil {
		// create sub list
		l.InstanceList[score] = map[string]DensifyGovernanceInstanceNode{}
	}

	// add item to sub list
	l.InstanceList[score][instance] = DensifyGovernanceInstanceNode{
		InstanceType: instance,
		BlendedScore: score,
	}
}

func (l *DensifyGovernanceInstanceList) GetSortedScoreList() []int {
	keys := make([]int, 0, len(l.InstanceList))
	for k, _ := range l.InstanceList {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
func (l *DensifyGovernanceInstanceList) GetMinScore() int {
	if l.InstanceList == nil || l.Length() == 0 {
		return 0
	}
	keys := l.GetSortedScoreList() // return first one in the sorted list
	return keys[0]
}
func (l *DensifyGovernanceInstanceList) GetMaxScore() int {
	if l.InstanceList == nil || l.Length() == 0 {
		return 0
	}
	keys := l.GetSortedScoreList()
	return keys[len(keys)-1] // return last one in the sorted list
}
func (l *DensifyGovernanceInstanceList) GetScoreItems(score int) map[string]DensifyGovernanceInstanceNode {
	return l.InstanceList[score]
}

func (r *DensifyRecommendation) GetInstanceGovernanceOK() (*DensifyGovernanceInstanceList, error) {
	return r.GetInstanceGovernance("OK")
}
func (r *DensifyRecommendation) GetInstanceGovernanceIncompatible() (*DensifyGovernanceInstanceList, error) {
	return r.GetInstanceGovernance("Technically Incompatible")
}
func (r *DensifyRecommendation) GetInstanceGovernanceInsufficientResources() (*DensifyGovernanceInstanceList, error) {
	return r.GetInstanceGovernance("Insufficient Resources")
}
func (r *DensifyRecommendation) GetInstanceGovernanceSpendTolerance() (*DensifyGovernanceInstanceList, error) {
	return r.GetInstanceGovernance("Outside Spend Tolerance")
}

func (r *DensifyRecommendation) GetInstanceGovernance(compatabilityLevel string) (*DensifyGovernanceInstanceList, error) {
	targets := r.InstanceGovernance.getCompatabilityList(compatabilityLevel)
	if targets == nil {
		return nil, fmt.Errorf("no instance governance list available for instance: %s", r.Name)
	}
	return targets, nil
}
func (g *DensifyInstanceGovernance) getCompatabilityList(compat string) *DensifyGovernanceInstanceList {
	l := DensifyGovernanceInstanceList{
		Compatability: compat,
		InstanceList:  map[int]map[string]DensifyGovernanceInstanceNode{},
	}

	compatLowerCase := strings.ToLower(compat)
	for i := 0; i < len(g.Targets); i++ {
		item := g.Targets[i]
		if strings.ToLower(item.Compatability) == compatLowerCase {
			l.AddNode(item.InstanceType, item.BlendedScore)
		}
	}

	return &l
}
