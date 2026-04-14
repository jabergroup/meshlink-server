package api

import "net/http"

const dashboardHTML = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>MeshLink - Dashboard</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    background: #0a0a0f;
    color: #e0e0e0;
    min-height: 100vh;
  }
  .header {
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
    padding: 24px 32px;
    border-bottom: 1px solid #2a2a4a;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .logo { font-size: 24px; font-weight: 700; color: #00d4aa; letter-spacing: -0.5px; }
  .logo span { color: #667; font-weight: 400; }
  .tagline { color: #556; font-size: 13px; margin-top: 2px; }
  .header-actions { display: flex; align-items: center; gap: 12px; }
  .status-badge {
    background: #0d2818; color: #00d4aa; padding: 6px 16px;
    border-radius: 20px; font-size: 13px; border: 1px solid #00d4aa33;
  }
  .btn-download {
    display: inline-flex; align-items: center; gap: 8px;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
    color: #00d4aa; padding: 8px 18px; border-radius: 8px;
    font-size: 13px; font-weight: 600; border: 1px solid #00d4aa44;
    cursor: pointer; text-decoration: none; transition: all 0.2s;
  }
  .btn-download:hover { background: #00d4aa18; box-shadow: 0 2px 12px #00d4aa33; transform: translateY(-1px); }
  .btn-download svg { width: 16px; height: 16px; }

  .container { max-width: 900px; margin: 0 auto; padding: 32px 24px; }

  .card {
    background: #12121f; border: 1px solid #1e1e35;
    border-radius: 12px; padding: 28px; margin-bottom: 20px;
  }
  .card h2 {
    font-size: 18px; color: #00d4aa; margin-bottom: 20px;
    display: flex; align-items: center; gap: 10px;
  }

  .steps { display: grid; gap: 16px; }
  .step {
    display: flex; gap: 16px; padding: 20px;
    background: #0d0d18; border-radius: 10px; border: 1px solid #1a1a30;
    transition: border-color 0.3s;
  }
  .step:hover { border-color: #00d4aa44; }
  .step-num {
    width: 36px; height: 36px; border-radius: 50%; flex-shrink: 0;
    background: #00d4aa22; color: #00d4aa; display: flex;
    align-items: center; justify-content: center;
    font-weight: 700; font-size: 15px;
  }
  .step-content h3 { font-size: 15px; font-weight: 600; margin-bottom: 6px; }
  .step-content p { color: #778; font-size: 13px; line-height: 1.5; }
  .step-content code {
    background: #1a1a30; padding: 2px 8px; border-radius: 4px;
    font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; color: #00d4aa;
  }

  .pair-section { text-align: center; padding: 40px 20px; }
  .pair-code {
    font-size: 52px; font-weight: 800; color: #fff; letter-spacing: 8px;
    font-family: 'SF Mono', 'Fira Code', monospace; margin: 24px 0;
    text-shadow: 0 0 40px #00d4aa44;
  }
  .pair-code .dash { color: #334; margin: 0 4px; }
  .pair-timer { color: #667; font-size: 14px; margin-top: 8px; }
  .pair-hint { color: #445; font-size: 13px; margin-top: 16px; line-height: 1.6; }

  .btn {
    display: inline-block; padding: 12px 32px; border-radius: 8px;
    border: none; cursor: pointer; font-size: 15px; font-weight: 600; transition: all 0.2s;
  }
  .btn-primary { background: linear-gradient(135deg, #00d4aa 0%, #00a88a 100%); color: #000; }
  .btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 20px #00d4aa44; }
  .btn-outline { background: transparent; color: #00d4aa; border: 1px solid #00d4aa44; }
  .btn-outline:hover { background: #00d4aa11; }
  .btn-group { display: flex; gap: 12px; justify-content: center; margin-top: 24px; }
  .btn-sm { padding: 8px 20px; font-size: 13px; }
  .btn-back { background: transparent; color: #667; border: 1px solid #2a2a4a; }
  .btn-back:hover { border-color: #445; color: #aaa; }

  .device-list { display: grid; gap: 12px; }
  .device-row {
    display: flex; align-items: center; justify-content: space-between;
    padding: 16px 20px; background: #0d0d18; border-radius: 8px; border: 1px solid #1a1a30;
  }
  .device-info { display: flex; align-items: center; gap: 12px; }
  .device-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
  .dot-green { background: #00d4aa; box-shadow: 0 0 8px #00d4aa88; }
  .dot-yellow { background: #f0c040; box-shadow: 0 0 8px #f0c04088; }
  .dot-red { background: #e04050; box-shadow: 0 0 8px #e0405088; }
  .dot-gray { background: #444; }
  .device-name { font-weight: 600; font-size: 15px; }
  .device-meta { color: #556; font-size: 13px; margin-top: 2px; }
  .device-role { font-size: 12px; padding: 3px 10px; border-radius: 12px; background: #1a1a30; color: #889; }

  .access-box {
    background: #0d0d18; border-radius: 8px; padding: 16px 20px; margin-top: 12px;
    font-family: 'SF Mono', 'Fira Code', monospace; font-size: 14px;
    color: #00d4aa; border: 1px solid #1a1a30; user-select: all;
    white-space: pre-line;
  }

  .empty-state { text-align: center; padding: 40px 20px; color: #445; }
  .empty-state p { margin-top: 8px; font-size: 14px; }

  #join-input {
    background: #0d0d18; border: 1px solid #2a2a4a; color: #fff;
    padding: 12px 20px; border-radius: 8px; font-size: 20px;
    text-align: center; letter-spacing: 4px; width: 200px;
    font-family: 'SF Mono', monospace; outline: none;
  }
  #join-input:focus { border-color: #00d4aa; }
  #join-input::placeholder { color: #334; letter-spacing: 2px; }

  .hidden { display: none; }
  @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.5} }
  .pulsing { animation: pulse 2s infinite; }

  .toast {
    position: fixed; top: 24px; right: 24px; background: #1a1a2e;
    border: 1px solid #00d4aa44; border-radius: 10px; padding: 14px 20px;
    color: #00d4aa; font-size: 14px; z-index: 1000; box-shadow: 0 8px 32px #0008;
    transform: translateX(120%); transition: transform 0.3s ease; max-width: 400px;
  }
  .toast.show { transform: translateX(0); }
</style>
</head>
<body>

<div class="toast" id="toast"></div>

<div class="header">
  <div>
    <div class="logo">MeshLink <span>Dashboard</span></div>
    <div class="tagline">Direct. Private. No Third Party.</div>
  </div>
  <div class="header-actions">
    <a href="/download/agent" class="btn-download">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
        <polyline points="7 10 12 15 17 10"/>
        <line x1="12" y1="15" x2="12" y2="3"/>
      </svg>
      Download Agent
    </a>
    <div class="status-badge" id="server-status">Server Online</div>
  </div>
</div>

<div class="container">

  <div class="card" id="action-card">
    <div class="pair-section">
      <h2 style="justify-content:center; font-size:22px; margin-bottom:8px;">Create a P2P Tunnel</h2>
      <p style="color:#556; margin-bottom:24px;">Connect two devices directly with end-to-end encryption</p>
      <div class="btn-group">
        <button class="btn btn-primary" onclick="createPair()">Create New Tunnel</button>
        <button class="btn btn-outline" onclick="showJoinInput()">Join with Code</button>
      </div>
      <div id="join-section" class="hidden" style="margin-top:24px;">
        <input id="join-input" type="text" maxlength="7" placeholder="000-000"
               oninput="formatCode(this)" onkeypress="if(event.key==='Enter')joinPair()">
        <div style="margin-top:12px;">
          <button class="btn btn-primary btn-sm" onclick="joinPair()">Connect</button>
          <button class="btn btn-back btn-sm" onclick="hideJoinInput()" style="margin-left:8px;">Cancel</button>
        </div>
      </div>
    </div>
  </div>

  <div class="card hidden" id="code-card">
    <div class="pair-section">
      <p style="color:#889;">Share this code with the other device</p>
      <div class="pair-code" id="pair-code">---<span class="dash">-</span>---</div>
      <div class="pair-timer pulsing" id="pair-timer">Waiting for peer to connect...</div>
      <div class="btn-group" style="margin-top:20px;">
        <button class="btn btn-outline btn-sm" onclick="copyCode()">Copy Code</button>
        <button class="btn btn-back btn-sm" onclick="cancelPair()">Cancel</button>
      </div>
      <div class="pair-hint" id="pair-hint"></div>
    </div>
  </div>

  <div class="card" id="guide-card">
    <h2>How It Works</h2>
    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-content">
          <h3>Download the Agent</h3>
          <p>Click <strong>Download Agent</strong> above and run <code>MeshLink-Setup.exe</code> on both devices. The installer will set up everything automatically.</p>
        </div>
      </div>
      <div class="step">
        <div class="step-num">2</div>
        <div class="step-content">
          <h3>Run the Agent</h3>
          <p>Double-click <code>meshlink-agent.exe</code>. On first run it will ask for the server URL. Enter: <code id="server-url-display">loading...</code></p>
        </div>
      </div>
      <div class="step">
        <div class="step-num">3</div>
        <div class="step-content">
          <h3>Create a Tunnel (Device A)</h3>
          <p>Right-click the MeshLink icon in the system tray and select <strong>Create Tunnel</strong>. A 6-digit code will appear. You can also create one from this dashboard.</p>
        </div>
      </div>
      <div class="step">
        <div class="step-num">4</div>
        <div class="step-content">
          <h3>Join from Device B</h3>
          <p>On the other device, right-click the MeshLink tray icon, select <strong>Join Tunnel</strong>, and enter the 6-digit code. The devices will connect directly via P2P.</p>
        </div>
      </div>
    </div>
  </div>

  <div class="card">
    <h2>Active Tunnels</h2>
    <div id="sessions-list">
      <div class="empty-state" id="empty-state">
        <p>No active tunnels</p>
        <p style="font-size:13px; color:#334;">Create a tunnel to get started</p>
      </div>
    </div>
  </div>

</div>

<script>
const API = '';
let currentCode = '';
let pollTimer = null;

function showToast(msg, duration) {
  var t = document.getElementById('toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(function(){ t.classList.remove('show'); }, duration || 3000);
}

document.getElementById('server-url-display').textContent = location.origin;

async function createPair() {
  try {
    var resp = await fetch(API + '/api/pair/create', { method: 'POST' });
    var data = await resp.json();
    currentCode = data.code;
    document.getElementById('pair-code').innerHTML =
      data.code.slice(0,3) + '<span class="dash">-</span>' + data.code.slice(3);
    document.getElementById('action-card').classList.add('hidden');
    document.getElementById('guide-card').classList.add('hidden');
    document.getElementById('code-card').classList.remove('hidden');
    document.getElementById('pair-timer').textContent = 'Waiting for peer to connect...';
    document.getElementById('pair-timer').style.color = '#667';
    document.getElementById('pair-timer').classList.add('pulsing');
    document.getElementById('pair-hint').innerHTML =
      'On the other device: open the MeshLink Agent tray icon<br>and click <strong>Join Tunnel</strong>, then enter this code.<br><br>' +
      'Or run: <code>meshlink-agent.exe -server ' + location.origin + ' -code ' + data.code + '</code>';
    pollForPeer(data.code);
  } catch(e) {
    showToast('Failed to connect to server: ' + e.message, 5000);
  }
}

function showJoinInput() {
  document.getElementById('join-section').classList.remove('hidden');
  document.getElementById('join-input').focus();
}

function hideJoinInput() {
  document.getElementById('join-section').classList.add('hidden');
  document.getElementById('join-input').value = '';
}

function formatCode(input) {
  var v = input.value.replace(/\D/g, '');
  if (v.length > 3) v = v.slice(0,3) + '-' + v.slice(3,6);
  input.value = v;
}

async function joinPair() {
  var code = document.getElementById('join-input').value.replace(/-/g,'');
  if (code.length !== 6) { showToast('Please enter a valid 6-digit code'); return; }

  try {
    var resp = await fetch(API + '/api/pair/status?code=' + code);
    if (!resp.ok) { showToast('Session not found or expired. Check the code.', 4000); return; }
    var data = await resp.json();

    if (data.status === 'connected') {
      showToast('This tunnel is already connected!', 3000);
    } else {
      showToast('Session found! Use the Agent to complete the connection.', 4000);
    }

    currentCode = code;
    document.getElementById('pair-code').innerHTML =
      code.slice(0,3) + '<span class="dash">-</span>' + code.slice(3);
    document.getElementById('action-card').classList.add('hidden');
    document.getElementById('guide-card').classList.add('hidden');
    document.getElementById('code-card').classList.remove('hidden');
    document.getElementById('pair-timer').textContent = 'Status: ' + data.status;
    document.getElementById('pair-hint').innerHTML =
      'Run on the joining device:<br><code>meshlink-agent.exe -server ' + location.origin + ' -code ' + code + '</code>';
    pollForPeer(code);

  } catch(e) {
    showToast('Could not check session: ' + e.message, 4000);
  }
}

function cancelPair() {
  if (pollTimer) { clearTimeout(pollTimer); pollTimer = null; }
  currentCode = '';
  document.getElementById('code-card').classList.add('hidden');
  document.getElementById('action-card').classList.remove('hidden');
  document.getElementById('guide-card').classList.remove('hidden');
}

function copyCode() {
  var code = currentCode.slice(0,3) + '-' + currentCode.slice(3);
  navigator.clipboard.writeText(code);
  showToast('Code copied: ' + code);
}

async function pollForPeer(code) {
  var check = async function() {
    try {
      var resp = await fetch(API + '/api/pair/status?code=' + code);
      var data = await resp.json();
      var timerEl = document.getElementById('pair-timer');

      if (data.status === 'connected') {
        timerEl.textContent = 'Tunnel connected!';
        timerEl.classList.remove('pulsing');
        timerEl.style.color = '#00d4aa';
        document.getElementById('pair-hint').innerHTML = 'Both devices are now connected via P2P tunnel.';
        refreshSessions();
        return;
      } else if (data.status === 'paired') {
        timerEl.textContent = 'Peer joined! Establishing tunnel...';
        timerEl.style.color = '#f0c040';
      } else {
        timerEl.textContent = 'Waiting for peer to connect...';
      }
    } catch(e) {}
    pollTimer = setTimeout(check, 2000);
  };
  check();
}

async function refreshSessions() {
  try {
    var resp = await fetch(API + '/api/sessions');
    var sessions = await resp.json();
    var list = document.getElementById('sessions-list');
    var empty = document.getElementById('empty-state');

    if (!sessions || sessions.length === 0) {
      empty.classList.remove('hidden');
      return;
    }
    empty.classList.add('hidden');

    var html = '<div class="device-list">';
    for (var i = 0; i < sessions.length; i++) {
      var s = sessions[i];
      var statusDot = s.status === 'connected' ? 'dot-green' :
                      s.status === 'paired' ? 'dot-yellow' :
                      s.status === 'waiting' ? 'dot-gray' : 'dot-red';
      html += '<div class="device-row">';
      html += '  <div class="device-info">';
      html += '    <div class="device-dot ' + statusDot + '"></div>';
      html += '    <div>';
      html += '      <div class="device-name">Tunnel ' + s.code.slice(0,3) + '-' + s.code.slice(3) + '</div>';
      html += '      <div class="device-meta">';
      if (s.server) html += s.server.name;
      if (s.server && s.client) html += ' &#8596; ';
      if (s.client) html += s.client.name;
      if (!s.server && !s.client) html += 'Waiting for devices...';
      html += '    </div></div>';
      html += '  </div>';
      html += '  <div class="device-role">' + s.status + '</div>';
      html += '</div>';

      if (s.status === 'connected') {
        html += '<div class="access-box">';
        var services = [];
        if (s.server && s.server.services) services = s.server.services;
        if (services.length === 0) services = ['ssh:22'];
        for (var j = 0; j < services.length; j++) {
          var parts = services[j].split(':');
          var name = parts[0];
          var port = parseInt(parts[1]) || 0;
          if (name === 'ssh') html += 'SSH: ssh -p ' + (port+10000) + ' user@localhost\n';
          else if (name === 'rdp') html += 'RDP: localhost:' + (port+10000) + '\n';
          else html += name + ': localhost:' + (port+10000) + '\n';
        }
        html += '</div>';
      }
    }
    html += '</div>';
    list.innerHTML = html;
  } catch(e) {}
}

setInterval(refreshSessions, 3000);
refreshSessions();
</script>

</body>
</html>`

// ServeDashboard serves the embedded HTML dashboard
func ServeDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}
