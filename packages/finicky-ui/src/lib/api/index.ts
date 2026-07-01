import { base, authToken, get, post } from "./client";

export const api = {
  initialData: () => get<Record<string, unknown>>("/initial-data"),
  getRules: () => get<unknown>("/rules"),
  saveRules: (payload: unknown) => post<unknown>("/rules", payload),
  getBrowsers: () => get<string[]>("/browsers"),
  getBrowserProfiles: (browser: string) =>
    get<string[]>(`/browser-profiles?browser=${encodeURIComponent(browser)}`),
  testUrl: (url: string) => post<unknown>("/test-url", { url }),
  eventsUrl: () => `${base()}/events?token=${encodeURIComponent(authToken())}`,
};
