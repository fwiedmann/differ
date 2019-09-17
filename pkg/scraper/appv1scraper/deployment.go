package appv1scraper

import (
	"github.com/fwiedmann/differ/pkg/opts"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Deployment type struct
type Deployment struct {
}

// GetWorkloadRessources scrapes all appsV1 targets
func (d Deployment) GetWorkloadRessources(c *kubernetes.Clientset, scrapedResources map[string][]string) error {
	deployments, err := c.AppsV1().Deployments("").List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, deployment := range deployments.Items {
		if _, ok := deployment.Annotations[opts.DifferAnnotation]; ok {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				scrapedResources[container.Image] = append(scrapedResources[container.Image], deployment.Name)
			}
		}
	}
	return nil
}
