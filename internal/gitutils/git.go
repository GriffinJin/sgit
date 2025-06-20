package gitutils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func IsGitRepo(workDir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = workDir
	return cmd.Run() == nil
}

func ResetHard(workDir string) error {
	cmd := exec.Command("git", "reset", "--hard")
	cmd.Dir = workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reset失败: %w\n输出: %s", err, string(output))
	}
	return nil
}

func CleanUntracked(workDir string) error {
	cmd := exec.Command("git", "clean", "-f", "-d")
	cmd.Dir = workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("清理失败: %w\n输出: %s", err, string(output))
	}
	return nil
}

func RemoveTargetDir(workDir string) error {
	targetPath := filepath.Join(workDir, "target")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil
	}

	fmt.Printf("发现target目录: %s\n", targetPath)
	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}
	return nil
}

// 新增：递归查找所有Git仓库
func FindGitRepos(root string, excludeDirs []string) ([]string, error) {
	var gitRepos []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过指定的目录
		for _, dir := range excludeDirs {
			if strings.Contains(path, dir) {
				return filepath.SkipDir
			}
		}

		// 检查是否Git仓库（包含.git目录）
		if info.IsDir() && info.Name() == ".git" {
			// 返回仓库根目录
			repo := filepath.Dir(path)
			gitRepos = append(gitRepos, repo)
			return filepath.SkipDir // 跳过.git目录本身
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("查找Git仓库失败: %w", err)
	}

	return gitRepos, nil
}

// Fetch 执行 git fetch
func Fetch(workDir string) error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("fetch失败: %w\n输出: %s", err, string(output))
	}
	return nil
}

// Pull 执行 git pull
func Pull(workDir string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull失败: %w\n输出: %s", err, string(output))
	}
	return nil
}

// GetCurrentBranch 获取当前分支
func GetCurrentBranch(workDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取分支失败: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// PullRepo 获取特定仓库的更新
func PullRepo(workDir string) error {
	if err := Fetch(workDir); err != nil {
		return err
	}

	branch, err := GetCurrentBranch(workDir)
	if err != nil {
		return err
	}

	fmt.Printf("正在拉取 %s (%s)...\n", filepath.Base(workDir), branch)
	return Pull(workDir)
}

// GetRemoteInfo 获取Git仓库的远程信息
func GetRemoteInfo(workDir string) (map[string]string, error) {
	remotes := make(map[string]string)

	// 获取远程列表
	cmd := exec.Command("git", "remote")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取远程列表失败: %w", err)
	}

	// 解析远程列表
	remoteList := strings.Fields(string(output))
	if len(remoteList) == 0 {
		return remotes, nil
	}

	// 获取每个远程的URL
	for _, remoteName := range remoteList {
		cmd := exec.Command("git", "remote", "get-url", remoteName)
		cmd.Dir = workDir
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("获取远程URL失败: %w", err)
		}
		remotes[remoteName] = strings.TrimSpace(string(output))
	}

	return remotes, nil
}

// GetRepoName 从路径中提取仓库名称
func GetRepoName(path string) string {
	// 尝试从.git配置中获取仓库名
	repoName := ""

	configPath := filepath.Join(path, ".git", "config")
	if fileInfo, err := os.Stat(configPath); err == nil && !fileInfo.IsDir() {
		data, err := os.ReadFile(configPath)
		if err == nil {
			re := regexp.MustCompile(`(?m)^$$remote "origin"$$.*?url\s*=\s*.+/(.+?)\.git`)
			matches := re.FindStringSubmatch(string(data))
			if len(matches) > 1 {
				repoName = matches[1]
			}
		}
	}

	// 如果无法从.git/config中获取，则使用目录名
	if repoName == "" {
		repoName = filepath.Base(path)
	}

	return repoName
}

// GetBranches 获取所有分支信息
func GetBranches(workDir string) (current string, branches []string, err error) {
	// 获取当前分支
	current, err = GetCurrentBranch(workDir)
	if err != nil {
		return "", nil, err
	}

	// 获取所有分支
	cmd := exec.Command("git", "branch", "--all", "--format=%(refname:short)")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return "", nil, fmt.Errorf("获取分支列表失败: %w", err)
	}

	branches = strings.Split(strings.TrimSpace(string(output)), "\n")
	return current, branches, nil
}

// GetBranchStatus 获取分支状态（是否落后/领先/不同步）
func GetBranchStatus(workDir, branch string) string {
	// 获取上游分支
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	cmd.Dir = workDir
	upstream, err := cmd.Output()
	if err != nil {
		return "" // 没有上游分支
	}
	upstreamName := strings.TrimSpace(string(upstream))

	// 检查分支状态
	cmd = exec.Command("git", "rev-list", "--left-right", branch+"..."+upstreamName)
	cmd.Dir = workDir
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()

	if err != nil {
		return ""
	}

	counts := strings.Split(out.String(), "\n")
	ahead := 0
	behind := 0

	for _, line := range counts {
		if strings.HasPrefix(line, ">") {
			ahead++
		} else if strings.HasPrefix(line, "<") {
			behind++
		}
	}

	status := ""
	if ahead > 0 {
		status += fmt.Sprintf("↑%d ", ahead)
	}
	if behind > 0 {
		status += fmt.Sprintf("↓%d ", behind)
	}

	return strings.TrimSpace(status)
}

// IsDirty 检查仓库是否有未提交的更改
func IsDirty(workDir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// SwitchOrCreateBranch 切换或创建分支
func SwitchOrCreateBranch(workDir, branchName string, create bool) error {
	// 先尝试切换到已有分支
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = workDir
	err := cmd.Run()

	if err == nil {
		return nil
	}

	// 如果切换失败且允许创建
	if create {
		// 创建新分支
		cmd = exec.Command("git", "checkout", "-b", branchName)
		cmd.Dir = workDir
		return cmd.Run()
	}

	return fmt.Errorf("分支 %s 不存在", branchName)
}

// SwitchAndTrackBranch 切换到分支并设置上游跟踪
func SwitchAndTrackBranch(workDir, branchName, remoteName string) error {
	remoteBranch := remoteName + "/" + branchName

	// 检查远程分支是否存在
	cmd := exec.Command("git", "ls-remote", "--heads", remoteName, branchName)
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("检查远程分支失败: %w", err)
	}

	if strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("远程分支 %s 不存在", remoteBranch)
	}

	// 切换到跟踪分支
	cmd = exec.Command("git", "checkout", "-b", branchName, remoteBranch, "--track")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("切换到跟踪分支失败: %w", err)
	}

	return nil
}

// FetchAll 获取所有远程的最新信息
func FetchAll(workDir string) error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = workDir
	return cmd.Run()
}
