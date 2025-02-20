// Initialize message buffer if not exists
window.messageBuffer ||= [];

addEventListener("error", (event) => {
  addMessageToLog(
    "Internal error. This is a bug, please report it to the developers: " +
      event.message,
    true
  );
});

// Function to add a message to the log
function addMessageToLog(msg, error = false) {
  // FIXME: Temporary hack to tag error messages properly, replace with a better solution
  if (msg.includes("error")) {
    error = true;
  }

  const logContent = document.getElementById("logContent");
  if (logContent) {
    const entry = document.createElement("li");
    let content = typeof msg === "string" ? msg : JSON.stringify(msg);

    // Extract date/time prefix if present
    const dateTimeMatch = content.match(
      /^(\d{4}\/\d{2}\/\d{2}) (\d{2}:\d{2}:\d{2}\.\d+) (.*)$/
    );
    if (dateTimeMatch) {
      const date = dateTimeMatch[1];
      const time = dateTimeMatch[2].replace(/(\.\d{2})\d+/, "$1"); // Truncate to 2 decimal places
      const dateSpan = document.createElement("span");
      dateSpan.className = "log-date";
      dateSpan.textContent = time;
      dateSpan.title = `${date} ${time}`; // Full date and time in tooltip
      entry.appendChild(dateSpan);
      content = dateTimeMatch[3];
    }

    const message = document.createElement("div");
    message.className = error ? "log-message error" : "log-message";
    message.textContent = content;

    entry.appendChild(message);

    logContent.appendChild(entry);
    // Auto-scroll to bottom
    logContent.scrollTop = logContent.scrollHeight;
  }
}

// Handle existing messages in buffer
window.messageBuffer.forEach(addMessageToLog);

window.finicky = {
  sendMessage: function (msg) {
    window.webkit.messageHandlers.finicky.postMessage(JSON.stringify(msg));
  },
  receiveMessage: function (msg) {
    try {
      const parsedMsg = JSON.parse(msg);
      window.messageBuffer.push(parsedMsg);
      handleMessage(parsedMsg);
    } catch (e) {
      // Handle plain text messages
      window.messageBuffer.push({ type: "log", message: e.toString() });
      handleMessage({ type: "log", message: e.toString() });
    }
  },
};

function handleMessage(msg) {
  switch (msg.type) {
    case "status":
      document.getElementById("status").textContent = msg.message;
      break;
    case "version":
      document.getElementById("version").textContent = msg.message;
      break;
    case "buildInfo":
      document.getElementById("buildInfo").textContent = msg.message;
      break;
    case "configInfo":
      document.getElementById("configInfo").textContent = " - " + msg.message;
      break;
    default:
      addMessageToLog(msg.message);
      break;
  }
}

// Add clear log functionality
document.getElementById("clearLog").addEventListener("click", function () {
  const logContent = document.getElementById("logContent");
  if (logContent) {
    logContent.innerHTML = "";
    window.messageBuffer = [];
  }
});
