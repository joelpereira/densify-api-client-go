package densify

import (
	"fmt"
	"sort"
	"strings"
)

type DensifyGuardrailsList struct {
	Compatibility string                                   `json:"compatibility"`
	InstanceList  map[int]map[string]DensifyGuardrailsNode `json:"nodeList"`
}
type DensifyGuardrailsNode struct {
	InstanceType       string  `json:"instanceType"`
	BlendedScore       int     `json:"blendedScore"`
	PercentOptimalCost float64 `json:"percentOptimalCost"`
}

func (l *DensifyGuardrailsList) Length() int {
	return len(l.InstanceList)
}
func (l *DensifyGuardrailsList) LengthInKey(score int) int {
	return len(l.InstanceList[score])
}
func (l *DensifyGuardrailsList) TotalLength() int {
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

func (l *DensifyGuardrailsList) AddNode(instance string, score int, percentOptimalCost float64) {
	// check that the main list was instantiated
	if l.InstanceList == nil {
		l.InstanceList = map[int]map[string]DensifyGuardrailsNode{}
	}
	// check that the sub list was instantiated
	if l.InstanceList[score] == nil {
		// create sub list
		l.InstanceList[score] = map[string]DensifyGuardrailsNode{}
	}

	// add item to sub list
	l.InstanceList[score][instance] = DensifyGuardrailsNode{
		InstanceType:       instance,
		BlendedScore:       score,
		PercentOptimalCost: percentOptimalCost,
	}
}

func (l *DensifyGuardrailsList) GetSortedScoreList() []int {
	keys := make([]int, 0, len(l.InstanceList))
	for k, _ := range l.InstanceList {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
func (l *DensifyGuardrailsList) GetMinScore() int {
	if l.InstanceList == nil || l.Length() == 0 {
		return 0
	}
	keys := l.GetSortedScoreList() // return first one in the sorted list
	return keys[0]
}
func (l *DensifyGuardrailsList) GetMaxScore() int {
	if l.InstanceList == nil || l.Length() == 0 {
		return 0
	}
	keys := l.GetSortedScoreList()
	return keys[len(keys)-1] // return last one in the sorted list
}
func (l *DensifyGuardrailsList) GetScoreItems(score int) map[string]DensifyGuardrailsNode {
	return l.InstanceList[score]
}

func (r *DensifyRecommendation) GetGuardrailsOK() (*DensifyGuardrailsList, error) {
	return r.GetGuardrailsCompatLevel("OK")
}
func (r *DensifyRecommendation) GetGuardrailsIncompatible() (*DensifyGuardrailsList, error) {
	return r.GetGuardrailsCompatLevel("Technically Incompatible")
}
func (r *DensifyRecommendation) GetGuardrailsInsufficientResources() (*DensifyGuardrailsList, error) {
	return r.GetGuardrailsCompatLevel("Insufficient Resources")
}
func (r *DensifyRecommendation) GetGuardrailsSpendTolerance() (*DensifyGuardrailsList, error) {
	return r.GetGuardrailsCompatLevel("Outside Spend Tolerance")
}

func (r *DensifyRecommendation) GetGuardrailsCompatLevel(compatibilityLevel string) (*DensifyGuardrailsList, error) {
	targets := r.Guardrails.getCompatibilityList(compatibilityLevel)
	if targets == nil {
		return nil, fmt.Errorf("no instance governance list available for instance: %s", r.Name)
	}
	return targets, nil
}
func (g *DensifyGuardrails) getCompatibilityList(compat string) *DensifyGuardrailsList {
	l := DensifyGuardrailsList{
		Compatibility: compat,
		InstanceList:  map[int]map[string]DensifyGuardrailsNode{},
	}

	compatLowerCase := strings.ToLower(compat)
	for i := 0; i < len(g.Targets); i++ {
		item := g.Targets[i]
		if strings.ToLower(item.Compatibility) == compatLowerCase {
			l.AddNode(item.InstanceType, item.BlendedScore, float64(item.PercentOptimalCost))
		}
	}

	return &l
}
