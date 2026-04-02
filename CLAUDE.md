# muke-cli

UESTC MOOC2 英语课考试自动提交工具。

## 常用命令

```bash
go run ./cmd/muke setup                           # 首次配置，选择英语课程
go run ./cmd/muke exams                           # 列出考试
go run ./cmd/muke exam <exam-id>                  # 查看题目和选项 ID
go run ./cmd/muke submit <exam-id> answers.json   # 提交答案
go run ./cmd/muke result <exam-id>                # 查看成绩
go run ./cmd/muke refresh                         # 清除 cookie 缓存并重新获取
```

## 项目结构

```
cmd/muke/main.go          CLI 入口，子命令分发
daemon/main.go            WebSocket 服务器，中转 CLI 和扩展的通信
internal/api/api.go       平台 HTTP API 封装
internal/api/types.go     请求/响应结构体
internal/client/client.go Cookie 获取（文件缓存 → Daemon → 扩展）
internal/config/config.go 本地配置读写（~/.muke/config.json，session 缓存在 ~/.muke/session）
internal/proto/message.go WebSocket 消息格式
extension/                Chrome 扩展（拦截请求头抓 cookie，连接 Daemon）
```

## 工作流程

1. 扩展连接 Daemon（ws://localhost:7788/extension），一直保持
2. CLI 运行时，先读 `~/.muke/session` 缓存，没有则连 Daemon 向扩展要 cookie
3. 拿到 cookie 后直接调用平台 API

## 错误处理

- **HTTP 404 / 401**：运行 `go run ./cmd/muke refresh` 刷新 cookie 后重试
- **扩展未连接**：确保 Chrome 已登录 MOOC2 并打开过课程页面
- **未配置课程**：运行 `go run ./cmd/muke setup`

## 提交考试流程（/submit-exam skill）

1. `muke exams` 列出考试，让用户选 exam_id
2. `muke exam <id>` 获取题目，记录 subject_id 和 option_id
3. 读取用户答案文件，生成 answers.json
4. 用户确认后 `muke submit <id> answers.json`
5. `muke result <id>` 查看成绩
