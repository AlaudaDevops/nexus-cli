package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/alauda/nexus-cli/pkg/config"
	"github.com/alauda/nexus-cli/pkg/nexus"
	"github.com/alauda/nexus-cli/pkg/output"
	"github.com/alauda/nexus-cli/pkg/service"
)

var (
	forceDelete bool
	dryRun      bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources defined in YAML configuration",
	Long: `Delete removes users, repositories, roles, and privileges defined in the YAML configuration file.
This is useful for cleaning up resources created by the apply command.`,
	Example: `  # Delete resources from file
  nexus-cli delete -c config.yaml

  # Dry run (show what would be deleted without actually deleting)
  nexus-cli delete -c config.yaml --dry-run

  # Force delete without confirmation
  nexus-cli delete -c config.yaml --force

  # With environment variables
  export NEXUS_URL=http://localhost:8081
  export NEXUS_USERNAME=admin
  export NEXUS_PASSWORD=admin123
  nexus-cli delete -c config.yaml`,
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete without confirmation")
	deleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")
}

func runDelete(_ *cobra.Command, _ []string) error {
	if cfgFile == "" {
		return fmt.Errorf("config file is required, use -c or --config flag")
	}

	// 创建格式化器
	formatter := output.NewFormatter(output.FormatText, os.Stdout)

	startTime := time.Now()

	// 检查配置文件是否存在
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", cfgFile)
	}

	// 获取 Nexus 认证信息
	url, username, password, err := config.GetNexusCredentials()
	if err != nil {
		return fmt.Errorf("failed to get Nexus credentials: %w", err)
	}

	formatter.Info(fmt.Sprintf("Connecting to Nexus at %s...", url))

	// 创建 Nexus 客户端
	client := nexus.NewClient(url, username, password)

	// 检查连接
	if err := client.CheckConnection(); err != nil {
		return fmt.Errorf("failed to connect to Nexus: %w", err)
	}

	formatter.Success("Successfully connected to Nexus")

	// 加载配置文件
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	formatter.Info(fmt.Sprintf("Loaded configuration from %s", cfgFile))

	// Dry run 模式
	if dryRun {
		formatter.Warning("DRY RUN MODE - No resources will be deleted")
		return showDeletePlan(cfg, formatter)
	}

	// 确认删除（除非使用 --force）
	if !forceDelete {
		fmt.Println("\nThe following resources will be DELETED:")
		_ = showDeletePlan(cfg, formatter)
		fmt.Print("\nAre you sure you want to delete these resources? (yes/no): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			formatter.Info("Delete canceled")
			return nil
		}
	}

	// 创建删除服务并执行
	svc := service.NewDeleteService(client, cfg, formatter)
	result, err := svc.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete resources: %w", err)
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 输出总结
	summary := &output.Summary{
		Total:     result.Total,
		Success:   result.Success,
		Failed:    result.Failed,
		Skipped:   result.Skipped,
		Errors:    result.Errors,
		Warnings:  result.Warnings,
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   time.Now().Format(time.RFC3339),
		Duration:  duration.String(),
	}

	return formatter.PrintSummary(summary)
}

func showDeletePlan(cfg *config.Config, _ *output.Formatter) error {
	if len(cfg.Users) > 0 {
		fmt.Println("\nUsers:")
		for _, user := range cfg.Users {
			fmt.Printf("  - %s\n", user.ID)
		}
	}

	if len(cfg.Repositories) > 0 {
		fmt.Println("\nRepositories:")
		for _, repo := range cfg.Repositories {
			fmt.Printf("  - %s (%s/%s)\n", repo.Name, repo.Format, repo.Type)
		}
	}

	if len(cfg.Roles) > 0 {
		fmt.Println("\nRoles:")
		for _, role := range cfg.Roles {
			fmt.Printf("  - %s\n", role.ID)
		}
	}

	if len(cfg.Privileges) > 0 {
		fmt.Println("\nPrivileges:")
		for _, priv := range cfg.Privileges {
			fmt.Printf("  - %s\n", priv.Name)
		}
	}

	return nil
}
