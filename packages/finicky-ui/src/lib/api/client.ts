const BASE_URL: string = (() => {
  const port = new URLSearchParams(window.location.search).get("apiPort");
  if (port) {
    const url = `http://127.0.0.1:${port}/api`;
    localStorage.setItem("devApiUrl", url);
    return url;
  }
  return (
    window.__FINICKY_API__ ??
    import.meta.env.VITE_API_URL ??
    localStorage.getItem("devApiUrl") ??
    "http://127.0.0.1:0/api"
  );
})();

// The native app injects a per-process secret into the WebView so the local
// REST/SSE API can reject requests from anything else (e.g. a website
// scanning localhost ports). Dev builds can pass the same token as a query
// param when pointed at a real running instance via VITE_API_URL. It's never
// persisted to localStorage — that secret shouldn't outlive the page/process
// that generated it.
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
