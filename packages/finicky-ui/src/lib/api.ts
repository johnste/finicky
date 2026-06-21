const BASE_URL: string = (() => {
  const port = new URLSearchParams(window.location.search).get("apiPort");
  if (port) {
    const url = `http://127.0.0.1:${port}/api`;
    localStorage.setItem("devApiUrl", url);
    return url;
  }
  return (
    (window as any).__FINICKY_API__ ??
    import.meta.env.VITE_API_URL ??
    localStorage.getItem("devApiUrl") ??
    "http://127.0.0.1:0/api"
  );
})();

// The native app injects a per-process secret into the WebView so the local
// REST/SSE API can reject requests from anything else (e.g. a website
// scanning localhost ports). Dev builds can pass the same token as a query
// param when pointed at a real running instance via VITE_API_URL.
const API_TOKEN: string = (() => {
  const token = new URLSearchParams(window.location.search).get("apiToken");
  if (token) {
    localStorage.setItem("devApiToken", token);
    return token;
  }
  return (
    (window as any).__FINICKY_API_TOKEN__ ??
    localStorage.getItem("devApiToken") ??
    ""
  );
})();

function base(): string {
  return BASE_URL;
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${base()}${path}`, {
    headers: { "X-Finicky-Token": API_TOKEN },
  });
  if (!res.ok) throw new Error(`GET ${path} failed: ${res.status}`);
  return res.json();
}

async function post<T>(path: string, body: unknown): Promise<T> {
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

export const api = {
  initialData: () => get<Record<string, unknown>>("/initial-data"),
  getRules: () => get<unknown>("/rules"),
  saveRules: (payload: unknown) => post<unknown>("/rules", payload),
  getBrowsers: () => get<string[]>("/browsers"),
  getBrowserProfiles: (browser: string) =>
    get<string[]>(`/browser-profiles?browser=${encodeURIComponent(browser)}`),
  testUrl: (url: string) => post<unknown>("/test-url", { url }),
  eventsUrl: () => `${base()}/events?token=${encodeURIComponent(API_TOKEN)}`,
};
