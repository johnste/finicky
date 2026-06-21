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

function base(): string {
  return BASE_URL;
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${base()}${path}`);
  if (!res.ok) throw new Error(`GET ${path} failed: ${res.status}`);
  return res.json();
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${base()}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
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
  eventsUrl: () => `${base()}/events`,
};
