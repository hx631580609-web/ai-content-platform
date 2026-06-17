$path = "e:\project\ai\ai-content-platform\admin\views\ai_assistant.html"
$content = Get-Content $path -Raw

$startTag = "<script>"
$endTag = "</script>"

$startIdx = $content.IndexOf($startTag)
$endIdx = $content.IndexOf($endTag, $startIdx)

if ($startIdx -lt 0 -or $endIdx -lt 0) {
  Write-Host "ERROR: script tags not found"
  exit 1
}

Write-Host "Replacing script block from $startIdx to $endIdx"

$newScript = @"
<script>
(function () {
  const chatHistory = [];
  let contentType = "article";
  const chatEl = document.getElementById("chat");

  function addRow(role, html, isError) {
    const wrap = document.createElement("div");
    wrap.className = "msg " + role;
    const av = document.createElement("div");
    av.className = "avatar";
    av.textContent = role === "user" ? "我" : "AI";
    const bubble = document.createElement("div");
    bubble.className = "bubble";
    if (isError) bubble.classList.add("error");
    bubble.innerHTML = html || "";
    wrap.appendChild(av);
    wrap.appendChild(bubble);
    chatEl.appendChild(wrap);
    chatEl.scrollTop = chatEl.scrollHeight;
    return bubble;
  }

  function esc(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  function renderMd(text) {
    let h = esc(text);
    h = h.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
    const lines = h.split("\n");
    let out = "";
    let inList = false;
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      if (/^\s*[-*+]\s+/.test(line)) {
        if (!inList) { out += "<ul>"; inList = true; }
        out += "<li>" + line.replace(/^\s*[-*+]\s+/, "") + "</li>";
      } else if (/^\s*\d+\.\s+/.test(line)) {
        if (!inList) { out += "<ul>"; inList = true; }
        out += "<li>" + line.replace(/^\s*\d+\.\s+/, "") + "</li>";
      } else {
        if (inList) { out += "</ul>"; inList = false; }
        out += line + "\n";
      }
    }
    if (inList) out += "</ul>";
    return out.replace(/\n/g, "<br>");
  }

  function row(label, value, color) {
    const safe = esc(value);
    const colorAttr = color ? " style=\"color:" + esc(color) + "\"" : "";
    return "<div class=\"row\"><span class=\"label\">" + esc(label) + "</span><span class=\"value\"" + colorAttr + ">" + safe + "</span></div>";
  }

  async function loadConfig() {
    try {
      const token = localStorage.getItem("token");
      const res = await fetch("/ai/config", {
        headers: token ? { "Authorization": "Bearer " + token } : {}
      });
      const data = await res.json();
      const card = document.getElementById("cfgCard");
      card.classList.add(data.configured ? "ok" : "bad");
      const baseUrl = data.base_url || "-";
      card.innerHTML =
        row("Provider", data.provider || "-") +
        row("Base URL", baseUrl) +
        row("Model", data.model || "-") +
        row("\u72B6\u6001", data.configured ? "\u5DF2\u914D\u7F6E \u2713" : "\u672A\u914D\u7F6E \u2717", data.configured ? "#10b981" : "#ef4444");
      document.getElementById("topInfo").textContent =
        data.configured ? "LLM: " + data.provider + " \u00B7 " + data.model
                        : "LLM \u672A\u914D\u7F6E - \u8BF7\u8BBE\u7F6E\u73AF\u5883\u53D8\u91CF LLM_API_KEY / LLM_BASE_URL";
    } catch (e) {
      document.getElementById("cfgCard").innerHTML =
        "<div style=\"color:#fca5a5\">\u65E0\u6CD5\u52A0\u8F7D\u914D\u7F6E: " + esc(e.message) + "</div>";
    }
  }

  function send() {
    const input = document.getElementById("user-input");
    const text = input.value.trim();
    if (!text) return;
    addRow("user", esc(text));
    input.value = "";
    chatHistory.push({ role: "user", content: text });

    const bubble = addRow("ai", "<span class=\"cursor\"></span>");
    document.getElementById("sendBtn").disabled = true;

    const messagesToSend = chatHistory.slice();
    let fullText = "";
    const token = localStorage.getItem("token");

    function handleEvent(raw) {
      const lines = raw.split("\n");
      let eventName = "";
      let dataStr = "";
      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.indexOf("event:") === 0) {
          eventName = line.slice(6).trim();
        } else if (line.indexOf("data:") === 0) {
          dataStr += line.slice(5).trim();
        }
      }
      if (!dataStr) return;
      let obj;
      try { obj = JSON.parse(dataStr); } catch (e) { return; }
      if (eventName === "delta" && obj.content) {
        fullText += obj.content;
        bubble.innerHTML = renderMd(fullText) + "<span class=\"cursor\"></span>";
        chatEl.scrollTop = chatEl.scrollHeight;
      } else if (eventName === "done") {
        bubble.innerHTML = renderMd(fullText);
        chatHistory.push({ role: "assistant", content: fullText });
        document.getElementById("sendBtn").disabled = false;
      } else if (eventName === "error") {
        bubble.innerHTML = renderMd("**错误: " + (obj.message || "未知错误") + "**");
        bubble.classList.add("error");
        document.getElementById("sendBtn").disabled = false;
      }
    }

    fetch("/ai/chat/stream", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? "Bearer " + token : ""
      },
      body: JSON.stringify({ messages: messagesToSend, stream: true, content_type: contentType })
    }).then(async function (res) {
      if (!res.ok) throw new Error("HTTP " + res.status);
      const reader = res.body.getReader();
      const decoder = new TextDecoder("utf-8");
      let buffer = "";
      while (true) {
        const chunk = await reader.read();
        if (chunk.done) break;
        buffer += decoder.decode(chunk.value, { stream: true });
        let idx;
        while ((idx = buffer.indexOf("\n\n")) >= 0) {
          handleEvent(buffer.slice(0, idx));
          buffer = buffer.slice(idx + 2);
        }
      }
      if (document.getElementById("sendBtn").disabled) {
        bubble.innerHTML = renderMd(fullText || "（无内容返回）");
        chatHistory.push({ role: "assistant", content: fullText });
        document.getElementById("sendBtn").disabled = false;
      }
    }).catch(function (err) {
      bubble.innerHTML = renderMd("**出错了: " + err.message + "**");
      bubble.classList.add("error");
      document.getElementById("sendBtn").disabled = false;
    });
  }

  document.getElementById("sendBtn").addEventListener("click", send);
  document.getElementById("user-input").addEventListener("keydown", function (e) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  });
  const chips = document.querySelectorAll(".prompt-chip");
  for (let i = 0; i < chips.length; i++) {
    chips[i].addEventListener("click", function () {
      document.getElementById("user-input").value = chips[i].getAttribute("data-prompt");
      document.getElementById("user-input").focus();
    });
  }
  const ctBtns = document.querySelectorAll(".ct-btn");
  for (let i = 0; i < ctBtns.length; i++) {
    ctBtns[i].addEventListener("click", function () {
      for (let j = 0; j < ctBtns.length; j++) ctBtns[j].classList.remove("active");
      ctBtns[i].classList.add("active");
      contentType = ctBtns[i].getAttribute("data-type");
      const map = { article: "文章", poster: "海报文案", video: "视频脚本" };
      document.getElementById("ctLabel").textContent = map[contentType] || "文章";
    });
  }

  loadConfig();
})();
</script>
"@

$prefix = $content.Substring(0, $startIdx)
$suffix = $content.Substring($endIdx + $endTag.Length)
$newContent = $prefix + $newScript + $suffix

[System.IO.File]::WriteAllText($path, $newContent)

Write-Host "Done! Script replaced successfully."
Write-Host "New script length: $($newScript.Length)"
