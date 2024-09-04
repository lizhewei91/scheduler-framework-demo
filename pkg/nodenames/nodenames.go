package nodenames

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type NodeNames struct {
	handle framework.Handle
}

var _ framework.FilterPlugin = &NodeNames{}
var _ framework.ScorePlugin = &NodeNames{}

const (
	// Name is the name of the plugin used in Registry and configurations.
	Name = "NodeNames"

	AnnotationKey = "nodeNames"
)

// Name returns name of the plugin. It is used in logs, etc.
func (ns *NodeNames) Name() string {
	return Name
}

func (ns *NodeNames) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	var wantNodes []string
	// 获取 Pod 的 annotations
	if nodeNames, exists := pod.Annotations[AnnotationKey]; exists {
		wantNodes = strings.Split(nodeNames, ",")
		if !Contains(wantNodes, nodeInfo.Node().Name) {
			klog.V(0).Infof("node: %v does not match expected: %v", nodeInfo.Node().Name, wantNodes)
			// 如果不匹配，则拒绝该节点
			return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("Node: %s does not match expected: %s", nodeInfo.Node().Name, wantNodes))
		}
	}
	klog.V(0).Infof("node: %v is match expected: %v", nodeInfo.Node().Name, wantNodes)
	// 如果没有指定 nodename 注解，则允许该节点通过
	return framework.NewStatus(framework.Success)
}

func (ns *NodeNames) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	score := rand.Int63n(10 + 1)
	klog.V(0).Infof("node: %v get score: %v\n", nodeName, score)
	return score, framework.NewStatus(framework.Success)
}

// ScoreExtensions of the Score plugin.
func (ns *NodeNames) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &NodeNames{handle: h}, nil
}

func Contains(slice []string, element string) bool {
	for _, e := range slice {
		if e == element {
			return true
		}
	}
	return false
}
