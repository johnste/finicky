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

// Track whether to auto-scroll with new messages
let shouldAutoScroll = true;

// Add scroll event listener to update auto-scroll preference
document.addEventListener("DOMContentLoaded", function () {
  const logContent = document.getElementById("logContent");
  if (logContent) {
    logContent.addEventListener(
      "scroll",
      function () {
        // Check if we're at the bottom of the log content
        const scrollPosition = logContent.scrollTop;
        const scrollHeight = logContent.scrollHeight;
        const clientHeight = logContent.clientHeight;
        shouldAutoScroll = scrollPosition + clientHeight >= scrollHeight - 50;
      },
      { passive: true }
    );
  }
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

    // Only auto-scroll if shouldAutoScroll is true
    if (shouldAutoScroll) {
      logContent.scrollTop = logContent.scrollHeight;
    }
  }
}

// Handle existing messages in buffer
window.messageBuffer.forEach(addMessageToLog);

function copyLogs() {
  const logContent = document.getElementById("logContent");
  if (logContent) {
    const logEntries = Array.from(logContent.children)
      .map((entry) => {
        const time = entry.querySelector(".log-time").title;
        const message = entry.querySelector(".log-message");
        // Get the log level from the message class
        const levelClass = Array.from(message.classList).find((cls) =>
          cls.startsWith("log-level-")
        );
        const level = levelClass
          ? levelClass.replace("log-level-", "").toUpperCase()
          : "INFO";

        // Get the main message text (direct text content, excluding child elements)
        let messageText = Array.from(message.childNodes)
          .filter((node) => node.nodeType === Node.TEXT_NODE)
          .map((node) => node.textContent.trim())
          .join(" ");

        // Get error message if it exists
        const errorEl = message.querySelector(".log-error");
        if (errorEl) {
          messageText += " | Error: " + errorEl.textContent.trim();
        }

        // Get additional fields if they exist
        const additionalFields = Array.from(
          message.querySelectorAll(".log-additional")
        );
        if (additionalFields.length > 0) {
          messageText +=
            " | " +
            additionalFields
              .map((field) => field.textContent.trim())
              .join(" | ");
        }

        return `[${time}] [${level.padEnd(5)}] ${messageText}`;
      })
      .join("\n"); // Single line break between entries

    navigator.clipboard
      .writeText(logEntries)
      .then(() => {
        const copyButton = document.getElementById("copyLog");
        const originalText = copyButton.textContent;
        copyButton.textContent = "Copied!";
        setTimeout(() => {
          copyButton.textContent = originalText;
        }, 2000);
      })
      .catch((err) => {
        console.error("Failed to copy logs:", err);
      });
  }
}

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

// Add copy logs functionality
const copyButton = document.createElement("button");
copyButton.id = "copyLog";
copyButton.textContent = "Copy All";
copyButton.addEventListener("click", copyLogs);
document.querySelector(".log-header-buttons").appendChild(copyButton);

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
