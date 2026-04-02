package proto

// Message 是 Daemon、CLI、扩展之间通信的统一格式
// 相当于 Go 里定义一个 RPC 消息结构体
type Message struct {
	ID     string `json:"id"`     // 请求 ID，用来匹配请求和响应
	Action string `json:"action"` // 动作，如 "get_cookies"
	Domain string `json:"domain"` // 目标域名
	//前面是发出方请求的，后面是响应结果回写
	Result string `json:"result"` // 扩展返回的结果（响应时填）
	Error  string `json:"error"`  // 错误信息（出错时填）
}
