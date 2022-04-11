package clientset

import (
	"context"
	"encoding/json"
	"fmt"

	clusterregistryv1alpha1 "github.com/banzaicloud/cluster-registry/api/v1alpha1"
	"github.com/banzaicloud/operator-tools/pkg/resources"
	"github.com/banzaicloud/operator-tools/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

const (
	resource    = "resourcesyncrules"
	syncDisable = "cluster-registry.k8s.cisco.com/resource-sync-disabled"
)

type ResourceSyncRuleInterface interface {
	List(ctx context.Context, ops metav1.ListOptions) (*clusterregistryv1alpha1.ResourceSyncRuleList, error)
	Get(ctx context.Context, name string, ops metav1.GetOptions) (*clusterregistryv1alpha1.ResourceSyncRule, error)
	Create(ctx context.Context, rule *unstructured.Unstructured, ops metav1.CreateOptions) (*clusterregistryv1alpha1.ResourceSyncRule, error)
	Patch(ctx context.Context, name string, pt k8stypes.PatchType, data []byte, ops metav1.PatchOptions) (*clusterregistryv1alpha1.ResourceSyncRule, error)

	TemplateToSyncRule(kind, ns, app, name, disable string) *unstructured.Unstructured
}

type resourceSyncRuleClient struct {
	dynamicClient dynamic.Interface
	ns            string
}

func (r *resourceSyncRuleClient) TemplateToSyncRule(kind, ns, app, name, disable string) *unstructured.Unstructured {
	unstructuredRule := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "clusterregistry.k8s.cisco.com/v1alpha1",
			"kind":       "ResourceSyncRule",
			"metadata": metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
				Annotations: map[string]string{
					syncDisable: disable,
				},
			},
			"spec": clusterregistryv1alpha1.ResourceSyncRuleSpec{
				GVK: resources.GroupVersionKind{Version: "v1", Kind: kind},
				Rules: []clusterregistryv1alpha1.SyncRule{
					{
						Matches: []clusterregistryv1alpha1.SyncRuleMatch{
							{ObjectKey: types.ObjectKey{Name: app, Namespace: ns}},
						},
					},
				},
			},
		},
	}

	return unstructuredRule
}

func (r *resourceSyncRuleClient) unstructuredToRule(unstructuredRule *unstructured.Unstructured) (*clusterregistryv1alpha1.ResourceSyncRule, error) {
	data, err := unstructuredRule.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var rule = &clusterregistryv1alpha1.ResourceSyncRule{}
	err = json.Unmarshal(data, rule)

	return rule, err
}

func (r *resourceSyncRuleClient) List(ctx context.Context, ops metav1.ListOptions) (*clusterregistryv1alpha1.ResourceSyncRuleList, error) {
	unstructuredRules, err := r.dynamicClient.Resource(clusterregistryv1alpha1.GroupVersion.WithResource(resource)).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %v", resource, err)
	}

	data, err := unstructuredRules.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var rules = &clusterregistryv1alpha1.ResourceSyncRuleList{}
	err = json.Unmarshal(data, rules)

	return rules, err
}

func (r *resourceSyncRuleClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*clusterregistryv1alpha1.ResourceSyncRule, error) {
	unstructuredRule, err := r.dynamicClient.Resource(clusterregistryv1alpha1.GroupVersion.WithResource(resource)).Namespace(r.ns).Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", resource, err)
	}

	rule, err := r.unstructuredToRule(unstructuredRule)
	return rule, err
}

func (r *resourceSyncRuleClient) Create(ctx context.Context, rule *unstructured.Unstructured, ops metav1.CreateOptions) (*clusterregistryv1alpha1.ResourceSyncRule, error) {
	unstructuredRule, err := r.dynamicClient.Resource(clusterregistryv1alpha1.GroupVersion.WithResource(resource)).Namespace(r.ns).Create(context.TODO(), rule, ops)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %v", resource, err)
	}

	return r.unstructuredToRule(unstructuredRule)
}

func (r *resourceSyncRuleClient) Patch(ctx context.Context, name string, pt k8stypes.PatchType, data []byte, ops metav1.PatchOptions) (result *clusterregistryv1alpha1.ResourceSyncRule, err error) {
	unstructuredRule, err := r.dynamicClient.Resource(clusterregistryv1alpha1.GroupVersion.WithResource(resource)).Namespace(r.ns).Patch(context.TODO(), name, pt, data, ops)
	if err != nil {
		return nil, fmt.Errorf("failed to patch %s: %v", resource, err)
	}

	return r.unstructuredToRule(unstructuredRule)
}
