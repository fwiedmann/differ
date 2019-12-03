/*
 * MIT License
 *
 * Copyright (c) 2019 Felix Wiedmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package metrics

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/fwiedmann/differ/pkg/opts"

	"github.com/fwiedmann/differ/pkg/store"

	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	m            = sync.Mutex{}
	gaugeMetrics = map[string]*prometheus.Desc{
		"differ_config": prometheus.NewDesc(
			"differ_config",
			"Shows the configuration of differ",
			[]string{"version", "namespace", "sleep_duration", "metrics_port", "metrics_path"},
			nil,
		),
		"differ_scraped_image": prometheus.NewDesc(
			"differ_scraped_image",
			"Scraped image with meta information",
			[]string{"image_name", "image_tag", "resource_type", "resource_name", "resource_api_version", "namespace"},
			nil,
		),
		"differ_update_image": prometheus.NewDesc(
			"differ_update_image",
			"Image which is an newer version available",
			[]string{"image_name", "image_tag", "resource_type", "resource_name", "resource_api_version", "namespace", "newer_image_tag"},
			nil,
		),
	}
)

type metric struct {
	labels     []string
	value      float64
	metricType prometheus.ValueType
	isStatic   bool
}

// metricStore holds all current metrics.
// metricName - labels as ID - metric struct
var metricStore = make(map[string]map[string]metric)

// customCollector satisfy the prometheus collector interface
type customCollector struct {
}

// Describe implementation of prometheus Collector
func (c *customCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, description := range gaugeMetrics {
		ch <- description
	}
}

// Collect implementation of prometheus Collector
func (c *customCollector) Collect(ch chan<- prometheus.Metric) {
	for metricName, metrics := range metricStore {
		description := gaugeMetrics[metricName]
		for _, metricValue := range metrics {
			ch <- prometheus.MustNewConstMetric(description, metricValue.metricType, metricValue.value, metricValue.labels...)
		}
	}
}

// DeleteNotScrapedResources which are not scraped by the last scrape
func DeleteNotScrapedResources(cache *store.Instance) {
	m.Lock()
	for metricName, metrics := range metricStore {
		for metricID, metric := range metrics {
			if metric.isStatic {
				continue
			} else {
				var found bool
				for _, imageName := range cache.GetDeepCopy() {
					for _, scrapedImage := range imageName {
						tmpMetricID := fmt.Sprintf("%s %s", scrapedImage.ImageName, scrapedImage.ImageTag)
						if metricID == tmpMetricID {
							found = true
						}
					}
				}
				if !found {
					delete(metricStore[metricName], metricID)
				}
			}
		}
	}
	m.Unlock()
}

func setGaugeValue(static bool, metricName string, value float64, labels ...string) {
	m.Lock()
	if _, found := gaugeMetrics[metricName]; !found {
		log.Warnf("Could not find %s metric in metrics pkg", metricName)
	} else {
		if _, found := metricStore[metricName]; !found {
			metricStore[metricName] = make(map[string]metric)
		}

		var id string

		for _, l := range labels {
			id += l
		}
		metricStore[metricName][id] = metric{
			labels:     labels,
			value:      value,
			metricType: prometheus.GaugeValue,
			isStatic:   static,
		}
	}
	m.Unlock()
}

// SetGaugeValue initialize or update metric value
func SetGaugeValue(metricName string, value float64, labels ...string) {
	setGaugeValue(false, metricName, value, labels...)

}

// SetStaticGaugeValue initialize or update a static metric
func SetStaticGaugeValue(metricName string, value float64, labels ...string) {
	setGaugeValue(true, metricName, value, labels...)
}

var promRegistry = prometheus.NewRegistry()

func init() {
	promRegistry.MustRegister(&customCollector{})
}

// StartMetricsEndpoint starts metrics endpoint
func StartMetricsEndpoint(o opts.MetricsEndpoint) error {

	server := http.NewServeMux()
	server.Handle(o.Path, promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}))

	if err := http.ListenAndServe(":"+strconv.Itoa(o.Port), server); err != nil {
		return err
	}
	return nil
}
