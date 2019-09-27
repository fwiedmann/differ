package appv1scraper

import (
	"github.com/fwiedmann/differ/pkg/opts"
	"github.com/fwiedmann/differ/pkg/scraper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Deployment type struct
type Deployment struct {
}

// GetWorkloadResources scrapes all appsV1 targets
func (d Deployment) GetWorkloadResources(c *kubernetes.Clientset, namespace string, scraperStore scraper.ResourceStore) error {
	deployments, err := c.AppsV1().Deployments(namespace).List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, deployment := range deployments.Items {
		if _, ok := deployment.Annotations[opts.DifferAnnotation]; ok {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				scraper.AddResource(container.Image, "apps/v1", "deployment", deployment.Namespace, deployment.Name, scraperStore)
			}
		}
	}
	return nil
}
