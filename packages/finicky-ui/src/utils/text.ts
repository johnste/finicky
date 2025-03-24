import type { LogEntry } from "../types";

export type TextPart = {
  type: "text" | "url";
  content: string;
};

/**
 * Splits text into parts, separating URLs from regular text
 * @param text The input text to process
 * @returns Array of text and URL parts
 */
export function splitTextAndUrls(text: string): TextPart[] {
  const urlRegex = /((https?|file):\/\/[^\s]+)/g;
  const parts: TextPart[] = [];
  let lastIndex = 0;
  let match;

  while ((match = urlRegex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push({
        type: "text",
        content: text.slice(lastIndex, match.index),
      });
    }
    parts.push({
      type: "url",
      content: match[0],
    });
    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < text.length) {
    parts.push({
      type: "text",
      content: text.slice(lastIndex),
    });
  }

  return parts;
}

/**
 * Formats a log entry into text parts with URLs properly handled
 * @param entry The log entry to format
 * @returns Array of text and URL parts
 */
export function formatLogEntry(entry: LogEntry): TextPart[] {
  let result = entry.msg;

  if (entry.error) {
    result += `\nError: ${entry.error}`;
  }

  // Add any additional fields
  const additionalFields = Object.entries(entry).filter(
    ([key]) => !["level", "msg", "time", "error"].includes(key)
  );

  if (additionalFields.length > 0) {
    result +=
      "\n" +
      additionalFields.map(([key, value]) => `${key}: ${value}`).join("\n");
  }

  return splitTextAndUrls(result);
}
