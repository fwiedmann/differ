package cmd

import (
	"github.com/fwiedmann/differ/pkg/controller"
	"github.com/fwiedmann/differ/pkg/opts"
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

		if err = c.Run(); err != nil {
			return err
		}
		return nil
	},
}

var configFile string
var logLevel string

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "./config.yaml", "Path to differ config file")
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Set loglevel. Default is info")
}

// Execute executes the rootCmd
func Execute() error {

	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
