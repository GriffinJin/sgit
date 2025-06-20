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
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	branchPath    string
	branchExclude string
	branchSimple  bool
	branchAll     bool // 显示所有分支
	branchSummary bool // 分支汇总
)

// branchCmd represents the branch command
var branchCmd = &cobra.Command{
	Use:   "branch [路径]",
	Short: "查看所有Git项目的分支信息",
	Long: `递归显示当前目录下所有Git项目的当前分支

示例:
  # 查看所有Git项目的当前分支
  sgit branch

  # 显示详细信息
  sgit branch --all

  # 使用简化表格格式
  sgit branch --simple

  # 显示分支汇总
  sgit branch --summary`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var rootDir string
		if branchPath != "" {
			rootDir = branchPath
		} else if len(args) > 0 {
			rootDir = args[0]
		} else {
			rootDir, _ = os.Getwd()
		}

		// 处理排除目录
		var excludeDirs []string
		if branchExclude != "" {
			excludeDirs = strings.Split(branchExclude, ",")
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

		if branchSimple {
			outputSimpleBranchInfo(repos, rootDir)
		} else if branchAll {
			outputDetailedBranchInfo(repos, rootDir)
		} else if branchSummary {
			outputBranchSummary(repos, rootDir)
		} else {
			outputCurrentBranches(repos, rootDir)
		}
	},
}

// 输出当前分支信息
func outputCurrentBranches(repos []string, rootDir string) {
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("\nFound %s Git repositories in %s\n\n",
		bold(len(repos)),
		bold(filepath.Base(rootDir)))

	for _, repoPath := range repos {
		// 获取仓库名称
		repoName := gitutils.GetRepoName(repoPath)

		// 计算相对于rootDir的相对路径
		relPath, err := filepath.Rel(rootDir, repoPath)
		if err != nil {
			relPath = repoPath
		}

		// 如果是当前目录，简化为.
		if relPath == "." {
			relPath = filepath.Base(repoPath)
		}

		// 获取当前分支
		branch, err := gitutils.GetCurrentBranch(repoPath)
		if err != nil {
			fmt.Printf("%s %s\n", cyan("► "+repoName), red("[ERROR: "+err.Error()+"]"))
			fmt.Printf("  Path: %s\n\n", cyan(relPath))
			continue
		}

		// 检查分支状态
		status := gitutils.GetBranchStatus(repoPath, branch)

		// 检查是否有未提交更改
		isDirty := ""
		if gitutils.IsDirty(repoPath) {
			isDirty = yellow(" (uncommitted changes)")
		}

		// 输出仓库信息
		branchColor := green
		if branch == "HEAD" || branch == "(detached from "+branch+")" {
			branchColor = yellow
		}

		fmt.Printf("%s %s%s\n",
			cyan("► "+repoName),
			branchColor(branch),
			isDirty)

		if status != "" {
			fmt.Printf("  Status: %s\n", status)
		}

		fmt.Printf("  Path: %s\n\n", cyan(relPath))
	}
}

// 输出所有分支详细信息
func outputDetailedBranchInfo(repos []string, rootDir string) {
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("\nFound %s Git repositories in %s\n\n",
		bold(len(repos)),
		bold(filepath.Base(rootDir)))

	for _, repoPath := range repos {
		// 获取仓库名称
		repoName := gitutils.GetRepoName(repoPath)

		// 计算相对于rootDir的相对路径
		relPath, err := filepath.Rel(rootDir, repoPath)
		if err != nil {
			relPath = repoPath
		}

		// 如果是当前目录，简化为.
		if relPath == "." {
			relPath = filepath.Base(repoPath)
		}

		// 获取分支信息
		currentBranch, allBranches, err := gitutils.GetBranches(repoPath)
		if err != nil {
			fmt.Printf("%s %s\n", cyan("► "+repoName), red("[ERROR: "+err.Error()+"]"))
			fmt.Printf("  Path: %s\n\n", cyan(relPath))
			continue
		}

		// 输出仓库头信息
		fmt.Printf("%s\n", cyan("► "+repoName))
		fmt.Printf("  Path: %s\n", cyan(relPath))

		// 检查是否有未提交更改
		isDirty := ""
		if gitutils.IsDirty(repoPath) {
			isDirty = yellow(" (uncommitted changes)")
		}

		// 输出当前分支
		fmt.Printf("  Current branch: %s%s\n", green(currentBranch), isDirty)

		// 输出上游状态
		status := gitutils.GetBranchStatus(repoPath, currentBranch)
		if status != "" {
			fmt.Printf("  Status: %s\n", status)
		}

		// 输出所有分支
		fmt.Println("  All branches:")

		// 分组本地和远程分支
		localBranches := []string{}
		remoteBranches := []string{}
		tags := []string{}

		for _, branch := range allBranches {
			switch {
			case strings.HasPrefix(branch, "remotes/"):
				remoteBranches = append(remoteBranches, strings.TrimPrefix(branch, "remotes/"))
			case strings.HasPrefix(branch, "tags/"):
				tags = append(tags, strings.TrimPrefix(branch, "tags/"))
			default:
				localBranches = append(localBranches, branch)
			}
		}

		// 排序
		sort.Strings(localBranches)
		sort.Strings(remoteBranches)
		sort.Strings(tags)

		// 输出本地分支
		if len(localBranches) > 0 {
			fmt.Println("    Local:")
			for _, branch := range localBranches {
				marker := " "
				if branch == currentBranch {
					marker = green("*")
				}
				fmt.Printf("      %s %s\n", marker, branch)
			}
		}

		// 输出远程分支
		if len(remoteBranches) > 0 {
			fmt.Println("    Remote:")
			for _, branch := range remoteBranches {
				fmt.Printf("      %s\n", yellow(branch))
			}
		}

		// 输出标签
		if len(tags) > 0 {
			fmt.Println("    Tags:")
			for _, tag := range tags {
				fmt.Printf("      %s\n", cyan(tag))
			}
		}

		fmt.Println()
	}
}

// 输出简化表格
func outputSimpleBranchInfo(repos []string, rootDir string) {
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("\n%s\n", bold("Git repository branches"))
	fmt.Printf("%s\n", strings.Repeat("-", 80))
	fmt.Printf("%-30s %-25s %-20s %s\n",
		color.CyanString("Repository"),
		color.CyanString("Branch"),
		color.CyanString("Status"),
		color.CyanString("Changes"))
	fmt.Println(strings.Repeat("-", 80))

	for _, repoPath := range repos {
		repoName := gitutils.GetRepoName(repoPath)

		// 获取当前分支
		branch, err := gitutils.GetCurrentBranch(repoPath)
		if err != nil {
			fmt.Printf("%-30s %s\n",
				repoName,
				red("ERROR: "+err.Error()))
			continue
		}

		// 获取分支状态
		status := gitutils.GetBranchStatus(repoPath, branch)

		// 检查是否有未提交更改
		isDirty := ""
		if gitutils.IsDirty(repoPath) {
			isDirty = yellow("✗")
		}

		// 突出显示 detached HEAD 状态
		branchColor := green
		if branch == "HEAD" || strings.HasPrefix(branch, "(detached from") {
			branchColor = yellow
		}

		fmt.Printf("%-30s %-25s %-20s %s\n",
			repoName,
			branchColor(branch),
			status,
			isDirty)
	}
	fmt.Println(strings.Repeat("-", 80))
}

// 输出分支汇总信息
func outputBranchSummary(repos []string, rootDir string) {
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("\n%s\n", bold("Branch summary for Git repositories"))
	fmt.Printf("%s\n", strings.Repeat("-", 60))

	// 收集分支统计信息
	branchCount := make(map[string]int)
	detachedRepos := []string{}
	errorRepos := []string{}
	dirtyRepos := []string{}

	for _, repoPath := range repos {
		repoName := gitutils.GetRepoName(repoPath)

		// 获取当前分支
		branch, err := gitutils.GetCurrentBranch(repoPath)
		if err != nil {
			errorRepos = append(errorRepos, repoName)
			continue
		}

		// 分离头指针状态
		if branch == "HEAD" || strings.HasPrefix(branch, "(detached from") {
			detachedRepos = append(detachedRepos, repoName)
			continue
		}

		// 记录分支
		branchCount[branch]++

		// 记录有未提交更改的仓库
		if gitutils.IsDirty(repoPath) {
			dirtyRepos = append(dirtyRepos, repoName)
		}
	}

	// 输出分支统计
	if len(branchCount) > 0 {
		fmt.Printf("%s\n", green("Branches:"))
		for branch, count := range branchCount {
			fmt.Printf("  %-30s: %d repositories\n", branch, count)
		}
		fmt.Println()
	}

	// 输出分离头指针的仓库
	if len(detachedRepos) > 0 {
		fmt.Printf("%s (%d repositories)\n", yellow("Detached HEAD"), len(detachedRepos))
		for i, repo := range detachedRepos {
			fmt.Printf("  %d. %s\n", i+1, repo)
		}
		fmt.Println()
	}

	// 输出有错误的仓库
	if len(errorRepos) > 0 {
		fmt.Printf("%s (%d repositories)\n", red("Errors"), len(errorRepos))
		for i, repo := range errorRepos {
			fmt.Printf("  %d. %s\n", i+1, repo)
		}
		fmt.Println()
	}

	// 输出有未提交更改的仓库
	if len(dirtyRepos) > 0 {
		fmt.Printf("%s (%d repositories)\n", yellow("Uncommitted changes"), len(dirtyRepos))
		for i, repo := range dirtyRepos {
			fmt.Printf("  %d. %s\n", i+1, repo)
		}
		fmt.Println()
	}

	fmt.Printf("%s\n", strings.Repeat("-", 60))
	fmt.Printf("Total repositories: %d\n", len(repos))
}

func init() {
	branchCmd.Flags().StringVarP(&branchPath, "path", "p", "", "指定搜索Git仓库的根路径")
	branchCmd.Flags().StringVarP(&branchExclude, "exclude", "e", "", "排除目录列表（逗号分隔）")
	branchCmd.Flags().BoolVarP(&branchSimple, "simple", "s", false, "简化表格输出")
	branchCmd.Flags().BoolVarP(&branchAll, "all", "a", false, "显示所有分支信息")
	branchCmd.Flags().BoolVarP(&branchSummary, "summary", "m", false, "显示分支汇总统计")

	rootCmd.AddCommand(branchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// branchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// branchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
