// Daemon —— WebSocket 服务器，常驻后台，充当 muke-cli 和 Chrome 扩展的中间人
//
// 数据流：
//   muke-cli  →  [WebSocket]  →  Daemon  →  [WebSocket]  →  Chrome 扩展
//   muke-cli  ←  [WebSocket]  ←  Daemon  ←  [WebSocket]  ←  Chrome 扩展

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"muke-cli/internal/proto"

	"github.com/gorilla/websocket"
)

// Daemon 维护两种连接：
//   - extension：Chrome 扩展（唯一，长期保持）
//   - cli：muke-cli（每次命令运行时临时连接）
type Daemon struct {
	mu        sync.Mutex                    // 保护并发访问，相当于 Go 的 sync.Mutex
	extension *websocket.Conn               // 扩展的连接
	pending   map[string]chan proto.Message // 等待中的请求：requestID → 结果通道
}

func newDaemon() *Daemon {
	return &Daemon{
		pending: make(map[string]chan proto.Message),
	}
}

// upgrader 把普通 HTTP 连接升级为 WebSocket 连接
// 相当于 Go HTTP server 的 handler 升级器
var upgrader = websocket.Upgrader{
	// 允许所有来源（本地开发用，生产环境应该限制）
	CheckOrigin: func(r *http.Request) bool { return true },
}

// handleExtension 处理 Chrome 扩展的 WebSocket 连接
// 扩展连上来之后，一直保持连接，等待转发过来的请求
func (d *Daemon) handleExtension(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("扩展连接升级失败:", err)
		return
	}
	defer conn.Close()

	// 保存扩展连接
	d.mu.Lock()
	d.extension = conn
	d.mu.Unlock()

	log.Println("✅ Chrome 扩展已连接")

	// 持续读取扩展发来的消息（都是对 CLI 请求的响应）
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("扩展断开连接:", err)
			d.mu.Lock()
			d.extension = nil
			d.mu.Unlock()
			return
		}

		// 解析消息
		var msg proto.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println("消息解析失败:", err)
			continue
		}

		log.Printf("← 扩展返回 [%s]: %s", msg.ID, msg.Result[:min(len(msg.Result), 50)])

		// 找到等待这个 ID 的 CLI 请求，把结果发过去
		// 相当于 Go channel 的 send
		d.mu.Lock()
		ch, ok := d.pending[msg.ID]
		if ok {
			delete(d.pending, msg.ID)
		}
		d.mu.Unlock()

		if ok {
			ch <- msg // 唤醒等待的 CLI 请求
		}
	}
}

// handleCLI 处理 muke-cli 的 WebSocket 连接
// CLI 连上来，发一个请求，等结果，拿到后断开
func (d *Daemon) handleCLI(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("CLI 连接升级失败:", err)
		return
	}
	defer conn.Close()

	log.Println("→ CLI 已连接")

	// 读取 CLI 发来的请求
	_, data, err := conn.ReadMessage()
	if err != nil {
		log.Println("读取 CLI 消息失败:", err)
		return
	}

	var req proto.Message
	if err := json.Unmarshal(data, &req); err != nil {
		log.Println("请求解析失败:", err)
		return
	}

	log.Printf("→ CLI 请求 [%s]: action=%s domain=%s", req.ID, req.Action, req.Domain)

	// 检查扩展是否在线
	d.mu.Lock()
	ext := d.extension
	d.mu.Unlock()

	if ext == nil {
		// 扩展没连上，直接返回错误
		errMsg, _ := json.Marshal(proto.Message{ID: req.ID, Error: "Chrome 扩展未连接，请确认扩展已安装并启动"})
		conn.WriteMessage(websocket.TextMessage, errMsg)
		return
	}

	// 创建一个 channel 等待扩展的响应
	// 相当于 Go 的 make(chan , 1)
	ch := make(chan proto.Message, 1)

	d.mu.Lock()
	d.pending[req.ID] = ch
	d.mu.Unlock()

	// 把请求转发给扩展
	reqData, _ := json.Marshal(req)
	if err := ext.WriteMessage(websocket.TextMessage, reqData); err != nil {
		log.Println("转发给扩展失败:", err)
		d.mu.Lock()
		delete(d.pending, req.ID)
		d.mu.Unlock()
		return
	}

	// 等待扩展返回结果（阻塞，直到收到响应）
	result := <-ch

	// 把结果发回给 CLI
	resultData, _ := json.Marshal(result)
	conn.WriteMessage(websocket.TextMessage, resultData)

	log.Printf("✅ 请求 [%s] 完成", req.ID)
}

func main() {
	d := newDaemon()

	// 注册两个路由：
	//   /extension  ← Chrome 扩展连这里
	//   /cli        ← muke-cli 连这里
	http.HandleFunc("/extension", d.handleExtension)
	http.HandleFunc("/cli", d.handleCLI)

	// 健康检查，muke-cli 启动时用来判断 Daemon 是否在跑
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	port := "7788"
	fmt.Println("Daemon 启动，监听端口", port)
	fmt.Println("等待 Chrome 扩展连接...")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("启动失败:", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
