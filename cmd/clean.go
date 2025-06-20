/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/griffin.jin/sgit/internal/gitutils"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	forceFlag     bool
	recursiveFlag bool
	excludeFlag   string
	pathFlag      string
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean [路径]",
	Short: "清理Git项目",
	Long: `执行深度清理操作，包括：
1. 重置所有更改 (git reset --hard)
2. 删除未跟踪文件 (git clean -f -d)
3. 删除target目录

支持递归模式：使用 --recursive 选项递归清理所有Git项目子目录`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var rootDir string
		if pathFlag != "" {
			rootDir = pathFlag
		} else if len(args) > 0 {
			rootDir = args[0]
		} else {
			rootDir, _ = os.Getwd()
		}

		// 确认操作
		if !forceFlag {
			fmt.Println("警告：此操作将永久删除所有未提交的更改！")
			fmt.Print("确认继续? (y/N): ")

			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("操作已取消")
				return
			}
		}

		// 处理排除目录
		var excludeDirs []string
		if excludeFlag != "" {
			excludeDirs = strings.Split(excludeFlag, ",")
			fmt.Printf("排除目录: %v\n", excludeDirs)
		}

		// 根据递归标志处理
		if recursiveFlag {
			repos, err := gitutils.FindGitRepos(rootDir, excludeDirs)
			if err != nil {
				fmt.Println(err)
				return
			}

			if len(repos) == 0 {
				fmt.Println("未找到任何Git仓库")
				return
			}

			fmt.Printf("找到 %d 个Git仓库:\n", len(repos))
			for _, repo := range repos {
				fmt.Printf("• %s\n", repo)
			}

			fmt.Println("\n开始清理所有仓库...")
			for i, repo := range repos {
				fmt.Printf("\n清理仓库 %d/%d: %s\n", i+1, len(repos), repo)
				if err := cleanRepo(repo); err != nil {
					fmt.Printf("清理失败: %v\n", err)
				} else {
					fmt.Printf("✅ %s 清理完成\n", repo)
				}
			}
			fmt.Println("\n✅ 所有仓库清理完成！")
		} else {
			if !gitutils.IsGitRepo(rootDir) {
				fmt.Println("错误：当前目录不是Git仓库")
				fmt.Println("提示：使用 --recursive 选项递归查找Git项目")
				return
			}

			fmt.Printf("清理仓库: %s\n", rootDir)
			if err := cleanRepo(rootDir); err != nil {
				fmt.Printf("清理失败: %v\n", err)
			} else {
				fmt.Printf("✅ %s 清理完成！\n", rootDir)
			}
		}
	},
}

func cleanRepo(repoPath string) error {
	if err := gitutils.ResetHard(repoPath); err != nil {
		return fmt.Errorf("重置失败: %w", err)
	}

	if err := gitutils.CleanUntracked(repoPath); err != nil {
		return fmt.Errorf("清理未跟踪文件失败: %w", err)
	}

	if err := gitutils.RemoveTargetDir(repoPath); err != nil {
		return fmt.Errorf("删除target目录失败: %w", err)
	}

	return nil
}

func init() {
	cleanCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "跳过确认提示")
	cleanCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "递归清理所有Git项目子目录")
	cleanCmd.Flags().StringVarP(&excludeFlag, "exclude", "e", "", "排除目录列表（逗号分隔）")
	cleanCmd.Flags().StringVarP(&pathFlag, "path", "p", "", "指定清理的根路径")
	rootCmd.AddCommand(cleanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cleanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
