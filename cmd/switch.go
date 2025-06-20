/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/griffin.jin/sgit/internal/gitutils"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	switchPath    string
	switchExclude string
	switchCreate  bool
	switchTrack   bool
	switchRemote  string
	switchFetch   bool
	switchForce   bool
	switchSilent  bool // 静默模式（输出更少）
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [branch name]",
	Short: "切换所有Git项目的分支",
	Long: `递归切换当前目录下所有Git项目到指定分支

示例:
  # 切换到main分支
  sgit switch main

  # 切换到新分支（如果不存在则创建）
  sgit switch feature/new-api --create

  # 切换到远程分支并跟踪（如果本地分支不存在）
  sgit switch origin/release/1.0 --track

  # 强制切换（放弃本地修改）
  sgit switch hotfix --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branchName := args[0]
		var rootDir string
		if switchPath != "" {
			rootDir = switchPath
		} else {
			rootDir, _ = os.Getwd()
		}

		// 处理排除目录
		var excludeDirs []string
		if switchExclude != "" {
			excludeDirs = strings.Split(switchExclude, ",")
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

		if !switchSilent {
			fmt.Printf("将在 %d 个仓库中切换分支: %s\n", len(repos), color.YellowString(branchName))
		}

		// 确认操作（强制模式下不确认）
		if switchForce && !switchSilent {
			fmt.Println(color.RedString("警告: 强制切换将放弃所有未提交的更改!"))
		}

		if !switchSilent && (switchForce || !switchCreate) {
			fmt.Print("确认继续? (y/N): ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("操作已取消")
				return
			}
		}

		// 设置默认远程名称
		if switchTrack && switchRemote == "" {
			switchRemote = "origin"
		}

		// 创建等待组和结果通道
		var wg sync.WaitGroup
		resultChan := make(chan switchResult, len(repos))

		// 启动切换任务
		for _, repo := range repos {
			wg.Add(1)
			go func(repoPath string) {
				defer wg.Done()

				// 获取相对路径用于显示
				relPath, _ := filepath.Rel(rootDir, repoPath)
				if relPath == "." {
					relPath = filepath.Base(repoPath)
				}

				result := switchResult{
					path: relPath,
					name: gitutils.GetRepoName(repoPath),
				}

				// 执行切换逻辑
				if switchForce {
					// 强制切换：重置当前更改并尝试切换
					result.switchForce(repoPath, branchName)
				} else if switchTrack {
					// 切换到远程分支并跟踪
					result.switchWithTracking(repoPath, branchName)
				} else if switchCreate {
					// 切换到分支（如果不存在则创建）
					result.switchOrCreate(repoPath, branchName)
				} else {
					// 尝试直接切换
					result.switchOnly(repoPath, branchName)
				}

				resultChan <- result
			}(repo)
		}

		// 等待所有任务完成并收集结果
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// 处理并显示结果
		successCount := 0
		failures := []switchResult{}

		for result := range resultChan {
			if result.err != nil {
				failures = append(failures, result)
				if !switchSilent {
					fmt.Printf("%s %s: %s\n",
						color.RedString("✗"),
						result.name,
						color.RedString(result.err.Error()))
				}
			} else {
				successCount++
				if !switchSilent {
					fmt.Printf("%s %s: %s\n",
						color.GreenString("✓"),
						result.name,
						color.GreenString("切换到 ")+color.YellowString(result.message))
				}
			}
		}

		// 显示摘要
		if !switchSilent {
			fmt.Printf("\n切换结果: %d/%d 成功\n", successCount, len(repos))
		}

		// 显示失败详情
		if len(failures) > 0 && !switchSilent {
			fmt.Println("\n失败详情:")
			for _, fail := range failures {
				fmt.Printf("  %s (%s): %s\n",
					color.RedString(fail.name),
					fail.path,
					color.RedString(fail.err.Error()))
			}
		}

		if successCount == len(repos) && !switchSilent {
			fmt.Println(color.GreenString("\n✅ 所有仓库切换完成!"))
		}
	},
}

// 用于传递切换结果的结构体
type switchResult struct {
	name    string
	path    string
	message string
	err     error
}

// 直接切换分支
func (r *switchResult) switchOnly(repoPath, branch string) {
	if err := gitutils.SwitchOrCreateBranch(repoPath, branch, false); err != nil {
		r.err = fmt.Errorf("切换失败: %w", err)
	} else {
		r.message = branch
	}
}

// 切换到分支或创建新分支
func (r *switchResult) switchOrCreate(repoPath, branch string) {
	if err := gitutils.SwitchOrCreateBranch(repoPath, branch, true); err != nil {
		r.err = fmt.Errorf("切换失败: %w", err)
	} else {
		r.message = branch
	}
}

// 切换到分支并设置上游跟踪
func (r *switchResult) switchWithTracking(repoPath, branch string) {
	// 首先获取远程信息
	if switchFetch {
		if err := gitutils.FetchAll(repoPath); err != nil {
			r.err = fmt.Errorf("获取远程信息失败: %w", err)
			return
		}
	}

	if err := gitutils.SwitchAndTrackBranch(repoPath, branch, switchRemote); err != nil {
		r.err = fmt.Errorf("创建跟踪分支失败: %w", err)
	} else {
		r.message = "跟踪分支: " + branch
	}
}

// 强制切换（放弃所有更改）
func (r *switchResult) switchForce(repoPath, branch string) {
	// 1. 放弃所有更改
	if err := gitutils.ResetHard(repoPath); err != nil {
		r.err = fmt.Errorf("重置失败: %w", err)
		return
	}

	// 2. 清理未跟踪文件
	if err := gitutils.CleanUntracked(repoPath); err != nil {
		r.err = fmt.Errorf("清理未跟踪文件失败: %w", err)
		return
	}

	// 3. 尝试切换到分支（如果不存在则创建）
	if switchCreate {
		if err := gitutils.SwitchOrCreateBranch(repoPath, branch, true); err != nil {
			r.err = fmt.Errorf("切换失败: %w", err)
		} else {
			r.message = "创建并切换到: " + branch
		}
	} else {
		if err := gitutils.SwitchOrCreateBranch(repoPath, branch, false); err != nil {
			r.err = fmt.Errorf("切换失败: %w", err)
		} else {
			r.message = "切换到: " + branch
		}
	}
}

func init() {
	switchCmd.Flags().StringVarP(&switchPath, "path", "p", "", "指定搜索Git仓库的根路径")
	switchCmd.Flags().StringVarP(&switchExclude, "exclude", "e", "", "排除目录列表（逗号分隔）")
	switchCmd.Flags().BoolVarP(&switchCreate, "create", "c", false, "如果分支不存在则创建新分支")
	switchCmd.Flags().BoolVarP(&switchTrack, "track", "t", false, "切换到远程分支并跟踪")
	switchCmd.Flags().StringVarP(&switchRemote, "remote", "r", "origin", "使用的远程名称（默认为origin）")
	switchCmd.Flags().BoolVarP(&switchFetch, "fetch", "f", false, "切换前获取所有远程信息")
	switchCmd.Flags().BoolVarP(&switchForce, "force", "", false, "强制切换（放弃所有未提交更改）")
	switchCmd.Flags().BoolVarP(&switchSilent, "silent", "s", false, "静默模式（仅输出错误）")

	rootCmd.AddCommand(switchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// switchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// switchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
