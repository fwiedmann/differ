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
	GetWorkloadResources(c *kubernetes.Clientset, namespace string, scrapedResources scraper.ResourceStore) error
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
		scraperStore := make(scraper.ResourceStore)

		kubernetesClient, err := util.InitKubernetesClient()
		if err != nil {
			return err
		}

		for _, s := range scrapers {
			if err := s.GetWorkloadResources(kubernetesClient, c.config.Namespace, scraperStore); err != nil {
				return err
			}
		}
		log.Debugf("Scraped resources:\n%+v", scraperStore)

		c.config.ControllerSleep()
	}
}
func init() {
	scrapers = append(scrapers, appv1scraper.Deployment{})
}
