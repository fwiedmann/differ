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
	GetWorkloadRessources(c *kubernetes.Clientset, namespace string, scrapedResources map[string][]scraper.ScrapedResource) error
}

// Controller type struct
type Controller struct {
	confing *opts.ControllerConfig
}

// New initalaize the differ controller
func New(c *opts.ControllerConfig) *Controller {
	return &Controller{
		confing: c,
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

		for _, scraper := range scrapers {
			if err := scraper.GetWorkloadRessources(kubernetesClient, c.confing.Namespace, scrapeResult); err != nil {
				return err
			}
		}
		log.Debugf("%+v", scrapeResult)

		c.confing.ControllerSleep()
	}
}
func init() {
	scrapers = append(scrapers, appv1scraper.Deployment{})
}
