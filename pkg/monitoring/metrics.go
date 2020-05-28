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

package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	KubernetesObservedContainerMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "differ_kubernetes_observed_container",
		Help:        "Represents a container in the cluster with meta information about the parent kubernetes object e.g. a Deployment. If a container gets deleted the gauge value will be set to zero.",
		ConstLabels: nil,
	}, []string{"container_name", "registry_url", "image", "image_tag", "namespace", "parent_object_api_version", "parent_object_kind", "parent_object_uid", "parent_object_name"})

	OciImageNewerTagAvailableMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "differ_oci_image_new_tag_available",
		Help:        "Represents a oci image with the current and the latest available tag",
		ConstLabels: nil,
	}, []string{"image", "registry_url", "image_tag", "latest_tag", "tag_regex_expression"})
)

func MetricsHandler() http.Handler {
	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(prometheus.NewGoCollector(), prometheus.NewBuildInfoCollector(), KubernetesObservedContainerMetric, OciImageNewerTagAvailableMetric)
	return promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
}
