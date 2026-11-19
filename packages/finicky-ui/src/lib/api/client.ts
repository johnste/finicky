// The UI is served by the app's own local HTTP server, so the API is on the
// same origin. In development, Vite proxies /api to a running instance (see
// vite.config.ts).
const BASE_URL = "/api";

// The native app injects a per-process secret into the WebView so the local
// REST/SSE API can reject requests from anything else (e.g. a website
// scanning localhost ports). Dev builds can pass the same token as a query
// param when proxying to a real running instance. It's never persisted to
// localStorage — that secret shouldn't outlive the page/process that
// generated it.
const API_TOKEN: string = (() => {
  const token = new URLSearchParams(window.location.search).get("apiToken");
  if (token) return token;
  return window.__FINICKY_API_TOKEN__ ?? "";
})();

export function base(): string {
  return BASE_URL;
}

export function authToken(): string {
  return API_TOKEN;
}

export async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${base()}${path}`, {
    headers: { "X-Finicky-Token": API_TOKEN },
  });
  if (!res.ok) throw new Error(`GET ${path} failed: ${res.status}`);
  return res.json();
}

export async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${base()}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Finicky-Token": API_TOKEN,
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
