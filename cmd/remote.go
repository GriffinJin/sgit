/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/griffin.jin/sgit/internal/gitutils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	remotePath    string
	remoteExclude string
	remoteSimple  bool // 简化输出模式
	remoteRaw     bool // 原始输出模式
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote [路径]",
	Short: "查看所有Git项目的远程路径",
	Long: `递归显示当前目录下所有Git项目的远程URL信息

示例:
  # 查看所有Git项目的远程URL
  sgit remote

  # 使用简单表格格式
  sgit remote --simple

  # 原始输出（适合脚本处理）
  sgit remote --raw`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var rootDir string
		if remotePath != "" {
			rootDir = remotePath
		} else if len(args) > 0 {
			rootDir = args[0]
		} else {
			rootDir, _ = os.Getwd()
		}

		// 处理排除目录
		var excludeDirs []string
		if remoteExclude != "" {
			excludeDirs = strings.Split(remoteExclude, ",")
		}

		// 查找所有Git仓库
		repos, err := gitutils.FindGitRepos(rootDir, excludeDirs)
		if err != nil {
			fmt.Printf("查找Git仓库失败: %v\n", err)
			return
		}

		if len(repos) == 0 {
			fmt.Println("未找到任何Git仓库")
			return
		}

		if remoteRaw {
			outputRawRemoteInfo(repos)
		} else if remoteSimple {
			outputSimpleRemoteInfo(repos)
		} else {
			outputDetailedRemoteInfo(repos)
		}
	},
}

// 输出原始信息（适合脚本处理）
func outputRawRemoteInfo(repos []string) {
	for _, repoPath := range repos {
		remotes, err := gitutils.GetRemoteInfo(repoPath)
		if err != nil {
			fmt.Printf("%s: ERROR: %v\n", repoPath, err)
			continue
		}

		if len(remotes) == 0 {
			fmt.Printf("%s: NO_REMOTES\n", repoPath)
		} else {
			for name, url := range remotes {
				fmt.Printf("%s:%s:%s\n", repoPath, name, url)
			}
		}
	}
}

// 输出简化的表格
func outputSimpleRemoteInfo(repos []string) {
	fmt.Printf("\n%-25s %-30s %-25s %-10s\n",
		color.YellowString("Repository"),
		color.YellowString("Remote"),
		color.YellowString("URL"),
		color.YellowString("Type"))
	fmt.Println(strings.Repeat("-", 90))

	for _, repoPath := range repos {
		repoName := gitutils.GetRepoName(repoPath)
		remotes, err := gitutils.GetRemoteInfo(repoPath)

		if err != nil {
			fmt.Printf("%-25s %-30s %s\n",
				repoName,
				color.RedString("ERROR"),
				color.RedString("%v", err))
			continue
		}

		if len(remotes) == 0 {
			fmt.Printf("%-25s %-30s %s\n",
				repoName,
				color.BlueString("LOCAL_ONLY"),
				"-")
			continue
		}

		first := true
		for name, url := range remotes {
			remoteType := getRemoteType(url)
			if first {
				fmt.Printf("%-25s %-30s %-25s %s\n",
					repoName,
					name,
					shortenURL(url, 25),
					color.GreenString(remoteType))
				first = false
			} else {
				fmt.Printf("%-25s %-30s %-25s %s\n",
					"",
					name,
					shortenURL(url, 25),
					color.GreenString(remoteType))
			}
		}
	}
	fmt.Println(strings.Repeat("-", 90))
	fmt.Printf("Total repositories: %d\n", len(repos))
}

// 输出详细远程信息
func outputDetailedRemoteInfo(repos []string) {
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	// 获取当前工作目录作为rootDir的替代
	cwd, _ := os.Getwd()
	fmt.Printf("\nFound %s Git repositories\n\n", bold(len(repos)))

	for _, repoPath := range repos {
		// 计算相对于当前工作目录的相对路径
		relPath, err := filepath.Rel(cwd, repoPath)
		if err != nil {
			relPath = repoPath
		}

		fmt.Printf("%s\n", bold("► "+gitutils.GetRepoName(repoPath)))
		fmt.Printf("  Path: %s\n", cyan(relPath))

		remotes, err := gitutils.GetRemoteInfo(repoPath)
		if err != nil {
			fmt.Printf("  %s %v\n", color.RedString("Error:"), err)
			fmt.Println()
			continue
		}

		if len(remotes) == 0 {
			fmt.Printf("  %s\n", color.YellowString("No remotes configured"))
		} else {
			for name, url := range remotes {
				remoteType := getRemoteType(url)
				fmt.Printf("  %s: %s (%s)\n",
					green(name),
					url,
					green(remoteType))
			}
		}
		fmt.Println()
	}
}

// 缩短URL显示
func shortenURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}

// 判断远程仓库类型
func getRemoteType(url string) string {
	switch {
	case strings.Contains(url, "github.com"):
		return "GitHub"
	case strings.Contains(url, "gitlab.com"):
		return "GitLab"
	case strings.Contains(url, "bitbucket.org"):
		return "Bitbucket"
	case strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"):
		return "HTTP"
	case strings.HasPrefix(url, "git://"):
		return "GIT"
	case strings.Contains(url, "@"):
		return "SSH"
	default:
		return "Custom"
	}
}

func init() {
	remoteCmd.Flags().StringVarP(&remotePath, "path", "p", "", "指定搜索Git仓库的根路径")
	remoteCmd.Flags().StringVarP(&remoteExclude, "exclude", "e", "", "排除目录列表（逗号分隔）")
	remoteCmd.Flags().BoolVarP(&remoteSimple, "simple", "s", false, "简化表格输出")
	remoteCmd.Flags().BoolVarP(&remoteRaw, "raw", "r", false, "原始输出（适合脚本处理）")
	rootCmd.AddCommand(remoteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remoteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
