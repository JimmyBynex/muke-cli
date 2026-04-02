# muke-cli 项目学习记录

## 项目架构

```
浏览器(Chrome)          本地进程              命令行
┌─────────────┐    WS    ┌────────┐    WS    ┌──────┐
│  扩展        │◄────────►│ Daemon │◄────────►│ CLI  │
│ background.js│          │ Go     │          │ Go   │
└─────────────┘          └────────┘          └──────┘
     抓 cookie                中转                发 API 请求
```

---

## 模块 A — Chrome 扩展（JS）

- [ ] Manifest V3 是什么，和 V2 的区别
- [ ] Service Worker 的生命周期（为什么内存会被清空）
- [ ] `webRequest` API：拦截请求、读请求体、读请求头
- [ ] `cookies` API：为什么拿不到 `session`（HttpOnly）
- [ ] `onBeforeSendHeaders` + `extraHeaders`：绕过 HttpOnly 限制
- [ ] 为什么扩展能拿到 session 而页面 JS 拿不到（权限模型）
- [ ] WebSocket 客户端：连接 Daemon、断线重连

---

## 模块 B — WebSocket 和 Daemon（Go）

- [ ] HTTP 和 WebSocket 的区别（握手、全双工、长连接）
- [ ] 为什么需要 Daemon（CLI 短进程 vs 扩展长连接）
- [ ] gorilla/websocket：Upgrade、ReadMessage、WriteMessage
- [ ] Go 并发基础：goroutine 是什么
- [ ] `sync.Mutex`：为什么 map 不是线程安全的
- [ ] RPC 模式：用 ID 匹配请求和响应
- [ ] `/extension` 和 `/cli` 两个 endpoint 的设计思路

---

## 模块 C — CLI 和 HTTP 客户端（Go）

- [ ] Go module 和包结构（module、package、import）
- [ ] 结构体（struct）和 JSON 序列化（encoding/json）
- [ ] HTTP 请求：`http.NewRequest`、设置 Header、读 Body
- [ ] 错误处理模式：`if err != nil`、`fatal()`
- [ ] `os.Args` 解析子命令
- [ ] 文件读写：`os.ReadFile`、`os.WriteFile`
- [ ] 接口设计：为什么拆成 `internal/api`、`internal/client`、`internal/config`

---

## 模块 D — 数据流全链路

- [ ] 从 `muke submit` 开始，完整调用链是什么
- [ ] 平台的考试流程：distribute → storage → multiple-subjects → submissions
- [ ] `exam_paper_instance_id` 和 `submission_id` 分别代表什么
- [ ] 答案格式：单选 `option_ids` vs 填空 `text_answers`
- [ ] Cookie 认证：为什么 `session` 是关键，Keycloak cookie 是什么

---

## 学习进度

| 模块 | 状态 | 备注 |
|------|------|------|
| A — Chrome 扩展 | 未开始 | |
| B — Daemon | 未开始 | |
| C — CLI | 未开始 | |
| D — 数据流 | 未开始 | |
