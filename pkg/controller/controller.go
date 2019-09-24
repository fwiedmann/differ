package controller

import (
	"github.com/fwiedmann/differ/pkg/controller/util"
	"github.com/fwiedmann/differ/pkg/opts"
	"github.com/fwiedmann/differ/pkg/scraper"
	"github.com/fwiedmann/differ/pkg/scraper/appv1scraper"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	scrapers []resourceScraper
)

type resourceScraper interface {
	GetWorkloadResources(c *kubernetes.Clientset, namespace string, scrapedResources map[string][]scraper.ScrapedResource) error
}

// Controller type struct
type Controller struct {
	config *opts.ControllerConfig
}

// New initialize the differ controller
func New(c *opts.ControllerConfig) *Controller {
	return &Controller{
		config: c,
	}
}

// Run starts differ controller loop
func (c *Controller) Run() error {
	for {

		scrapeResult := make(map[string][]scraper.ScrapedResource)

		kubernetesClient, err := util.InitKubernetesClient()
		if err != nil {
			return err
		}

		for _, s := range scrapers {
			if err := s.GetWorkloadResources(kubernetesClient, c.config.Namespace, scrapeResult); err != nil {
				return err
			}
		}
		log.Debugf("%+v", scrapeResult)
		imagesSortedByRegistry := make(map[string][]string)
		for image := range scrapeResult {
			util.AddToSortedImages(image, imagesSortedByRegistry)
		}
		log.Debugf("%+v", imagesSortedByRegistry)
		c.config.ControllerSleep()
	}
}
func init() {
	scrapers = append(scrapers, appv1scraper.Deployment{})
}
