package controller

import (
	"github.com/fwiedmann/differ/pkg/controller/util"
	"github.com/fwiedmann/differ/pkg/opts"
	"github.com/fwiedmann/differ/pkg/scraper/appv1scraper"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	scrapers []scraper
)

type scraper interface {
	GetWorkloadRessources(c *kubernetes.Clientset, namespace string, scrapedResources map[string]map[string]map[string]string) error
}

type Controller struct {
	confing *opts.ControllerConfig
}

func New(c *opts.ControllerConfig) *Controller {
	return &Controller{
		confing: c,
	}
}

func (c *Controller) Run() error {
	for {

		scrapeResult := make(map[string]map[string]map[string]string)

		kubernetesClient, err := util.InitKubernetesClient()
		if err != nil {
			return err
		}

		for _, scraper := range scrapers {
			if err := scraper.GetWorkloadRessources(kubernetesClient, c.confing.Namespace, scrapeResult); err != nil {
				return err
			}
		}
		log.Debug(scrapeResult)

		c.confing.ControllerSleep()
	}
}
func init() {
	scrapers = append(scrapers, appv1scraper.Deployment{})
}
