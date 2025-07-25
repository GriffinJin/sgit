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
}

fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();

    match &cli.command {
        Commands::Info => info_command(".")?,
        Commands::Clean => clean_command(".")?,
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

    for dir in dirs {
        let output = Command::new("git")
            .arg("-C")
            .arg(&dir)
            .arg("remote")
            .arg("get-url")
            .arg("origin")
            .output();

        let remote = match output {
            Ok(out) if out.status.success() => String::from_utf8_lossy(&out.stdout).trim().to_string(),
            _ => "<no origin>".to_string(),
        };

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

        println!("仓库: {}\n  远程: {}\n  分支: {}\n", dir, remote, branch);
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
