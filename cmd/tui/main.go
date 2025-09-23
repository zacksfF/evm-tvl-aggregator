package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zacksfF/evm-tvl-aggregator/internal/logger"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "tvl-tui",
		Short: "TVL Aggregator Terminal User Interface",
		Long: `A beautiful terminal interface for monitoring TVL (Total Value Locked) 
across multiple DeFi protocols and blockchains.

Features:
- Real-time TVL monitoring
- Interactive protocol selection
- Beautiful charts and visualizations
- Multi-chain support`,
		Run: runTUI,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tvl-aggregator.yaml)")
	rootCmd.PersistentFlags().String("api-url", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")

	// Bind flags to viper
	viper.BindPFlag("api.url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("ui.no_color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Add subcommands
	rootCmd.AddCommand(dashboardCmd())
	rootCmd.AddCommand(monitorCmd())
	rootCmd.AddCommand(configCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".tvl-aggregator")
	}

	// Environment variables
	viper.SetEnvPrefix("TVL")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("api.url", "http://localhost:8080")
	viper.SetDefault("api.timeout", "30s")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("ui.refresh_interval", "5s")
	viper.SetDefault("ui.no_color", false)

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
	
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}
}

func runTUI(cmd *cobra.Command, args []string) {
	app, err := tui.NewApp()
	if err != nil {
		log.Fatalf("Failed to create TUI app: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("TUI app failed: %v", err)
	}
}

func dashboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard",
		Short: "Show TVL dashboard",
		Long:  "Display an interactive dashboard with TVL data across all protocols",
		Run: func(cmd *cobra.Command, args []string) {
			app, err := tui.NewApp()
			if err != nil {
				log.Fatalf("Failed to create TUI app: %v", err)
			}

			app.SetMode(tui.DashboardMode)
			if err := app.Run(); err != nil {
				log.Fatalf("Dashboard failed: %v", err)
			}
		},
	}
}

func monitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor [protocol]",
		Short: "Monitor specific protocol",
		Long:  "Monitor TVL changes for a specific protocol in real-time",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			app, err := tui.NewApp()
			if err != nil {
				log.Fatalf("Failed to create TUI app: %v", err)
			}

			if len(args) > 0 {
				app.SetProtocol(args[0])
			}

			app.SetMode(tui.MonitorMode)
			if err := app.Run(); err != nil {
				log.Fatalf("Monitor failed: %v", err)
			}
		},
	}

	cmd.Flags().String("chain", "", "Filter by specific chain")
	cmd.Flags().Duration("interval", 0, "Refresh interval")

	return cmd
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Manage TUI configuration settings",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Configuration file: %s\n", viper.ConfigFileUsed())
			fmt.Printf("API URL: %s\n", viper.GetString("api.url"))
			fmt.Printf("Log Level: %s\n", viper.GetString("log.level"))
			fmt.Printf("Refresh Interval: %s\n", viper.GetString("ui.refresh_interval"))
		},
	})

	return cmd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}