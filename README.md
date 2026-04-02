# muke-cli

用 Claude Code 自动完成 UESTC MOOC2 英语课考试。

## 架构

项目思路来自 [opencli](https://github.com/jackwener/opencli)，将 AI 工具与本地 CLI 结合，让 Claude Code 通过 skill 驱动整个考试流程。

使用前需要在浏览器中登录好 MOOC2 账号。浏览器侧通过加载的 Chrome 扩展脚本，拦截请求头中的 cookie（包括 HttpOnly 的 session）并缓存。本地运行一个持久的 Daemon 服务，作为扩展和 CLI 之间的中间人。CLI 则模仿常规命令行工具的设计，向平台 API 发送请求完成考试提交。

```
  浏览器 (Chrome)              本地进程              命令行
┌──────────────────┐  WS  ┌──────────────┐  WS   ┌─────────┐
│  Chrome 扩展      ◄────►    Daemon       ◄────►    CLI    │
│  background.js             (Go 常驻)              (Go)    │
└──────────────────┘      └──────────────┘       └─────────┘
    拦截请求头抓 cookie          中转                发 API 请求
```

**实际使用极简：**

- 首次运行多问答一次配置课程号
- 之后每次只需在 Claude Code 里执行 `/submit-exam`，告诉 Claude 答案文件路径和要考的考试，剩下全自动完成

> 目前唯一需要手动做的是：自己下载答案文件并提供路径。

> 最初版本支持更多课程自定义配置，为了方便 one-shot 执行做了精简。有需要完整版的可以联系我，或者自己 fork 修改。

> 以及还没有做更多的适配

## 前置条件

- Go 1.21+
- Chrome 浏览器
- Claude Code CLI

## 安装

### 1. 安装 Chrome 扩展

1. 打开 Chrome → 地址栏输入 `chrome://extensions/`
2. 开启右上角**开发者模式**
3. 点击**加载已解压的扩展程序**，选择本项目的 `extension/` 目录

### 2. 克隆项目

```bash
git clone https://github.com/JimmyBynex/muke-cli.git
cd muke-cli
```

### 3. 安装依赖

```bash
go mod tidy
```

## 使用

### 首次配置

确保 Chrome 扩展已启用，登录 MOOC2 后运行：

```bash
go run ./cmd/muke setup
```

选择你的英语课程 ID，保存到本地。

### 用 Claude Code 自动提交考试

在项目目录下打开 Claude Code，运行：

```
/submit-exam
```

Claude 会自动：
1. 列出考试列表，让你选择
2. 获取题目结构
3. 读取你准备的答案文件
4. 生成并确认 answers.json
5. 提交答案
6. 显示成绩

### 答案文件格式

准备一个 `.txt` 或 `.md` 文件，每行一道题的答案，例如：

```
1. A
2. B
3. C
4. 填空答案
```

### 手动命令

```bash
go run ./cmd/muke exams                           # 列出考试
go run ./cmd/muke exam <exam-id>                  # 查看题目
go run ./cmd/muke submit <exam-id> answers.json   # 提交答案
go run ./cmd/muke result <exam-id>                # 查看成绩
go run ./cmd/muke refresh                         # 刷新 cookie
```

## Cookie 说明

扩展通过拦截请求头捕获 `session` cookie（HttpOnly，无法通过 JS 直接读取）。

所有本地数据存在 `~/.muke/` 目录下：
- `~/.muke/session` — cookie 缓存
- `~/.muke/config.json` — 课程配置

cookie 过期后运行 `muke refresh` 重新获取。
