package client

import (
	"muke-cli/internal/proto"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

const daemonURL = "ws://localhost:7788/cli"
const healthURL = "http://localhost:7788/health"

// sessionFile 返回 cookie 缓存文件路径（项目目录/.muke/session）
func sessionFile() string {
	root, err := findProjectRoot()
	if err != nil {
		// 找不到项目根目录时回退到当前目录
		root, _ = os.Getwd()
	}
	return filepath.Join(root, ".muke", "session")
}

// GetCookies 获取 Cookie：先读本地缓存，读不到再走 Daemon
func GetCookies(domain string) (string, error) {
	// 1. 尝试读本地缓存
	if cookie, err := os.ReadFile(sessionFile()); err == nil && len(cookie) > 0 {
		return string(cookie), nil
	}

	// 2. 确保 Daemon 在跑
	if err := ensureDaemon(); err != nil {
		return "", err
	}

	// 3. 走 Daemon → 扩展
	cookie, err := getFromDaemon(domain)
	if err != nil {
		return "", err
	}

	// 4. 保存到本地，下次直接用
	SaveSession(cookie)
	return cookie, nil
}

// ClearSession 删除本地缓存的 cookie
func ClearSession() {
	os.Remove(sessionFile())
}

// SaveSession 把 cookie 写到本地文件
func SaveSession(cookie string) {
	dir := filepath.Dir(sessionFile())
	os.MkdirAll(dir, 0700)
	os.WriteFile(sessionFile(), []byte(cookie), 0600)
}

// ensureDaemon 检查 Daemon 是否在运行，不在则启动它
func ensureDaemon() error {
	// 先 ping health 接口
	if daemonAlive() {
		return nil
	}

	fmt.Println("Daemon 未运行，正在启动...")

	// 找 daemon.exe（和 mooc.exe 同目录）或用 go run
	exe, err := os.Executable()
	if err == nil {
		daemonExe := filepath.Join(filepath.Dir(exe), "daemon.exe")
		if _, err := os.Stat(daemonExe); err == nil {
			cmd := exec.Command(daemonExe)
			cmd.Start()
			return waitDaemon()
		}
	}

	// 找不到 daemon.exe，用 go run（开发环境）
	// 找项目根目录（向上找到包含 go.mod 的目录）
	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("找不到项目根目录，请手动启动 Daemon: %w", err)
	}

	cmd := exec.Command("go", "run", "./daemon")
	cmd.Dir = root
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Daemon 失败: %w", err)
	}

	return waitDaemon()
}

// daemonAlive 检查 Daemon 是否响应
func daemonAlive() bool {
	resp, err := http.Get(healthURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// waitDaemon 等待 Daemon 启动（最多 5 秒）
func waitDaemon() error {
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		if daemonAlive() {
			fmt.Println("Daemon 已启动")
			return nil
		}
	}
	return fmt.Errorf("Daemon 启动超时")
}

// findProjectRoot 向上查找包含 go.mod 的目录
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("未找到 go.mod")
}

// getFromDaemon 连接 Daemon 获取 cookie
func getFromDaemon(domain string) (string, error) {
	conn, _, err := websocket.DefaultDialer.Dial(daemonURL, nil)
	if err != nil {
		return "", fmt.Errorf("连接 Daemon 失败: %w", err)
	}
	defer conn.Close()

	id := fmt.Sprintf("req-%d", time.Now().UnixNano())
	msg := proto.Message{ID: id, Action: "get_cookies", Domain: domain}
	data, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}

	_, resp, err := conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var msg2 proto.Message
	if err := json.Unmarshal(resp, &msg2); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if msg2.Error != "" {
		return "", fmt.Errorf("%s", msg2.Error)
	}
	return msg2.Result, nil
}
