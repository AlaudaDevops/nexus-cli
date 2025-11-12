// Package cmd provides command-line interface for Nexus CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "nexus-cli",
	Short: "Nexus Repository Manager CLI Tool",
	Long: `nexus-cli is a command-line tool for managing Nexus Repository Manager.
It allows you to create users, repositories, roles, and permissions from YAML configuration files.

Set these environment variables before using:
  NEXUS_URL      - Nexus server URL (e.g., http://localhost:8081)
  NEXUS_USERNAME - Admin username
  NEXUS_PASSWORD - Admin password`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (required)")
}
