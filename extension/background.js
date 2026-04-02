// ============================================================
// background.js — Chrome 扩展的后台脚本（Service Worker）
//
// 它一直在后台运行，页面感知不到它，不会触发反调试检测。
// 相当于 Go 里的一个 goroutine，默默监听事件。
// ============================================================


// ------------------------------------------------------------
// 第一部分：监听网络请求
//
// chrome.webRequest.onBeforeRequest 是一个事件，
// 相当于 Go 里的 http.HandleFunc，每次请求经过就会触发。
// ------------------------------------------------------------

chrome.webRequest.onBeforeRequest.addListener(
  // 第一个参数：回调函数，每次有请求就调这里
  // details 是请求的详细信息，Chrome 自动填进来
  function(details) {

    // details 里有什么：
    //   details.url         请求的完整 URL
    //   details.method      GET / POST / PUT ...
    //   details.requestBody 请求体（POST 的数据在这里，GET 没有）
    //   details.requestId   这次请求的唯一 ID
    //   details.timeStamp   发生时间（毫秒时间戳）

    // 过滤掉静态资源和 HEAD 请求，只看 API
    if (details.method === "HEAD") return
    if (details.url.includes("/static/")) return

    console.log(`=== ${details.method} ===`)
    console.log("URL:", details.url)

    // requestBody 的结构（两种情况）：
    //
    // 情况 A：表单提交（application/x-www-form-urlencoded）
    //   details.requestBody.formData = { key: ["value"], ... }
    //
    // 情况 B：JSON 提交（application/json）
    //   details.requestBody.raw = [ { bytes: ArrayBuffer } ]
    //   需要自己把 bytes 转成字符串，见下面的 parseBody 函数

    const body = parseBody(details.requestBody)
    console.log("解析后的 Body:", JSON.stringify(body))
  },

  // 第二个参数：过滤条件，只监听这个域名下的请求
  { urls: ["https://mooc2.uestc.edu.cn/*"] },

  // 第三个参数：需要哪些额外信息
  // "requestBody" 表示我要看请求体，不加这个 details.requestBody 是 null
  ["requestBody"]
)


// ------------------------------------------------------------
// 第二部分：工具函数——把原始 bytes 转成字符串
//
// JSON 请求的 body 是二进制的 ArrayBuffer，
// 需要手动解码成字符串，再 JSON.parse 才能读。
// ------------------------------------------------------------

function parseBody(requestBody) {
  if (!requestBody) {
    return null
  }

  // 情况 A：表单数据，Chrome 已经帮你解析好了
  if (requestBody.formData) {
    return requestBody.formData
  }

  // 情况 B：原始字节（JSON 请求走这里）
  if (requestBody.raw && requestBody.raw.length > 0) {
    // raw 是数组，每项有一个 bytes 字段（ArrayBuffer 类型）
    const bytes = requestBody.raw[0].bytes

    // TextDecoder 相当于 Go 的 string([]byte{...})
    // 把二进制 bytes 解码成 UTF-8 字符串
    const text = new TextDecoder("utf-8").decode(bytes)

    // 尝试解析成 JSON 对象，方便后续使用
    try {
      return JSON.parse(text)
    } catch {
      // 不是 JSON，直接返回原始字符串
      return text
    }
  }

  return null
}


// ------------------------------------------------------------
// 第三部分：监听响应头（可选，用来抓 token）
//
// 有些平台把认证信息放在响应头里（如 Set-Cookie、Authorization）
// 如果只靠 Cookie 不够，可以在这里抓。
// 现在先注释掉，需要时再打开。
// ------------------------------------------------------------

// chrome.webRequest.onHeadersReceived.addListener(
//   function(details) {
//     console.log("响应头:", details.responseHeaders)
//   },
//   { urls: ["https://mooc2.uestc.edu.cn/*"] },
//   ["responseHeaders"]
// )


// ------------------------------------------------------------
// 第四部分：连接到本地 Daemon，等待指令
// ------------------------------------------------------------

// 缓存完整的 Cookie 字符串（包含 session 等 HttpOnly cookie）
let cachedCookies = ""

chrome.webRequest.onBeforeSendHeaders.addListener(
  function(details) {
    const h = details.requestHeaders.find(h => h.name.toLowerCase() === "cookie")
    if (h && h.value.includes("session=")) {
      cachedCookies = h.value
    }
  },
  //过滤条件
  { urls: ["https://mooc2.uestc.edu.cn/*"] },
  //需要的额外数据
  //一般直接使用requestheader就行，但是cookie需要extra
  ["requestHeaders", "extraHeaders"]
)

const DAEMON_URL = "ws://localhost:7788/extension"

function connectDaemon() {
  const ws = new WebSocket(DAEMON_URL)

  ws.onopen = () => {
    console.log("✅ 已连接到 Daemon")
  }

  // envent是浏览器传进来的
  // 收到 Daemon 转发来的 CLI 请求
  ws.onmessage = async (event) => {
    const req = JSON.parse(event.data)
    console.log("收到请求:", req)

    let result = ""
    let error = ""

    try {
      if (req.action === "get_cookies") {
        // 优先用缓存的完整 Cookie（含 session），fallback 到 chrome.cookies.getAll
        result = cachedCookies || await getCookies(req.domain)
      } else {
        error = "未知 action: " + req.action
      }
    } catch (e) {
      error = e.toString()
    }

    // 把结果发回给 Daemon（带上原始 req.id 用于匹配）
    ws.send(JSON.stringify({ id: req.id, result, error }))
  }

  ws.onclose = () => {
    console.log("Daemon 连接断开，3秒后重连...")
    setTimeout(connectDaemon, 3000)  // 自动重连
  }

  ws.onerror = (e) => {
    console.log("连接错误（Daemon 可能未启动）:", e.message)
  }
}

// 获取指定域名的所有 Cookie，返回 "key=value; key2=value2" 格式的字符串
function getCookies(domain) {
  return new Promise((resolve) => {
    chrome.cookies.getAll({ domain }, (cookies) => {
      const str = cookies.map(c => c.name + "=" + c.value).join("; ")
      resolve(str)
    })
  })
}

// 扩展启动时连接 Daemon
console.log("Campus CLI Bridge 已启动，监听 mooc2.uestc.edu.cn")
connectDaemon()
