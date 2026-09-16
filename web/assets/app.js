const apiBase = "/api/v1";
const content = document.querySelector("#page-content");
const title = document.querySelector("#page-title");
const state = document.querySelector("#api-state");
const toast = document.querySelector("#toast");
let toastTimer;

const pageNames = { overview: "Overview", nodes: "Node management", sessions: "Signing sessions", transactions: "Transactions", audit: "Audit log" };

async function api(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, { headers: { "Content-Type": "application/json", ...(options.headers || {}) }, ...options });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok || payload.code !== 0) throw new Error(payload.message || "Request failed");
  return payload.data;
}
function esc(value = "") { return String(value).replace(/[&<>'"]/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[char]); }
function formatTime(value) { return value ? new Date(value).toLocaleString() : "-"; }
function status(value) { return `<span class="status ${esc(value)}">${esc(value)}</span>`; }
function empty(message) { return `<div class="empty">${esc(message)}</div>`; }
function showToast(message) { toast.textContent = message; toast.classList.add("show"); clearTimeout(toastTimer); toastTimer = setTimeout(() => toast.classList.remove("show"), 3200); }
function setApiState(ok) { state.textContent = ok ? "API connected" : "API unavailable"; state.className = `api-state ${ok ? "online" : "offline"}`; }

async function overview() {
  const [metrics, walletData, sessionData] = await Promise.all([api("/system/metrics"), api("/wallets"), api("/sign-sessions")]);
  const wallet = walletData.list[0];
  const sessions = sessionData.list.slice(0, 5);
  return `<div class="summary-grid">
    <article class="stat"><span class="stat-label">Online signing nodes</span><span class="stat-value">${metrics.online_nodes}</span><span class="stat-hint">All committee members reachable</span></article>
    <article class="stat"><span class="stat-label">Signing sessions</span><span class="stat-value">${metrics.active_sessions}</span><span class="stat-hint">Current service instance</span></article>
    <article class="stat"><span class="stat-label">Completed signatures</span><span class="stat-value">${metrics.completed_sessions}</span><span class="stat-hint">2-of-3 approvals accepted</span></article>
    <article class="stat"><span class="stat-label">Threshold policy</span><span class="stat-value">2 / 3</span><span class="stat-hint">tss-lib parameter threshold = 1</span></article>
  </div><div class="section-grid"><section class="panel"><div class="panel-head"><h2>Recent signing sessions</h2><a class="button secondary small" href="#sessions">Open sessions</a></div><div class="table-wrap"><table><thead><tr><th>Session</th><th>Digest</th><th>Approvals</th><th>Status</th><th>Created</th></tr></thead><tbody>
  ${sessions.length ? sessions.map((item) => `<tr><td class="code">${esc(item.id)}</td><td class="code">${esc(item.digest)}</td><td>${item.approvers.length}/2</td><td>${status(item.status)}</td><td>${formatTime(item.created_at)}</td></tr>`).join("") : `<tr><td colspan="5">${empty("No signing session has been created.")}</td></tr>`}
  </tbody></table></div></section><div><section class="panel"><div class="panel-head"><h2>Wallet profile</h2></div><div class="panel-body"><div class="key-value"><b>Wallet ID</b><span class="code">${esc(wallet.id)}</span></div><div class="key-value"><b>Address</b><span class="code">${esc(wallet.address)}</span></div><div class="key-value"><b>Network</b><span>${esc(wallet.network)}</span></div><div class="key-value"><b>Status</b><span>${status(wallet.status)}</span></div></div></section><section class="panel"><div class="panel-head"><h2>Protocol status</h2></div><div class="panel-body"><div class="callout">The coordinator manages session metadata and protocol routing. Private key shares are not exposed by this API or stored in the coordinator state.</div></div></section></div></div>`;
}

async function nodes() {
  const data = await api("/participants");
  return `<section class="panel"><div class="panel-head"><h2>Committee nodes</h2><span class="muted">3 nodes / 2 approvals required</span></div><div class="table-wrap"><table><thead><tr><th>Node ID</th><th>Availability</th><th>Endpoint</th><th>Key epoch</th><th>Last heartbeat</th></tr></thead><tbody>${data.list.map((item) => `<tr><td class="code">${esc(item.id)}</td><td>${status(item.status)}</td><td class="code">${esc(item.endpoint)}</td><td>${item.key_epoch}</td><td>${formatTime(item.last_heartbeat)}</td></tr>`).join("")}</tbody></table></div></section><section class="panel"><div class="panel-head"><h2>Committee policy</h2></div><div class="panel-body"><div class="callout">The participant endpoint includes a heartbeat route and a resharing request route. In this demo they manage metadata only; the existing tss-lib integration test verifies the real protocol call path separately.</div></div></section>`;
}

async function sessions() {
  const data = await api("/sign-sessions");
  return `<section class="panel"><div class="panel-head"><h2>Create signing session</h2><span class="muted">Create a session for the test wallet and message digest.</span></div><div class="panel-body"><form id="create-session" class="form-row"><label>Transaction digest<input required name="digest" maxlength="160" placeholder="e.g. 4d7f..." /></label><label>Wallet ID<input readonly value="wallet-demo-001" /></label><button class="button" type="submit">Create session</button></form></div></section><section class="panel"><div class="panel-head"><h2>Signing session queue</h2><span class="muted">Two distinct participants are required.</span></div><div class="table-wrap"><table><thead><tr><th>Session ID</th><th>Digest</th><th>Approvers</th><th>Status</th><th>Created</th><th></th></tr></thead><tbody>${data.list.length ? data.list.map((item) => `<tr><td class="code">${esc(item.id)}</td><td class="code">${esc(item.digest)}</td><td>${item.approvers.length ? esc(item.approvers.join(", ")) : "No approvals"}</td><td>${status(item.status)} <span class="muted">${item.approvers.length}/2</span></td><td>${formatTime(item.created_at)}</td><td>${["pending", "signing"].includes(item.status) ? `<button class="button secondary small approve" data-id="${esc(item.id)}">Approve</button>` : ""}</td></tr>`).join("") : `<tr><td colspan="6">${empty("No signing session has been created.")}</td></tr>`}</tbody></table></div></section>`;
}

async function transactions() {
  const data = await api("/transactions");
  return `<section class="panel"><div class="panel-head"><h2>Transactions</h2><span class="muted">Transfer records bound to threshold signing sessions.</span></div><div class="table-wrap"><table><thead><tr><th>Transaction ID</th><th>Recipient</th><th>Amount</th><th>Session</th><th>Status</th><th>Created</th></tr></thead><tbody>${data.list.length ? data.list.map((item) => `<tr><td class="code">${esc(item.id)}</td><td class="code">${esc(item.to_address)}</td><td>${esc(item.amount)}</td><td class="code">${esc(item.session_id || "-")}</td><td>${status(item.status)}</td><td>${formatTime(item.created_at)}</td></tr>`).join("") : `<tr><td colspan="6">${empty("No transaction has been created yet.")}</td></tr>`}</tbody></table></div></section>`;
}

async function audit() {
  const data = await api("/audit-logs");
  return `<section class="panel"><div class="panel-head"><h2>Operational audit trail</h2><span class="muted">Service-layer audit events.</span></div><div class="panel-body">${data.list.length ? `<ol class="timeline">${data.list.map((item) => `<li><strong>${esc(item.detail)}</strong><span class="muted"> by ${esc(item.actor)}${item.session_id ? ` for <span class="code">${esc(item.session_id)}</span>` : ""}</span><time>${formatTime(item.created_at)} | ${esc(item.action)} | ${esc(item.result)}</time></li>`).join("")}</ol>` : empty("No operation has been recorded.")}</div></section>`;
}

const renders = { overview, nodes, sessions, transactions, audit };
async function render() {
  const chosen = renders[location.hash.slice(1)] ? location.hash.slice(1) : "overview";
  title.textContent = pageNames[chosen];
  document.querySelectorAll(".nav a").forEach((link) => link.classList.toggle("active", link.dataset.page === chosen));
  content.innerHTML = `<div class="panel"><div class="empty">Loading ${pageNames[chosen].toLowerCase()}...</div></div>`;
  try { content.innerHTML = await renders[chosen](); setApiState(true); bindPageActions(chosen); } catch (error) { setApiState(false); content.innerHTML = `<section class="panel"><div class="panel-body"><div class="callout">Could not load data: ${esc(error.message)}. Start the service with <span class="code">go run . httpServer</span>.</div></div></section>`; }
}

function bindPageActions(page) {
  if (page !== "sessions") return;
  document.querySelector("#create-session").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try { await api("/sign-sessions", { method: "POST", body: JSON.stringify({ wallet_id: "wallet-demo-001", digest: form.get("digest") }) }); showToast("Signing session created."); render(); } catch (error) { showToast(error.message); }
  });
  document.querySelectorAll(".approve").forEach((button) => button.addEventListener("click", async () => {
    const node = window.prompt("Approve as node-1, node-2, or node-3", "node-1");
    if (!node) return;
    try { const result = await api(`/sign-sessions/${encodeURIComponent(button.dataset.id)}/approve`, { method: "POST", body: JSON.stringify({ node_id: node }) }); showToast(result.status === "success" ? "Threshold reached. Signing session completed." : "Node approval recorded."); render(); } catch (error) { showToast(error.message); }
  }));
}

document.querySelector("#refresh-button").addEventListener("click", render);
window.addEventListener("hashchange", render);
render();
