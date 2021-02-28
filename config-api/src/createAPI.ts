import urlParse from "url-parse";
import { ConfigAPI, UrlObject, Matcher, Options } from "./types";

declare const finickyInternalAPI: any;

/**
 * Extends finicky js api with some utility functions.
 */
export function createAPI(overrides: Partial<ConfigAPI> = {}): ConfigAPI {
  if (typeof finickyInternalAPI === "undefined") {
    throw new Error(`[finicky warn] finickyInternalAPI not available`);
  }

  return {
    log: (...messages: string[]) => finickyInternalAPI.log(messages.join(" ")),
    notify: (title: string, subtitle: string) =>
      finickyInternalAPI.notify(title, subtitle),
    getBattery: () => finickyInternalAPI.getBattery(),
    getKeys: () => finickyInternalAPI.getKeys(),
    getSystemInfo: () => finickyInternalAPI.getSystemInfo(),
    parseUrl,
    getUrlParts: parseUrl,
    matchHostnames,
    matchDomains: matchHostnames,
    ...overrides,
  };
}

/**
 * Helper function to parse urls
 */
export function parseUrl(urlString: string) {
  const url = urlParse(urlString);

  // Mistake in the urlParse typings. query should be a string unless parsing of query is enabled
  const search = (url.query as unknown) as string;

  return {
    username: url.username,
    host: url.hostname,
    protocol: url.protocol.replace(":", ""),
    pathname: url.pathname,
    search: search.replace("?", ""),
    password: url.password,
    port: url.port ? +url.port : undefined,
    hash: url.hash.replace("#", ""),
  } as UrlObject;
}

/**
 * Helper function to match hostnames
 */
export function matchHostnames(matchers: Matcher | Matcher[], ...args: any[]) {
  if (args.length > 0) {
    throw new Error(
      "finicky.matchHostnames only accepts one argument. See https://johnste.github.io/finicky-docs/interfaces/_finickyapi_.finicky.html#matchdomains for more information"
    );
  }

  matchers = Array.isArray(matchers) ? matchers : [matchers];

  matchers.forEach((matcher) => {
    if (matcher instanceof RegExp || typeof matcher === "string") {
      return;
    }
    throw new Error(
      `finicky.matchHostnames: Unrecognized hostname "${matcher}"`
    );
  });

  return function Matcher({ url }: Options) {
    const { host } = url;
    return (matchers as Matcher[]).some((matcher) => {
      if (matcher instanceof RegExp) {
        return matcher.test(host);
      } else if (typeof matcher === "string") {
        return matcher === host;
      }

      return false;
    });
  };
}
