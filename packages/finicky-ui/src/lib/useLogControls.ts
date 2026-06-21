import { useState } from "react";
import { toast } from "./toast";
import type { LogEntry } from "../types";

function formatEntry(entry: LogEntry): string {
  const time = new Date(entry.time).toISOString();
  const base = `[${time}] [${entry.level.padEnd(5)}] ${entry.msg}`;
  const extra = Object.entries(entry)
    .filter(([key]) => !["level", "msg", "time", "error"].includes(key))
    .map(([key, value]) => `${key}: ${value}`)
    .join(" | ");
  const parts = [base];
  if (extra) parts.push(extra);
  if (entry.error) parts.push(`Error: ${entry.error}`);
  return parts.join(" | ");
}

export function useLogControls(messageBuffer: LogEntry[]) {
  const [showDebug, setShowDebug] = useState(
    () => localStorage.getItem("showDebugLogs") === "true"
  );

  function handleShowDebugChange(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.checked;
    setShowDebug(val);
    localStorage.setItem("showDebugLogs", String(val));
  }

  async function copyLogs() {
    const text = messageBuffer
      .filter((entry) => showDebug || entry.level.toLowerCase() !== "debug")
      .map(formatEntry)
      .join("\n");
    try {
      await navigator.clipboard.writeText(text);
      toast.success("Logs copied to clipboard");
    } catch (err) {
      toast.error("Failed to copy logs", err instanceof Error ? err.message : String(err));
    }
  }

  return { showDebug, handleShowDebugChange, copyLogs };
}
