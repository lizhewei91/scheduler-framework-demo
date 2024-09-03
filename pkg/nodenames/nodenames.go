package nodenames

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
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
		for _, nodeName := range wantNodes {
			if nodeInfo.Node().Name == nodeName {
				klog.V(0).Infof("node: %v is match expected: %v", nodeInfo.Node().Name, wantNodes)
				// 如果节点名称与指定的名称匹配，则通过筛选
				return framework.NewStatus(framework.Success)
			} else {
				// 如果不匹配，则拒绝该节点
				return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("Node name %s does not match expected %s", nodeInfo.Node().Name, wantNodes))
			}
		}

	}
	// 如果没有指定 nodename 注解，则允许该节点通过
	return framework.NewStatus(framework.Success)
}

func (ns *NodeNames) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	score := rand.Int63n(extenderv1.MaxExtenderPriority + 1)
	klog.V(0).Infof("pod: %v/%v, node: %v get score: %v\n", pod.Namespace, pod.Name, nodeName, score)
	return score, framework.NewStatus(framework.Success)
}

func (ns *NodeNames) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var minScore, maxScore int64 = 10, 1

	// Find the min and max scores
	for _, score := range scores {
		if score.Score > maxScore {
			maxScore = score.Score
		}
		if score.Score < minScore {
			minScore = score.Score
		}
	}

	// Normalize the scores to a scale of 0-100
	for i := range scores {
		scores[i].Score = ((scores[i].Score - minScore) * 100) / (maxScore - minScore)
	}
	return framework.NewStatus(framework.Success)
}

// ScoreExtensions of the Score plugin.
func (ns *NodeNames) ScoreExtensions() framework.ScoreExtensions {
	return ns
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &NodeNames{handle: h}, nil
}
