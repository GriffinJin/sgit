/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/griffin.jin/sgit/internal/gitutils"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	pullRecursive bool
	pullPath      string
	pullExclude   string
	pullParallel  bool // 是否并行执行
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull [路径]",
	Short: "拉取所有Git项目的更新",
	Long: `递归拉取所有Git项目的更新（执行fetch && pull）
支持递归模式：使用 --recursive 选项递归拉取所有Git项目子目录

示例:
  # 拉取当前目录下所有Git项目的更新
  sgit pull --recursive

  # 拉取指定目录下所有Git项目的更新
  sgit pull --recursive --path ~/projects`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var rootDir string
		if pullPath != "" {
			rootDir = pullPath
		} else if len(args) > 0 {
			rootDir = args[0]
		} else {
			rootDir, _ = os.Getwd()
		}

		// 处理排除目录
		var excludeDirs []string
		if pullExclude != "" {
			excludeDirs = strings.Split(pullExclude, ",")
			fmt.Printf("排除目录: %v\n", excludeDirs)
		}

		// 查找所有Git仓库
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

		// 显示摘要信息
		baseRepo := filepath.Base(rootDir)
		baseRepos := len(repos)
		if rootDir == "." {
			cwd, _ := os.Getwd()
			baseRepo = filepath.Base(cwd)
		}

		fmt.Printf("\n将在 %s 目录下的 %d 个仓库执行拉取操作\n", baseRepo, baseRepos)
		fmt.Print("确认继续? (y/N): ")

		var confirm string
		_, _ = fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("操作已取消")
			return
		}

		fmt.Println("\n开始拉取...")

		// 使用工作池并行拉取
		if pullParallel {
			pullReposParallel(repos)
		} else {
			pullReposSequential(repos)
		}

		fmt.Println("\n✅ 所有仓库拉取完成！")
	},
}

// 顺序拉取仓库
func pullReposSequential(repos []string) {
	for i, repo := range repos {
		fmt.Printf("\n仓库 %d/%d: %s\n", i+1, len(repos), repo)
		if err := gitutils.PullRepo(repo); err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
		} else {
			fmt.Println("✅ 拉取成功")
		}
	}
}

// 并行拉取仓库（使用工作池）
func pullReposParallel(repos []string) {
	jobs := make(chan string, len(repos))
	results := make(chan struct {
		repo   string
		result string
	}, len(repos))

	// 创建4个工作goroutine
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for repo := range jobs {
				err := gitutils.PullRepo(repo)
				if err != nil {
					results <- struct {
						repo   string
						result string
					}{repo: repo, result: fmt.Sprintf("❌ 错误: %v", err)}
				} else {
					results <- struct {
						repo   string
						result string
					}{repo: repo, result: "✅ 拉取成功"}
				}
			}
		}(i)
	}

	// 发送任务
	for _, repo := range repos {
		jobs <- repo
	}
	close(jobs)

	// 等待所有工作完成
	wg.Wait()
	close(results)

	// 收集并打印结果
	fmt.Println("\n拉取结果:")
	var resultsList []struct {
		repo   string
		result string
	}
	for res := range results {
		resultsList = append(resultsList, res)
	}

	// 按仓库路径排序结果
	sort.Slice(resultsList, func(i, j int) bool {
		return resultsList[i].repo < resultsList[j].repo
	})

	for _, res := range resultsList {
		fmt.Printf("• %s: %s\n", filepath.Base(res.repo), res.result)
	}
}

func init() {
	pullCmd.Flags().BoolVarP(&pullRecursive, "recursive", "r", true, "递归拉取所有Git项目子目录（默认开启）")
	pullCmd.Flags().StringVarP(&pullPath, "path", "p", "", "指定查找Git仓库的根路径")
	pullCmd.Flags().StringVarP(&pullExclude, "exclude", "e", "", "排除目录列表（逗号分隔）")
	pullCmd.Flags().BoolVarP(&pullParallel, "parallel", "", false, "并行执行拉取操作（提高速度）")

	rootCmd.AddCommand(pullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
