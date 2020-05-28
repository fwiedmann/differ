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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fwiedmann/differ/pkg/monitoring"

	"k8s.io/client-go/kubernetes"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/informers"

	"github.com/fwiedmann/differ/pkg/observing"

	"github.com/fwiedmann/differ/pkg/config"
	kubernetes_client "github.com/fwiedmann/differ/pkg/kubernetes-client"

	"github.com/fwiedmann/differ/pkg/registry"

	"github.com/fwiedmann/differ/pkg/differentiating"

	"github.com/fwiedmann/differ/pkg/storage/memory"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		kubernetesAPIClient, err := kubernetes_client.InitKubernetesAPIClient(isDevMode)
		if err != nil {
			return err
		}

		if err := checkKubernetesAPIPermissions(ctx, kubernetesAPIClient, conf.Namespace); err != nil {
			return err
		}

		storage := memory.NewMemoryStorage()

		service := differentiating.NewOCIRegistryService(ctx, storage, func(c http.Client, img registry.OciImage) differentiating.OciRegistryAPIClient {
			return &registry.OciAPIClient{
				Image:  img,
				Client: c,
			}
		})
		event := make(chan differentiating.NotificationEvent)
		service.Notify(event)

		sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubernetesAPIClient, 0, informers.WithNamespace(conf.Namespace))
		appv1DaemonSetInformer := sharedInformerFactory.Apps().V1().DaemonSets().Informer()
		err = observing.StartKubernetesObserverService(ctx, kubernetesAPIClient, appv1DaemonSetInformer, conf.Namespace, observing.NewKubernetesAPPV1DaemonSetSerializer, service)
		if err != nil {
			return err
		}

		appv1DeploymentInformer := sharedInformerFactory.Apps().V1().Deployments().Informer()
		err = observing.StartKubernetesObserverService(ctx, kubernetesAPIClient, appv1DeploymentInformer, conf.Namespace, observing.NewKubernetesAPPV1DeploymentSerializer, service)
		if err != nil {
			return err
		}

		appv1StatefulSetInformer := sharedInformerFactory.Apps().V1().StatefulSets().Informer()
		err = observing.StartKubernetesObserverService(ctx, kubernetesAPIClient, appv1StatefulSetInformer, conf.Namespace, observing.NewKubernetesAPPV1StatefulSetSerializer, service)
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		mux.Handle("/metrics", monitoring.MetricsHandler())

		server := http.Server{
			Addr:              ":8080",
			Handler:           mux,
			ReadTimeout:       time.Second * 10,
			ReadHeaderTimeout: 0,
			WriteTimeout:      time.Second,
		}

		go func() {
			panic(server.ListenAndServe())
		}()

		osNotifyChan := initOSNotifyChan()
		for {
			select {
			case e := <-event:
				log.Debugf("%+v", e)
			case osSignal := <-osNotifyChan:
				log.Warnf("received os %s signal, start  graceful shutdown of controller...", osSignal.String())
				shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), time.Second*10)

				if err := server.Shutdown(shutdownCtx); err != nil {
					log.Error(err)
				}

				shutdownCtxCancel()
				cancel()
				return nil
			}
		}
	},
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

func checkKubernetesAPIPermissions(ctx context.Context, c kubernetes.Interface, namespace string) error {
	var isErr bool
	_, err := c.AppsV1().Deployments(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		isErr = true
		log.Error(err)
	}

	_, err = c.AppsV1().StatefulSets(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		isErr = true
		log.Error(err)
	}

	_, err = c.AppsV1().DaemonSets(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		isErr = true
		log.Error(err)
	}

	_, err = c.CoreV1().Secrets(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		isErr = true
		log.Error(err)
	}

	if isErr {
		return fmt.Errorf("differ error: please check your kubernetes API permissions")
	}
	return nil
}
