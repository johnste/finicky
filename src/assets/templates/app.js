// Initialize message buffer if not exists
window.messageBuffer ||= [];

addEventListener("error", (event) => {
  addMessageToLog({
    level: "ERROR",
    msg:
      "Internal error. This is a bug, please report it to the developers: " +
      event.message,
    time: new Date().toISOString(),
  });
});

// Function to add a message to the log
function addMessageToLog({ level, msg, time, error, ...rest }) {
  const logContent = document.getElementById("logContent");
  if (logContent) {
    const entry = document.createElement("li");
    entry.className = `log-entry log-entry-level-${level.toLowerCase()}`;

    const dateSpan = document.createElement("span");
    dateSpan.className = "log-date";
    dateSpan.textContent = time.split(" ")[0];
    dateSpan.title = time;
    entry.appendChild(dateSpan);

    const timeSpan = document.createElement("span");
    timeSpan.className = "log-time";
    const hoursAndMinutes = time.split(" ")[1].slice(0, 6);
    const seconds = String(
      parseFloat(time.split(" ")[1].slice(6)).toFixed(2)
    ).padStart(5, "0");
    timeSpan.textContent = hoursAndMinutes + seconds;
    timeSpan.title = time;
    entry.appendChild(timeSpan);

    const messageEl = document.createElement("div");
    messageEl.className = `log-message log-level-${level.toLowerCase()}`;

    messageEl.textContent = msg;

    if (error) {
      const errorEl = document.createElement("div");
      errorEl.className = "log-error";
      errorEl.textContent = error;
      messageEl.appendChild(errorEl);
    }

    // Add any additional fields from rest
    for (const [label, value] of Object.entries(rest)) {
      const additionalEl = document.createElement("div");
      additionalEl.className = "log-additional";
      additionalEl.textContent = `${label}: ${value}`;
      messageEl.appendChild(additionalEl);
    }

    entry.appendChild(messageEl);

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
    const parsedMsg = JSON.parse(msg);
    window.messageBuffer.push(parsedMsg);
    handleMessage(parsedMsg);
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
      addMessageToLog(JSON.parse(msg.message));
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

// Add debug toggle functionality
const toggleDebugButton = document.getElementById("toggleDebug");
toggleDebugButton.addEventListener("click", function () {
  const logContent = document.getElementById("logContent");
  if (logContent) {
    logContent.classList.toggle("show-debug");

    // Update button text and style based on the current state
    if (logContent.classList.contains("show-debug")) {
      toggleDebugButton.textContent = "Hide Debug";
      toggleDebugButton.classList.add("active");
    } else {
      toggleDebugButton.textContent = "Show Debug";
      toggleDebugButton.classList.remove("active");
    }
  }
});
