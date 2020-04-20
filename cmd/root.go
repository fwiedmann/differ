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
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fwiedmann/differ/pkg/registries/worker"

	differController "github.com/fwiedmann/differ/pkg/controller"

	"github.com/fwiedmann/differ/pkg/registries"

	"github.com/fwiedmann/differ/pkg/config"
	kubernetes_client "github.com/fwiedmann/differ/pkg/kubernetes-client"
	"github.com/fwiedmann/differ/pkg/observer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var observerKindsToInit []observer.Kind

type controller interface {
	Start(ctx context.Context)
}

func init() {
	observerKindsToInit = append(observerKindsToInit, observer.AppV1Deployment, observer.AppV1DaemonSet, observer.AppV1StatefulSet)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "./config.yaml", "Path to differ config file")
	rootCmd.Flags().String("loglevel", "info", "Set loglevel. Default is info")
	rootCmd.Flags().Bool("devmode", false, "Run differ from outside a cluster")
}

var configFile string

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

		newTagChan := make(chan worker.Event, 100)
		registryStore := registries.NewRegistriesStore(newTagChan)
		ctx, cancel := context.WithCancel(context.Background())
		eventGenerator := observer.NewGenerator(kubernetesAPIClient, conf.Namespace)
		observerConfig := observer.NewObserverConfig(conf.Namespace, kubernetesAPIClient, eventGenerator, registryStore)

		err = initAllObservers(ctx, observerConfig)
		if err != nil {
			return err
		}

		controllerErrorChan := make(chan error)

		imageTagListener := differController.NewRegistryEventListener(newTagChan)

		startControllers(ctx, &imageTagListener)

		osNotifyChan := initOSNotifyChan()
		select {
		case osSignal := <-osNotifyChan:
			log.Warnf("received os %s signal, start  graceful shutdown of controller...", osSignal.String())
			cancel()
			return nil
		case err := <-controllerErrorChan:
			cancel()
			return err
		}
	},
}

func initAllObservers(ctx context.Context, observerConfig observer.Config) error {
	for _, observerKindToInit := range observerKindsToInit {
		o, err := observer.NewObserver(observerKindToInit, observerConfig)
		if err != nil {
			return err
		}
		go o.StartObserving(ctx)
	}
	return nil
}

func startControllers(ctx context.Context, controllers ...controller) {
	for _, c := range controllers {
		go c.Start(ctx)
	}
}

func initOSNotifyChan() <-chan os.Signal {
	notifyChan := make(chan os.Signal, 3)
	signal.Notify(notifyChan, syscall.SIGTERM, syscall.SIGINT)
	return notifyChan
}

// Execute executes the rootCmd
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
