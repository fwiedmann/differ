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

// GetWorkloadRessources scrapes all appsV1 targets
func (d Deployment) GetWorkloadRessources(c *kubernetes.Clientset, namespace string, scrapedResources map[string][]scraper.ScrapedResource) error {
	deployments, err := c.AppsV1().Deployments(namespace).List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, deployment := range deployments.Items {
		if _, ok := deployment.Annotations[opts.DifferAnnotation]; ok {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				scraper.AddRessource(container.Image, "apps/v1", "deployment", deployment.Namespace, deployment.Name, scrapedResources)
			}
		}
	}
	return nil
}
