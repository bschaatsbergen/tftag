package cmd

import (
	"os"

	"github.com/bschaatsbergen/tftag/pkg/core"
	"github.com/bschaatsbergen/tftag/pkg/model"
	"github.com/bschaatsbergen/tftag/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	example = "# Apply tags in the current working directory\n" +
		"tftag" +
		"\n" +
		"# Apply tags in a specific directory\n" +
		"tftag -d modules/iam\n"
)

var (
	version string

	flags model.Flags

	rootCmd = &cobra.Command{
		Use:     "tftag",
		Short:   "tftag - DRY approach to tagging Terraform resources",
		Version: version, // The version is set during the build by making using of `go build -ldflags`
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			core.Main(flags.Directory)
		},
	}
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "set log level to debug")
	rootCmd.PersistentFlags().StringVarP(&flags.Directory, "dir", "d", ".", "set the working directory for tftag to run in")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := utils.ConfigureLogLevel(flags.Debug); err != nil {
			return err
		}
		return nil
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
