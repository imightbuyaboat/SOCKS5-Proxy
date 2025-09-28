console.log("script.js: Loaded successfully");

async function postJSON(url, body = null) {
  console.log(`Sending request to ${url}`, body);
  let opts = { method: "POST", headers: { "Content-Type": "application/json" } };
  if (body) opts.body = JSON.stringify(body);
  try {
    let res = await fetch(url, opts);
    let txt = await res.text();
    console.log(`Response from ${url}:`, { status: res.status, text: txt });
    try {
      return { ok: res.ok, data: JSON.parse(txt) };
    } catch {
      return { ok: res.ok, data: { error: txt } };
    }
  } catch (err) {
    console.error(`Error fetching ${url}:`, err);
    return { ok: false, data: { error: err.message } };
  }
}

function showMessage(msg, isError = false) {
  console.log("Showing message:", msg, "isError:", isError);
  let div = document.getElementById("message");
  div.textContent = msg;
  div.className = "message " + (isError ? "error" : "success");
  div.style.display = "block";
  setTimeout(() => { div.style.display = "none"; }, 4000);
}

document.getElementById("saveBtn").onclick = async () => {
  console.log("Save button clicked");
  let body = {
    tcp_relay_server_address: document.getElementById("tcp_relay_server_address").value,
    udp_relay_server_address: document.getElementById("udp_relay_server_address").value,
  };
  let res = await postJSON("/save", body);
  if (res.ok) {
    showMessage(res.data.status);
  } else {
    showMessage(res.data.error, true);
    if (res.data.current_config) {
      console.log("Restoring fields to current config:", res.data.current_config);
      document.getElementById("tcp_relay_server_address").value = res.data.current_config.tcp_relay_server_address;
      document.getElementById("udp_relay_server_address").value = res.data.current_config.udp_relay_server_address;
    }
  }
};

document.getElementById("startBtn").onclick = async () => {
  console.log("Start button clicked");
  let res = await postJSON("/start");
  if (res.ok) showMessage(res.data.status);
  else showMessage(res.data.error, true);
};

document.getElementById("stopBtn").onclick = async () => {
  console.log("Stop button clicked");
  let res = await postJSON("/stop");
  if (res.ok) showMessage(res.data.status);
  else showMessage(res.data.error, true);
};

document.getElementById("logsBtn").onclick = async () => {
  console.log("Logs button clicked");
  try {
    let res = await fetch("/logs");
    let data = await res.json();
    console.log("Logs response:", data);
    let formattedLogs = data.logs
      .split('\n')
      .map(line => {
        try {
          const log = JSON.parse(line);
          return `${log.ts} [${log.level.toUpperCase()}] ${log.msg} ${Object.entries(log)
            .filter(([k]) => !['ts', 'level', 'msg', 'caller'].includes(k))
            .map(([k, v]) => `${k}=${JSON.stringify(v)}`)
            .join(' ')}`;
        } catch (e) {
          return line;
        }
      })
      .join('\n');
    document.getElementById("logsOutput").textContent = formattedLogs;
    document.getElementById("logsModal").style.display = "block";
  } catch (err) {
    console.error("Error fetching logs:", err);
    showMessage("Failed to fetch logs", true);
  }
};

document.getElementById("closeLogs").onclick = () => {
  console.log("Close logs clicked");
  document.getElementById("logsModal").style.display = "none";
};

document.getElementById("clearLogsBtn").onclick = async () => {
  console.log("Clear logs button clicked");
  try {
    let res = await fetch("/clear-logs", { method: "POST" });
    let data = await res.json();
    console.log("Clear logs response:", data);
    if (res.ok) {
      showMessage(data.status);
      document.getElementById("logsOutput").textContent = "";
    } else {
      showMessage(data.error, true);
    }
  } catch (err) {
    console.error("Error clear logs:", err);
    showMessage("Failed to clear logs", true);
  }
};