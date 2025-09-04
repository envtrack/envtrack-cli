package initialize

import "strconv"

func getWebHtml() string {
	html := `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<title>envtrack - commands debug</title>
<style>body{font-family:system-ui,Segoe UI,Roboto,Arial;background:#f7f7f8;padding:16px}#json{background:#fff;padding:12px;border:1px solid #ddd;border-radius:6px;overflow:auto;max-height:80vh}button.toggle{border:none;background:transparent;color:#0366d6;cursor:pointer;padding:0;margin-right:6px}pre{margin:0}</style>
</head>
<body>
<h1>Commands configuration (debug)</h1>
<p>Connected: <span id="status">disconnected</span></p>
<div id="json">Loading...</div>

<!-- Minimal jsontree.js implementation embedded so no external deps are required -->
<script>
// jsontree.js - tiny JSON tree renderer
function renderJsonTree(container, value) {
  container.innerHTML = '';
  const node = renderNode(value);
  container.appendChild(node);
}

function renderNode(value) {
  if (value === null) return textNode('null');
  if (typeof value === 'object') {
    if (Array.isArray(value)) return renderCollection(value, 'array');
    return renderCollection(value, 'object');
  }
  return textNode(value);
}

function textNode(v) {
  const pre = document.createElement('pre');
  pre.textContent = JSON.stringify(v);
  return pre;
}

function renderCollection(coll, kind) {
  const wrapper = document.createElement('div');
  const header = document.createElement('div');
  const toggle = document.createElement('button');
  toggle.className = 'toggle';
  toggle.textContent = kind === 'array' ? '[' + coll.length + ']' : '{' + Object.keys(coll).length + '}';
  header.appendChild(toggle);
  wrapper.appendChild(header);

  const children = document.createElement('div');
  children.style.marginLeft = '16px';
  children.style.display = 'none';

  if (kind === 'array') {
    coll.forEach((it, i) => {
      const row = document.createElement('div');
      const idx = document.createElement('span');
      idx.style.color = '#6a737d';
      idx.textContent = i + ': ';
      row.appendChild(idx);
      row.appendChild(renderNode(it));
      children.appendChild(row);
    });
  } else {
    Object.keys(coll).forEach(k => {
      const row = document.createElement('div');
      const key = document.createElement('span');
      key.style.color = '#22863a';
      key.textContent = k;
      row.appendChild(key);
      row.appendChild(renderNode(coll[k]));
      children.appendChild(row);
    });
  }

  toggle.addEventListener('click', function(){
    if (children.style.display === 'none') {
      children.style.display = '';
      toggle.textContent = (kind === 'array' ? '['+coll.length+']' : '{'+Object.keys(coll).length+'}') + ' ▼';
    } else {
      children.style.display = 'none';
      toggle.textContent = (kind === 'array' ? '['+coll.length+']' : '{'+Object.keys(coll).length+'}');
    }
  });

  wrapper.appendChild(children);
  return wrapper;
}
</script>

<script>
(function(){
  var host = location.hostname;
  var port = location.port || "` + strconv.Itoa(webPort) + `";
  var proto = location.protocol === 'https:' ? 'wss' : 'ws';
  var url = proto + '://' + host + (port ? ':' + port : '') + '/ws';
  var ws = new WebSocket(url);
  var status = document.getElementById('status');
  var out = document.getElementById('json');
  status.textContent = 'connecting';
  ws.onopen = function(){ status.textContent = 'connected'; console.debug('ws open'); };
  // Avoid re-rendering identical payloads which can look like "reloading" in the UI.
  ws.onmessage = function(e){
    console.debug('ws message', e.data && e.data.length ? ('len='+e.data.length) : e.data);
    // keep last payload and skip if identical
    if (window.__lastDebugJson!= undefined) return;
    window.__lastDebugJson = e.data;
    try {
      var obj = JSON.parse(e.data);
      if (obj.parsedCommands) {
			let oldCmds = obj.parsedCommands;
			obj.parsedCommands = {};
			// we transform the keys from cmd["a.b.c"] to cmd[a][b][c] for better readability
			Object.keys(oldCmds).forEach(function(k) {
				var parts = k.split(':');
				var cur = obj.parsedCommands;
				for (var i = 0; i < parts.length; i++) {
					var p = parts[i];
					if (i === parts.length - 1) {
						// leaf -> set value
						cur[p] = oldCmds[k];
					} else {
						// ensure intermediate object exists
						if (cur[p] === undefined || typeof cur[p] !== 'object') {
							cur[p] = {};
						}
						cur = cur[p];
					}
				}
			});
			
		}
      renderJsonTree(out, obj);
    } catch (err) {
      out.textContent = e.data;
    }
  };
  ws.onerror = function(){ status.textContent = 'error'; };
  ws.onclose = function(){ if(status.textContent !== 'error') status.textContent = 'closed'; };
})();
</script>
</body>
</html>`
	return html
}
