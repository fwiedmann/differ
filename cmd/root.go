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

package cmd

import (
	"github.com/fwiedmann/differ/pkg/config"
	"github.com/fwiedmann/differ/pkg/controller"
	"github.com/fwiedmann/differ/pkg/event"
	kubernetes_client "github.com/fwiedmann/differ/pkg/kubernetes-client"
	"github.com/fwiedmann/differ/pkg/metrics"
	"github.com/fwiedmann/differ/pkg/observer"
	"github.com/fwiedmann/differ/pkg/observer/appv1"
	"github.com/spf13/cobra"
)

var observers []controller.Observer

var rootCmd = cobra.Command{
	Use:          "differ",
	Short:        "",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		o, err := config.NewConfig(configFile, cmd.Version)
		if err != nil {
			return err
		}

		conf := o.GetConfig()
		isDevMode, err := cmd.Flags().GetBool("devmode")
		if err != nil {
			return err
		}

		kubernetesAPIClient, err := kubernetes_client.InitKubernetesAPIClient(isDevMode)
		if err != nil {
			return err
		}

		communicationChannels := event.InitCommunicationChannels(len(observers))
		observerConfig := observer.Config{
			NamespaceToScrape:                    conf.Namespace,
			KubernetesAPIClient:                  kubernetesAPIClient,
			KubernetesEventCommunicationChannels: communicationChannels,
			EventGenerator:                       event.NewGenerator(kubernetesAPIClient, conf.Namespace),
		}

		initAllObservers(observerConfig)
		controllerErrorChan := make(chan error)

		c := controller.NewDifferController(communicationChannels, controllerErrorChan, observers...)

		go func() {
			if err := metrics.StartMetricsEndpoint(conf.Metrics); err != nil {
				panic(err)
			}
		}()

		go c.StartController()

		return <-controllerErrorChan
	},
}

var configFile string

func init() {
	observers = append(observers, appv1.NewDeploymentObserver(), appv1.NewDaemonSetObserver(), appv1.NewStatefulSetObserver())
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "./config.yaml", "Path to differ config file")
	rootCmd.Flags().String("loglevel", "info", "Set loglevel. Default is info")
	rootCmd.Flags().Bool("devmode", false, "Run differ from outside a cluster")
}

func initAllObservers(observerConfig observer.Config) {
	for _, resourceObserver := range observers {
		resourceObserver.InitObserverWithKubernetesSharedInformer(observerConfig)
	}
}

// Execute executes the rootCmd
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
