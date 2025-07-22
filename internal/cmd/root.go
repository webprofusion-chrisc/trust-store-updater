package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/trust-store-updater/internal/config"
	"github.com/trust-store-updater/internal/updater"
)

var (
	cfgFile string
	dryRun  bool
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trust-store-updater",
	Short: "Cross-platform tool for updating certificate trust stores",
	Long: `Trust Store Updater is a cross-platform tool that can update operating system 
and application trust stores with new root certificates. It supports Linux, macOS, 
and Windows, and uses configuration to determine which target stores to update.`,
	RunE: runUpdate,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./trust-store-config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be updated without making changes")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	config.InitConfig(cfgFile)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	updaterService := updater.New(cfg, verbose, dryRun)
	return updaterService.UpdateTrustStores()
}
