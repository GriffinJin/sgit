# sgit

增强版 Git 工具，支持批量管理当前目录及子目录下的多个 Git 仓库。

## 功能特性
- **批量显示仓库信息**：递归查找当前目录下所有 Git 仓库，显示其所有远程（支持多 remote）及当前分支，输出结构美观直观。
- **批量强制清理仓库**：对所有仓库执行 `git reset --hard` 和 `git clean -f`，一键还原所有改动。
- **批量拉取仓库**：对所有仓库执行 `git fetch && git pull`，一键同步所有仓库。

## 安装

### 方法一：下载预编译二进制文件（推荐）

1. 前往 [Releases](https://github.com/YOUR_USERNAME/sgit/releases) 页面
2. 下载对应你系统的二进制文件
3. 解压并添加到 PATH 环境变量

**Linux/macOS:**
```bash
# 下载并解压
wget https://github.com/YOUR_USERNAME/sgit/releases/latest/download/sgit-x86_64-unknown-linux-gnu
chmod +x sgit-x86_64-unknown-linux-gnu
sudo mv sgit-x86_64-unknown-linux-gnu /usr/local/bin/sgit
```

**Windows:**
```powershell
# 下载 sgit-x86_64-pc-windows-msvc.exe 并重命名为 sgit.exe
# 添加到系统 PATH
```

### 方法二：从源码编译

1. 确保已安装 [Rust](https://www.rust-lang.org/tools/install)。
2. 克隆本仓库并编译：

```bash
git clone https://github.com/YOUR_USERNAME/sgit.git
cd sgit
cargo build --release
```

3. 将可执行文件加入 PATH。

## 使用方法

```bash
# 显示所有仓库的远程地址和分支（支持多 remote，输出美观）
target/release/sgit info

# 输出示例：
# 找到 2 个 Git 仓库:
#
# repo1           [main]
#   > origin: git@github.com:user/repo1.git
#   > upstream: git@github.com:other/repo1.git
#
# repo2           [dev]
#   > origin: git@github.com:user/repo2.git
#
# 每个仓库之间有空行，分支用[]包裹，支持多 remote，结构清晰。

# 强制清理所有仓库的改动
target/release/sgit clean

# 拉取所有仓库（fetch + pull）
target/release/sgit pull
```

## 依赖
- [clap](https://crates.io/crates/clap)：命令行参数解析
- [anyhow](https://crates.io/crates/anyhow)：错误处理

## 注意事项
- `clean` 命令会丢弃所有未提交的更改，请谨慎使用！
- 仅会递归查找当前目录下的一级子目录。

## License
MIT 