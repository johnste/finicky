import { writable } from 'svelte/store';

export interface TestUrlResult {
  browser: string;
  url: string;
  openInBackground: boolean;
  profile?: string;
}

export const testUrlResult = writable<TestUrlResult | null>(null);
