use clap::{Parser, Subcommand};
use std::fs;
use std::path::Path;
use std::process::Command;

#[derive(Parser)]
#[command(name = "sgit", about = "增强版 Git 工具")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// 显示当前目录及子目录下所有 Git 仓库的远程地址和分支
    Info,
    /// 强制清除当前目录下所有 Git 仓库的改动（reset + clean）
    Clean,
    /// 拉取当前目录下所有 Git 仓库（fetch + pull）
    Pull,
}

fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();

    match &cli.command {
        Commands::Info => info_command(".")?,
        Commands::Clean => clean_command(".")?,
        Commands::Pull => pull_command(".")?,
    }

    Ok(())
}

fn find_git_dirs(root: &str) -> anyhow::Result<Vec<String>> {
    let root_path = Path::new(root);
    let mut dirs = vec![];

    let mut to_check = vec![root_path.to_path_buf()];
    for entry in fs::read_dir(root_path)? {
        let entry = entry?;
        if entry.file_type()?.is_dir() {
            to_check.push(entry.path());
        }
    }

    for dir in to_check {
        let git_dir = dir.join(".git");
        if git_dir.exists() && git_dir.is_dir() {
            dirs.push(dir.to_string_lossy().to_string());
        }
    }

    Ok(dirs)
}

fn info_command(root: &str) -> anyhow::Result<()> {
    let dirs = find_git_dirs(root)?;

    if dirs.is_empty() {
        println!("当前目录下没有找到 Git 仓库");
        return Ok(());
    }

    println!("找到 {} 个 Git 仓库:\n", dirs.len());

    for dir in dirs {
        let output = Command::new("git")
            .arg("-C")
            .arg(&dir)
            .arg("rev-parse")
            .arg("--abbrev-ref")
            .arg("HEAD")
            .output();

        let branch = match output {
            Ok(out) if out.status.success() => String::from_utf8_lossy(&out.stdout).trim().to_string(),
            _ => "<unknown>".to_string(),
        };

        // 获取所有 remote 名称
        let remotes_output = Command::new("git")
            .arg("-C")
            .arg(&dir)
            .arg("remote")
            .output();
        let remotes = match remotes_output {
            Ok(out) if out.status.success() => {
                String::from_utf8_lossy(&out.stdout)
                    .lines()
                    .map(|s| s.trim().to_string())
                    .collect::<Vec<_>>()
            },
            _ => vec![],
        };

        // 获取仓库目录名，避免只显示"."
        let repo_name = if dir == "." {
            std::env::current_dir()
                .ok()
                .and_then(|p| p.file_name().and_then(|n| n.to_str().map(|s| s.to_string())))
                .or_else(|| std::env::current_dir().ok().map(|p| p.to_string_lossy().to_string()))
                .unwrap_or_else(|| dir.clone())
        } else {
            let path = std::path::Path::new(&dir);
            path.file_name()
                .and_then(|n| n.to_str().map(|s| s.to_string()))
                .unwrap_or_else(|| dir.clone())
        };
        println!("{:<15} [{}]", repo_name, branch);
        if remotes.is_empty() {
            println!("  > <no remote>");
        } else {
            for remote in remotes {
                let url_output = Command::new("git")
                    .arg("-C")
                    .arg(&dir)
                    .arg("remote")
                    .arg("get-url")
                    .arg(&remote)
                    .output();
                let url = match url_output {
                    Ok(out) if out.status.success() => String::from_utf8_lossy(&out.stdout).trim().to_string(),
                    _ => "<no url>".to_string(),
                };
                println!("  > {}: {}", remote, url);
            }
        }
        println!(""); // 每个仓库后加一空行
    }

    Ok(())
}

fn clean_command(root: &str) -> anyhow::Result<()> {
    let dirs = find_git_dirs(root)?;

    for dir in dirs {
        println!("清理仓库: {}", dir);

        // git reset --hard
        let reset_status = Command::new("git")
            .arg("-C")
            .arg(&dir)
            .arg("reset")
            .arg("HEAD")
            .arg("--hard")
            .status()?;

        // git clean -f
        let clean_status = Command::new("git")
            .arg("-C")
            .arg(&dir)
            .arg("clean")
            .arg("-f")
            .status()?;

        if reset_status.success() && clean_status.success() {
            println!("  ✅ 成功清理\n");
        } else {
            println!("  ❌ 清理失败\n");
        }
    }

    Ok(())
}

fn pull_command(root: &str) -> anyhow::Result<()> {
    let dirs = find_git_dirs(root)?;

    for dir in dirs {
        println!("拉取仓库: {}", dir);

        // git fetch
        let fetch_status = Command::new("git")
            .arg("-C")
            .arg(&dir)
            .arg("fetch")
            .status()?;

        // git pull
        let pull_status = Command::new("git")
            .arg("-C")
            .arg(&dir)
            .arg("pull")
            .status()?;

        if fetch_status.success() && pull_status.success() {
            println!("  ✅ 拉取成功\n");
        } else {
            println!("  ❌ 拉取失败\n");
        }
    }

    Ok(())
}
