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
	"github.com/fwiedmann/differ/pkg/controller"
	"github.com/fwiedmann/differ/pkg/opts"
	"github.com/fwiedmann/differ/pkg/scraper"
	"github.com/fwiedmann/differ/pkg/scraper/appv1scraper"
	"github.com/spf13/cobra"
)

var rootCmd = cobra.Command{
	Use:          "differ",
	Short:        "",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		o, err := opts.Init(configFile, logLevel)
		if err != nil {
			return err
		}

		c := controller.New(o)

		resourceStore := make(scraper.ResourceStore)

		if err = c.Run(scrapers, resourceStore); err != nil {
			return err
		}
		return nil
	},
}

var (
	scrapers []controller.ResourceScraper
)

var configFile string
var logLevel string

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "./config.yaml", "Path to differ config file")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Set loglevel. Default is info")
	scrapers = append(scrapers, appv1scraper.Deployment{})
}

// Execute executes the rootCmd
func Execute() error {

	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
