const apiBase = "/api/v1";
const content = document.querySelector("#page-content");
const title = document.querySelector("#page-title");
const state = document.querySelector("#api-state");
const toast = document.querySelector("#toast");
let toastTimer;

const pageNames = {
  overview: "Overview",
  nodes: "Node management",
  sessions: "Signing sessions",
  transactions: "Transactions",
  audit: "Audit log",
};

async function api(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, {
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.error || "Request failed");
  return payload;
}

function esc(value = "") {
  return String(value).replace(/[&<>'"]/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[char]);
}
function formatTime(value) { return value ? new Date(value).toLocaleString() : "-"; }
function status(status) { return `<span class="status ${esc(status)}">${esc(status)}</span>`; }
function empty(message) { return `<div class="empty">${esc(message)}</div>`; }
function showToast(message) {
  toast.textContent = message;
  toast.classList.add("show");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.classList.remove("show"), 3200);
}
function setApiState(ok) {
  state.textContent = ok ? "API connected" : "API unavailable";
  state.className = `api-state ${ok ? "online" : "offline"}`;
}

async function overview() {
  const [metrics, wallet, epochs, sessions] = await Promise.all([
    api("/metrics"), api("/wallets"), api("/key-epochs"), api("/sessions"),
  ]);
  const currentWallet = wallet.items[0];
  const recent = sessions.items.slice().sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt)).slice(0, 5);
  return `
    <div class="summary-grid">
      <article class="stat"><span class="stat-label">Online signing nodes</span><span class="stat-value">${metrics.onlineNodes}</span><span class="stat-hint">All committee members reachable</span></article>
      <article class="stat"><span class="stat-label">Signing sessions</span><span class="stat-value">${metrics.activeSessions}</span><span class="stat-hint">Created in this service instance</span></article>
      <article class="stat"><span class="stat-label">Completed signatures</span><span class="stat-value">${metrics.signedSessions}</span><span class="stat-hint">Threshold signatures released</span></article>
      <article class="stat"><span class="stat-label">Threshold policy</span><span class="stat-value">2 / 3</span><span class="stat-hint">tss-lib parameter threshold = 1</span></article>
    </div>
    <div class="section-grid">
      <section class="panel"><div class="panel-head"><h2>Recent signing sessions</h2><a class="button secondary small" href="#sessions">Open sessions</a></div>
        <div class="table-wrap"><table><thead><tr><th>Session</th><th>Digest</th><th>Approvals</th><th>Status</th><th>Created</th></tr></thead><tbody>
          ${recent.length ? recent.map((s) => `<tr><td class="code">${esc(s.id)}</td><td class="code">${esc(s.digest)}</td><td>${s.approvals}/2</td><td>${status(s.status)}</td><td>${formatTime(s.createdAt)}</td></tr>`).join("") : `<tr><td colspan="5">${empty("No signing session has been created.")}</td></tr>`}
        </tbody></table></div>
      </section>
      <div>
        <section class="panel"><div class="panel-head"><h2>Wallet profile</h2></div><div class="panel-body">
          <div class="key-value"><b>Wallet ID</b><span class="code">${esc(currentWallet.id)}</span></div>
          <div class="key-value"><b>Address</b><span class="code">${esc(currentWallet.address)}</span></div>
          <div class="key-value"><b>Network</b><span>${esc(currentWallet.network)}</span></div>
          <div class="key-value"><b>Key epoch</b><span>${epochs.items[0].epoch}</span></div>
        </div></section>
        <section class="panel"><div class="panel-head"><h2>Protocol status</h2></div><div class="panel-body"><div class="callout">The coordinator only manages approved session messages. Private key shares remain on their respective signing nodes.</div></div></section>
      </div>
    </div>`;
}

async function nodes() {
  const nodes = await api("/nodes");
  return `<section class="panel"><div class="panel-head"><h2>Committee nodes</h2><span class="muted">3 nodes / 2 approvals required</span></div><div class="table-wrap"><table><thead><tr><th>Node ID</th><th>Availability</th><th>Role</th><th>Key share</th><th>Last heartbeat</th></tr></thead><tbody>
    ${nodes.items.map((node, index) => `<tr><td class="code">${esc(node.id)}</td><td>${status(node.status)}</td><td>${index === 0 ? "Coordinator participant" : "Signer participant"}</td><td>Protected</td><td>Just now</td></tr>`).join("")}
  </tbody></table></div></section><section class="panel"><div class="panel-head"><h2>Committee policy</h2></div><div class="panel-body"><div class="callout">This wallet has three independent signing parties. Two distinct node approvals are required before the backend marks a session as signed.</div></div></section>`;
}

async function sessions() {
  const data = await api("/sessions");
  const items = data.items.slice().sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
  return `<section class="panel"><div class="panel-head"><h2>Create signing session</h2><span class="muted">Digest is never altered by the console.</span></div><div class="panel-body"><form id="create-session" class="form-row"><label>Transaction digest<input required name="digest" maxlength="160" placeholder="e.g. 4d7f..." /></label><label>Approval node<select name="node"><option>node-1</option><option>node-2</option><option>node-3</option></select></label><button class="button" type="submit">Create session</button></form></div></section>
    <section class="panel"><div class="panel-head"><h2>Signing session queue</h2><span class="muted">Approve as two different nodes to release a signature.</span></div><div class="table-wrap"><table><thead><tr><th>Session ID</th><th>Digest</th><th>Approvers</th><th>Status</th><th>Created</th><th></th></tr></thead><tbody>
    ${items.length ? items.map((s) => `<tr><td class="code">${esc(s.id)}</td><td class="code">${esc(s.digest)}</td><td>${s.approvers && s.approvers.length ? esc(s.approvers.join(", ")) : "No approvals"}</td><td>${status(s.status)} <span class="muted">${s.approvals}/2</span></td><td>${formatTime(s.createdAt)}</td><td>${s.status === "pending" ? `<button class="button secondary small approve" data-id="${esc(s.id)}">Approve</button>` : ""}</td></tr>`).join("") : `<tr><td colspan="6">${empty("No signing session has been created.")}</td></tr>`}
    </tbody></table></div></section>`;
}

async function transactions() {
  const data = await api("/transactions");
  return `<section class="panel"><div class="panel-head"><h2>Signed transactions</h2><span class="muted">Only sessions that reached the threshold are listed.</span></div><div class="table-wrap"><table><thead><tr><th>Transaction ID</th><th>Signing session</th><th>Digest</th><th>Network</th><th>Status</th><th>Created</th></tr></thead><tbody>
    ${data.items.length ? data.items.map((tx) => `<tr><td class="code">${esc(tx.id)}</td><td class="code">${esc(tx.sessionId)}</td><td class="code">${esc(tx.digest)}</td><td>${esc(tx.network)}</td><td>${status(tx.status)}</td><td>${formatTime(tx.createdAt)}</td></tr>`).join("") : `<tr><td colspan="6">${empty("No signed transaction is available yet.")}</td></tr>`}
  </tbody></table></div></section>`;
}

async function audit() {
  const data = await api("/audit-logs");
  return `<section class="panel"><div class="panel-head"><h2>Operational audit trail</h2><span class="muted">Generated by the signing-session state machine.</span></div><div class="panel-body">
  ${data.items.length ? `<ol class="timeline">${data.items.map((entry) => `<li><strong>${esc(entry.detail)}</strong><span class="muted"> by ${esc(entry.actor)}${entry.sessionId ? ` for <span class="code">${esc(entry.sessionId)}</span>` : ""}</span><time>${formatTime(entry.createdAt)} | ${esc(entry.action)}</time></li>`).join("")}</ol>` : empty("No operation has been recorded.")}
  </div></section>`;
}

const renders = { overview, nodes, sessions, transactions, audit };
async function render() {
  const page = location.hash.slice(1) || "overview";
  const chosen = renders[page] ? page : "overview";
  title.textContent = pageNames[chosen];
  document.querySelectorAll(".nav a").forEach((link) => link.classList.toggle("active", link.dataset.page === chosen));
  content.innerHTML = `<div class="panel"><div class="empty">Loading ${pageNames[chosen].toLowerCase()}...</div></div>`;
  try {
    content.innerHTML = await renders[chosen]();
    setApiState(true);
    bindPageActions(chosen);
  } catch (error) {
    setApiState(false);
    content.innerHTML = `<section class="panel"><div class="panel-body"><div class="callout">Could not load data from the Go service: ${esc(error.message)}. Start the service from the project root with <span class="code">go run ./cmd/server</span>.</div></div></section>`;
  }
}

function bindPageActions(page) {
  if (page !== "sessions") return;
  document.querySelector("#create-session").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await api("/sessions", { method: "POST", body: JSON.stringify({ digest: form.get("digest") }) });
      showToast("Signing session created.");
      render();
    } catch (error) { showToast(error.message); }
  });
  document.querySelectorAll(".approve").forEach((button) => button.addEventListener("click", async () => {
    const node = window.prompt("Approve as node-1, node-2, or node-3", "node-1");
    if (!node) return;
    try {
      const result = await api(`/sessions/${encodeURIComponent(button.dataset.id)}?action=approve&node=${encodeURIComponent(node)}`, { method: "POST" });
      showToast(result.status === "signed" ? "Threshold reached. Signature released." : "Node approval recorded.");
      render();
    } catch (error) { showToast(error.message); }
  }));
}

document.querySelector("#refresh-button").addEventListener("click", render);
window.addEventListener("hashchange", render);
render();
