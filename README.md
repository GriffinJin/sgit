# sgit

增强版 Git 工具，支持批量管理当前目录及子目录下的多个 Git 仓库。

## 功能特性
- **批量显示仓库信息**：递归查找当前目录下所有 Git 仓库，显示其远程地址和当前分支。
- **批量强制清理仓库**：对所有仓库执行 `git reset --hard` 和 `git clean -f`，一键还原所有改动。

## 安装

1. 确保已安装 [Rust](https://www.rust-lang.org/tools/install)。
2. 克隆本仓库并编译：

```bash
git clone <your_repo_url>
cd sgit
cargo build --release
```

3. 可选：将可执行文件加入 PATH。

## 使用方法

```bash
# 显示所有仓库的远程地址和分支
target/release/sgit info

# 强制清理所有仓库的改动
target/release/sgit clean
```

## 依赖
- [clap](https://crates.io/crates/clap)：命令行参数解析
- [anyhow](https://crates.io/crates/anyhow)：错误处理

## 注意事项
- `clean` 命令会丢弃所有未提交的更改，请谨慎使用！
- 仅会递归查找当前目录下的一级子目录。

## License
MIT 