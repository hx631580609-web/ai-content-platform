(function() {
  var chatHistory = [];
  var contentMode = "article";
  var chatEl = document.getElementById("chat");
  var ctLabelEl = document.getElementById("ctLabel");

  var typeLabels = { "article": "文章", "poster": "海报", "video": "脚本" };

  function addMessage(role, text, isError) {
    var wrap = document.createElement("div");
    wrap.className = "msg " + role;
    var av = document.createElement("div");
    av.className = "avatar";
    av.textContent = role === "user" ? "我" : "AI";
    var bubble = document.createElement("div");
    bubble.className = "bubble";
    if (isError) bubble.classList.add("error");
    bubble.textContent = text;
    wrap.appendChild(av);
    wrap.appendChild(bubble);
    chatEl.appendChild(wrap);
    chatEl.scrollTop = chatEl.scrollHeight;
    return bubble;
  }

  function getToken() {
    return localStorage.getItem("token") || "";
  }

  function loadConfig() {
    var token = getToken();
    var headers = {};
    if (token) headers["Authorization"] = "Bearer " + token;
    fetch("/ai/config", { headers: headers })
      .then(function(res) { return res.json(); })
      .then(function(data) {
        var card = document.getElementById("cfgCard");
        card.classList.add(data.configured ? "ok" : "bad");
        var rows = "";
        rows += "<div>提供商：" + (data.provider || "-") + "</div>";
        rows += "<div>接口地址：" + (data.base_url || "-") + "</div>";
        rows += "<div>模型：" + (data.model || "-") + "</div>";
        rows += "<div>状态：" + (data.configured ? "已就绪" : "未配置") + "</div>";
        card.innerHTML = rows;
        document.getElementById("topInfo").textContent =
          data.configured ? "大模型：" + data.provider + " / " + data.model : "大模型未配置 - 请设置 LLM_API_KEY 和 LLM_BASE_URL 环境变量";
      })
      .catch(function(e) {
        var card = document.getElementById("cfgCard");
        card.innerHTML = "<div>配置加载失败：" + e.message + "</div>";
      });
  }

  function sendMessage() {
    var input = document.getElementById("user-input");
    var text = input.value.trim();
    if (!text) return;
    addMessage("user", text);
    input.value = "";
    chatHistory.push({ role: "user", content: text });

    var bubble = addMessage("ai", "思考中...");
    document.getElementById("sendBtn").disabled = true;

    var messagesToSend = chatHistory.slice();
    var fullText = "";
    var token = getToken();

    var headers = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = "Bearer " + token;

    var body = JSON.stringify({ messages: messagesToSend, stream: true, content_type: contentMode });

    fetch("/ai/chat/stream", {
      method: "POST",
      headers: headers,
      body: body
    }).then(function(res) {
      if (!res.ok) {
        return res.text().then(function(text) {
          throw new Error("HTTP " + res.status + ": " + text);
        });
      }
      var reader = res.body.getReader();
      var decoder = new TextDecoder("utf-8");
      var buffer = "";

      function processChunk(chunk) {
        if (chunk.done) {
          chatHistory.push({ role: "assistant", content: fullText });
          document.getElementById("sendBtn").disabled = false;
          // 显示操作按钮
          showActionButtons(bubble, fullText);
          return;
        }
        buffer += decoder.decode(chunk.value, { stream: true });
        var idx;
        while ((idx = buffer.indexOf("\n\n")) >= 0) {
          var eventText = buffer.slice(0, idx);
          buffer = buffer.slice(idx + 2);
          var lines = eventText.split("\n");
          var eventName = "";
          var dataStr = "";
          for (var i = 0; i < lines.length; i++) {
            var line = lines[i];
            if (line.indexOf("event:") === 0) {
              eventName = line.slice(6).trim();
            } else if (line.indexOf("data:") === 0) {
              dataStr += line.slice(5).trim();
            }
          }
          if (!dataStr) continue;
          var obj = null;
          try { obj = JSON.parse(dataStr); } catch (e) { continue; }
          if (eventName === "delta" && obj.content) {
            fullText += obj.content;
            bubble.textContent = fullText;
            chatEl.scrollTop = chatEl.scrollHeight;
          } else if (eventName === "done") {
            bubble.textContent = fullText;
            chatHistory.push({ role: "assistant", content: fullText });
            document.getElementById("sendBtn").disabled = false;
          } else if (eventName === "error") {
            bubble.textContent = "错误：" + (obj.message || "未知");
            bubble.classList.add("error");
            document.getElementById("sendBtn").disabled = false;
          }
        }
        return reader.read().then(processChunk);
      }

      return reader.read().then(processChunk);
    }).catch(function(err) {
      bubble.textContent = "请求失败：" + err.message;
      bubble.classList.add("error");
      document.getElementById("sendBtn").disabled = false;
    });
  }

  document.getElementById("sendBtn").addEventListener("click", sendMessage);
  document.getElementById("user-input").addEventListener("keydown", function(e) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  });

  var chips = document.querySelectorAll(".prompt-chip");
  for (var i = 0; i < chips.length; i++) {
    (function(chip) {
      chip.addEventListener("click", function() {
        document.getElementById("user-input").value = chip.getAttribute("data-prompt");
        document.getElementById("user-input").focus();
      });
    })(chips[i]);
  }

  var ctBtns = document.querySelectorAll(".ct-btn");
  for (var i = 0; i < ctBtns.length; i++) {
    (function(btn) {
      btn.addEventListener("click", function() {
        for (var j = 0; j < ctBtns.length; j++) ctBtns[j].classList.remove("active");
        btn.classList.add("active");
        contentMode = btn.getAttribute("data-type");
        if (ctLabelEl) ctLabelEl.textContent = typeLabels[contentMode] || contentMode;
      });
    })(ctBtns[i]);
  }

  var backBtn = document.getElementById("backBtn");
  if (backBtn) {
    backBtn.addEventListener("click", function() {
      location.href = "/admin/dashboard";
    });
  }

  loadConfig();

  // ========== 操作按钮相关功能 ==========

  var currentAIResponse = "";

  function showActionButtons(bubbleEl, responseText) {
    currentAIResponse = responseText;
    
    // 移除已存在的按钮
    var existingButtons = bubbleEl.querySelector('.action-buttons');
    if (existingButtons) {
      existingButtons.remove();
    }
    
    // 创建按钮容器
    var buttonsDiv = document.createElement('div');
    buttonsDiv.className = 'action-buttons';
    
    // 根据内容类型显示不同的按钮
    var buttons = [];
    if (contentMode === 'article') {
      buttons = [
        { id: 'saveAsContent', text: '💾 保存为内容', type: 'primary' },
        { id: 'publishToWebsite', text: '🌐 发布到网站', type: 'success' },
        { id: 'convertToVideo', text: '🎬 转换为视频脚本', type: '' }
      ];
    } else if (contentMode === 'video') {
      buttons = [
        { id: 'saveAsContent', text: '💾 保存为脚本', type: 'primary' },
        { id: 'generateArticle', text: '📝 生成文章', type: '' }
      ];
    } else if (contentMode === 'poster') {
      buttons = [
        { id: 'saveAsContent', text: '💾 保存为海报文案', type: 'primary' },
        { id: 'generateImage', text: '🎨 生成海报图片', type: '' }
      ];
    }
    
    // 添加通用按钮
    buttons.push({ id: 'copyContent', text: '📋 复制内容', type: '' });
    
    // 创建按钮
    buttons.forEach(function(btn) {
      var button = document.createElement('button');
      button.className = 'action-btn' + (btn.type ? ' ' + btn.type : '');
      button.innerHTML = btn.text;
      button.onclick = function() {
        handleActionButtonClick(btn.id, responseText);
      };
      buttonsDiv.appendChild(button);
    });
    
    bubbleEl.appendChild(buttonsDiv);
  }

  function handleActionButtonClick(actionId, content) {
    switch(actionId) {
      case 'saveAsContent':
        saveContentToLibrary(content);
        break;
      case 'publishToWebsite':
        publishToWebsite(content);
        break;
      case 'convertToVideo':
        convertToVideoScript(content);
        break;
      case 'generateArticle':
        generateArticleFromContent(content);
        break;
      case 'copyContent':
        copyToClipboard(content);
        break;
      default:
        showToast('功能开发中...', 'error');
    }
  }

  function saveContentToLibrary(content) {
    // 显示加载中提示
    showToast('正在分析内容...', 'success');
    
    var token = getToken();
    var headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    
    // 使用 AI 提取标题和摘要
    var extractPrompt = '请从以下内容中提取：\n1. 标题（简洁明了，20 字以内）\n2. 摘要（100 字以内的概述）\n\n内容如下：\n' + content.substring(0, 1000);
    
    fetch('/ai/chat', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify({
        messages: [{ role: 'user', content: extractPrompt }],
        stream: false
      })
    })
    .then(function(res) {
      if (!res.ok) throw new Error('AI 提取失败');
      return res.json();
    })
    .then(function(data) {
      var reply = data.reply || '';
      
      // 解析 AI 返回的结果
      var titleMatch = reply.match(/(?:标题)[:：]\s*(.+)/);
      var summaryMatch = reply.match(/(?:摘要)[:：]\s*(.+)/);
      
      var title = titleMatch ? titleMatch[1].trim() : 'AI 生成的内容 - ' + new Date().toLocaleDateString();
      var summary = summaryMatch ? summaryMatch[1].trim() : '';
      
      // 确认后保存
      if (confirm('AI 已提取内容信息：\n\n标题：' + title + '\n摘要：' + (summary ? summary.substring(0, 50) + '...' : '无') + '\n\n是否保存到内容库？')) {
        var body = JSON.stringify({
          title: title,
          summary: summary || '',
          type: contentMode,
          input_type: 'ai_generated',
          content_data: content,
          generated_content: content,
          status: 'draft'
        });
        
        fetch('/contents', {
          method: 'POST',
          headers: headers,
          body: body
        })
        .then(function(res) {
          if (res.ok) {
            return res.json();
          }
          throw new Error('保存失败');
        })
        .then(function(data) {
          showToast('内容已保存到内容库！', 'success');
          if (confirm('是否跳转到内容管理页面查看？')) {
            window.location.href = '/admin/contents';
          }
        })
        .catch(function(err) {
          showToast(err.message, 'error');
        });
      }
    })
    .catch(function(err) {
      // AI 提取失败时使用默认值
      var defaultTitle = 'AI 生成的内容 - ' + new Date().toLocaleDateString();
      
      if (confirm('AI 内容分析失败，使用默认标题：' + defaultTitle + '\n\n是否继续保存？')) {
        var body = JSON.stringify({
          title: defaultTitle,
          summary: '',
          type: contentMode,
          input_type: 'ai_generated',
          content_data: content,
          generated_content: content,
          status: 'draft'
        });
        
        fetch('/contents', {
          method: 'POST',
          headers: headers,
          body: body
        })
        .then(function(res) {
          if (res.ok) {
            return res.json();
          }
          throw new Error('保存失败');
        })
        .then(function(data) {
          showToast('内容已保存到内容库！', 'success');
          if (confirm('是否跳转到内容管理页面查看？')) {
            window.location.href = '/admin/contents';
          }
        })
        .catch(function(err) {
          showToast(err.message, 'error');
        });
      }
    });
  }

  function publishToWebsite(content) {
    // 显示加载中提示
    showToast('正在分析内容并准备发布...', 'success');
    
    // 调用 AI 自动提取标题、摘要和分类
    var token = getToken();
    var headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    
    // 使用 AI 提取文章信息
    var extractPrompt = '请从以下内容中提取：\n1. 文章标题（简洁明了，20 字以内）\n2. 文章摘要（100 字以内的概述）\n3. 文章分类（从以下选择：签证政策、商务指南、市场动态、旅行攻略、文化习俗）\n\n内容如下：\n' + content.substring(0, 1000);
    
    fetch('/ai/chat', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify({
        messages: [{ role: 'user', content: extractPrompt }],
        stream: false
      })
    })
    .then(function(res) {
      if (!res.ok) throw new Error('AI 提取失败');
      return res.json();
    })
    .then(function(data) {
      var reply = data.reply || '';
      
      // 解析 AI 返回的结果
      var titleMatch = reply.match(/(?:文章标题 | 标题)[:：]\s*(.+)/);
      var summaryMatch = reply.match(/(?:文章摘要 | 摘要)[:：]\s*(.+)/);
      var categoryMatch = reply.match(/(?:文章分类 | 分类)[:：]\s*(.+)/);
      
      var title = titleMatch ? titleMatch[1].trim() : 'AI 生成的内容 - ' + new Date().toLocaleDateString();
      var summary = summaryMatch ? summaryMatch[1].trim() : content.substring(0, 200) + '...';
      var category = categoryMatch ? categoryMatch[1].trim() : '签证政策';
      
      // 生成 URL 标识
      var slug = generateSlug(title);
      
      // 确认后发布
      if (confirm('AI 已提取文章信息：\n\n标题：' + title + '\n分类：' + category + '\n摘要：' + summary.substring(0, 50) + '...\n\n是否确认发布到网站？')) {
        var body = JSON.stringify({
          title: title,
          slug: slug,
          summary: summary,
          content: content,
          category: category,
          status: 'published',
          is_ai_generated: true
        });
        
        fetch('/blog-posts', {
          method: 'POST',
          headers: headers,
          body: body
        })
        .then(function(res) {
          if (res.ok) {
            return res.json();
          }
          throw new Error('发布失败');
        })
        .then(function(data) {
          showToast('文章已成功发布到网站！', 'success');
          if (confirm('是否跳转到博客管理页面？')) {
            window.location.href = '/admin/blog-posts';
          }
        })
        .catch(function(err) {
          showToast(err.message, 'error');
        });
      }
    })
    .catch(function(err) {
      // AI 提取失败时使用默认值
      var defaultTitle = 'AI 生成的内容 - ' + new Date().toLocaleDateString();
      var defaultSummary = content.substring(0, 200) + '...';
      var defaultSlug = generateSlug(defaultTitle);
      
      if (confirm('AI 内容分析失败，使用默认信息：\n\n标题：' + defaultTitle + '\n\n是否继续发布？')) {
        var body = JSON.stringify({
          title: defaultTitle,
          slug: defaultSlug,
          summary: defaultSummary,
          content: content,
          category: '签证政策',
          status: 'published',
          is_ai_generated: true
        });
        
        var token = getToken();
        var headers = { 'Content-Type': 'application/json' };
        if (token) headers['Authorization'] = 'Bearer ' + token;
        
        fetch('/blog-posts', {
          method: 'POST',
          headers: headers,
          body: body
        })
        .then(function(res) {
          if (res.ok) {
            return res.json();
          }
          throw new Error('发布失败');
        })
        .then(function(data) {
          showToast('文章已成功发布到网站！', 'success');
          if (confirm('是否跳转到博客管理页面？')) {
            window.location.href = '/admin/blog-posts';
          }
        })
        .catch(function(err) {
          showToast(err.message, 'error');
        });
      }
    });
  }

  function convertToVideoScript(content) {
    var prompt = '请将以下内容转换为视频脚本格式，包含场景描述、旁白、画面建议等：\n\n' + content;
    document.getElementById('user-input').value = prompt;
    contentMode = 'video';
    updateContentTypeUI();
    showToast('已切换到视频脚本模式，请点击发送', 'success');
  }

  function generateArticleFromContent(content) {
    var prompt = '请根据以下视频脚本生成一篇完整的文章：\n\n' + content;
    document.getElementById('user-input').value = prompt;
    contentMode = 'article';
    updateContentTypeUI();
    showToast('已切换到文章模式，请点击发送', 'success');
  }

  function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function() {
        showToast('已复制到剪贴板！', 'success');
      }, function(err) {
        fallbackCopy(text);
      });
    } else {
      fallbackCopy(text);
    }
  }

  function fallbackCopy(text) {
    var textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    document.body.appendChild(textArea);
    textArea.select();
    try {
      document.execCommand('copy');
      showToast('已复制到剪贴板！', 'success');
    } catch (err) {
      showToast('复制失败', 'error');
    }
    document.body.removeChild(textArea);
  }

  function updateContentTypeUI() {
    var ctBtns = document.querySelectorAll('.ct-btn');
    for (var i = 0; i < ctBtns.length; i++) {
      ctBtns[i].classList.remove('active');
      if (ctBtns[i].getAttribute('data-type') === contentMode) {
        ctBtns[i].classList.add('active');
      }
    }
    if (ctLabelEl) {
      ctLabelEl.textContent = typeLabels[contentMode] || contentMode;
    }
  }

  function generateSlug(title) {
    return title.toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
  }

  function showToast(message, type) {
    var existingToast = document.querySelector('.toast');
    if (existingToast) {
      existingToast.remove();
    }
    
    var toast = document.createElement('div');
    toast.className = 'toast' + (type === 'error' ? ' error' : '');
    toast.textContent = message;
    document.body.appendChild(toast);
    
    setTimeout(function() {
      toast.classList.add('show');
    }, 10);
    
    setTimeout(function() {
      toast.classList.remove('show');
      setTimeout(function() {
        toast.remove();
      }, 300);
    }, 3000);
  }
})();
