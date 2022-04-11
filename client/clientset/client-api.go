package clientset

import (
	"log"

	clusterregistryv1alpha1 "github.com/banzaicloud/cluster-registry/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	sch := runtime.NewScheme()
	err := clusterregistryv1alpha1.AddToScheme(sch)
	if err != nil {
		log.Fatalf("Error adding clusterregistryv1alpha1 to scheme: %v", err)
	}
}

type ResourceSyncRuleClientset interface {
	ResourceSyncRule(namespace string) ResourceSyncRuleInterface
}

// Clientsets -> Creates a new composition clientset
type Clientsets struct {
	dynamicClient dynamic.Interface
	kubernetes.Clientset
}

func NewForConfig(c *rest.Config) (*Clientsets, error) {
	config := c

	// create kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error creating cluster clientset from config: %v", err)
	}

	// create dynamic clientset for clusterregistryv1alpha1
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("error building dynamic clientset: %v", err)
	}

	return &Clientsets{dynamicClient: dynamicClient, Clientset: *clientset}, nil
}

// ResourceSyncRuleV1 -> extends the clientset to include ResourceSyncRuleInterface
func (c *Clientsets) ResourceSyncRuleV1(namespace string) ResourceSyncRuleInterface {
	return &resourceSyncRuleClient{
		dynamicClient: c.dynamicClient,
		ns:            namespace,
	}
}
