// ── State ────────────────────────────────────────────────────────────────────
let prevRequests = {};   // { name: count }
let maxRequests = 1;
let logEntryCount = 0;

// ── Helpers ───────────────────────────────────────────────────────────────────
function fmtUptime(secs) {
    secs = Math.floor(secs);
    if (secs < 60) return `${secs}s`;
    if (secs < 3600) return `${Math.floor(secs / 60)}m ${secs % 60}s`;
    return `${Math.floor(secs / 3600)}h ${Math.floor((secs % 3600) / 60)}m`;
}

function now() {
    return new Date().toLocaleTimeString('en-US', { hour12: false });
}

function addLog(serverName, delta) {
    const feed = document.getElementById('log-feed');
    // Remove placeholder
    if (feed.querySelector('.log-text')?.textContent.includes('waiting')) {
        feed.innerHTML = '';
        logEntryCount = 0;
    }
    logEntryCount++;
    document.getElementById('log-count').textContent = `${logEntryCount} entries`;

    feed.insertAdjacentHTML('afterbegin', `
    <div class="log-entry">
      <span class="log-time">${now()}</span>
      <span class="log-prompt">$</span>
      <span class="log-server">${serverName}</span>
      <span class="log-arrow">←</span>
      <span class="log-text">routed request</span>
      <span class="log-ok">[+${delta} OK]</span>
    </div>`);
    // Keep only 40 entries
    while (feed.children.length > 40) feed.lastChild.remove();
}

// ── Render ────────────────────────────────────────────────────────────────────
function renderServers(data) {
    const servers = data.servers;
    const grid = document.getElementById('server-grid');
    const flow = document.getElementById('flow-track');
    maxRequests = Math.max(1, ...servers.map(s => s.requests));

    // Find the last-routed server (highest individual count change)
    let latestServer = null;
    let maxDelta = 0;
    servers.forEach(s => {
        const prev = prevRequests[s.name] || 0;
        const delta = s.requests - prev;
        if (delta > maxDelta) { maxDelta = delta; latestServer = s.name; }
    });

    // ── Server Cards ──────────────────────────────────────────────────────────
    if (grid.children.length !== servers.length) {
        grid.innerHTML = servers.map((s, i) => `
        <div class="server-card" id="card-${i}">
          <div class="card-header">
            <span class="server-name">${s.name}</span>
            <span class="health-badge online" id="dot-${i}">ONLINE</span>
          </div>
          <div class="card-url">${s.url}</div>
          <div class="card-stat-label">Requests Served</div>
          <div class="card-requests" id="req-${i}">0</div>
          <div class="bar-wrap"><div class="bar-fill" id="bar-${i}"></div></div>
        </div>`).join('');
    }

    // Expose active server index for globe animation
    window.globeActiveIdx = (latestServer && maxDelta > 0)
        ? servers.findIndex(s => s.name === latestServer)
        : -1;

    servers.forEach((s, i) => {
        const card = document.getElementById(`card-${i}`);
        const badge = document.getElementById(`dot-${i}`);
        const reqEl = document.getElementById(`req-${i}`);
        const bar = document.getElementById(`bar-${i}`);

        const isActive = s.name === latestServer && maxDelta > 0;
        const isOffline = !s.healthy;

        // Card class
        card.className = `server-card${isActive ? ' active' : ''}${isOffline ? ' offline' : ''}`;

        // Health badge
        if (isOffline) {
            badge.textContent = 'OFFLINE';
            badge.className = 'health-badge offline';
        } else if (isActive) {
            badge.textContent = 'ACTIVE';
            badge.className = 'health-badge active';
        } else {
            badge.textContent = 'ONLINE';
            badge.className = 'health-badge online';
        }

        reqEl.textContent = s.requests;
        reqEl.className = `card-requests${isActive ? ' counting' : ''}`;
        bar.style.width = `${Math.round((s.requests / maxRequests) * 100)}%`;

        // Log
        if (maxDelta > 0 && isActive) {
            addLog(s.name, maxDelta);
        }
    });

    // ── Flow Track ────────────────────────────────────────────────────────────
    flow.innerHTML = [
        `<div class="flow-node">
           <div class="flow-circle lb-node">LB</div>
           <div class="flow-label">Balancer</div>
         </div>`,
        ...servers.map((s, i) => {
            const isActive = s.name === latestServer && maxDelta > 0;
            const isOffline = !s.healthy;
            return `
            <div class="flow-arrow">→</div>
            <div class="flow-node">
              <div class="flow-circle${isActive ? ' active' : ''}${isOffline ? ' offline' : ''}">${i + 1}</div>
              <div class="flow-label">${s.name}</div>
            </div>`;
        })
    ].join('');

    servers.forEach(s => { prevRequests[s.name] = s.requests; });
}

function renderStats(data) {
    document.getElementById('total-requests').textContent = data.total_requests.toLocaleString();
    document.getElementById('uptime').textContent = fmtUptime(data.uptime_seconds);
    const healthyCount = data.servers.filter(s => s.healthy).length;
    document.getElementById('active-count').textContent = `${healthyCount} / ${data.servers.length}`;
}

// ── Fetch Loop ─────────────────────────────────────────────────────────────────
async function fetchMetrics() {
    try {
        const res = await fetch('/metrics');
        const data = await res.json();
        renderStats(data);
        renderServers(data);
    } catch (e) {
        console.warn('[RR-BALANCER] metrics fetch failed, retrying…', e);
    }
}

fetchMetrics();
setInterval(fetchMetrics, 2000);
