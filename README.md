# muke-cli

用 Claude Code 自动完成 UESTC MOOC2 英语课考试。

## 架构

```
浏览器(Chrome)          本地进程              命令行
┌─────────────┐    WS    ┌────────┐    WS    ┌──────┐
│  扩展        │◄────────►│ Daemon │◄────────►│ CLI  │
│ background.js│          │ Go     │          │ Go   │
└─────────────┘          └────────┘          └──────┘
     抓 cookie                中转                发 API 请求
```

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
git clone https://github.com/你的用户名/muke-cli.git
cd muke-cli
```

### 3. 安装依赖

```bash
go mod tidy
```

## 使用

### 首次配置

登录 MOOC2 后，随便打开一个课程页面（让扩展捕获 cookie），然后：

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

准备一个文本文件，每行一道题的答案，例如：

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

cookie 缓存在 `~/.muke/session`，过期后运行 `muke refresh` 重新获取。
