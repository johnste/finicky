import { Matcher, Options, UrlObject } from "./types";
import urlParse from "url-parse";

type LogFunction = (message: string) => void;
type NotifyFunction = (title: string, subtitle: string) => void;
type BatteryFunction = () =>
  | { chargePercentage: number; isCharging: boolean }
  | undefined;

/**
 * Extends finicky js api with some utility functions.
 */
export function createAPI(
  options: {
    log?: LogFunction;
    notify?: NotifyFunction;
    getBattery?: BatteryFunction;
  } = {}
) {
  const log: LogFunction =
    options.log ||
    ((message: string) => {
      // @ts-ignore
      if (typeof finickyInternalAPI !== "undefined") {
        // @ts-ignore
        finickyInternalAPI.log(message);
      } else {
        console.log(`[finicky log] ${message}`);
      }
    });

  const notify: NotifyFunction =
    options.notify ||
    ((title: string, subtitle: string) => {
      // @ts-ignore
      if (typeof finickyInternalAPI !== "undefined") {
        // @ts-ignore
        finickyInternalAPI.notify(title, subtitle);
      } else {
        console.log(`[finicky notify] ${title} ${subtitle}`);
      }
    });

  const getBattery: BatteryFunction =
    options.getBattery ||
    (() => {
      // @ts-ignore
      if (typeof finickyInternalAPI !== "undefined") {
        // @ts-ignore
        let status = finickyInternalAPI.getBattery();
        return status;
      } else {
        return undefined;
      }
    });

  const getUrlParts = (urlString: string): UrlObject => {
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
    };
  };

  const matchDomains = (matchers: Matcher | Matcher[], ...args: any[]) => {
    if (args.length > 0) {
      throw new Error(
        "finicky.matchDomains/matchHostnames only accepts one argument. See https://johnste.github.io/finicky-docs/interfaces/_finickyapi_.finicky.html#matchdomains for more information"
      );
    }

    if (!Array.isArray(matchers)) {
      matchers = [matchers];
    }

    matchers.forEach((matcher) => {
      if (matcher instanceof RegExp || typeof matcher === "string") {
        return;
      }
      throw new Error(
        `finicky.matchDomains/matchHostnames: Unrecognized hostname "${matcher}"`
      );
    });

    return function ({ url }: Options) {
      const domain = url.host;
      return (matchers as Matcher[]).some((matcher) => {
        if (matcher instanceof RegExp) {
          return matcher.test(domain);
        } else if (typeof matcher === "string") {
          return matcher === domain;
        }

        return false;
      });
    };
  };

  // Warn when using deprecated API methods
  const onUrl = () => {
    log(
      "finicky.onUrl is no longer supported in this version of Finicky, please go to https://github.com/johnste/finicky for updated documentation"
    );
    notify(
      "finicky.onUrl is no longer supported",
      "Check the Finicky website for updated documentation"
    );
  };

  const setDefaultBrowser = () => {
    log(
      "finicky.setDefaultBrowser is no longer supported in this version of Finicky, please go to https://github.com/johnste/finicky for updated documentation"
    );
    notify(
      "finicky.setDefaultBrowser is no longer supported",
      "Check the Finicky website for updated documentation"
    );
  };

  return {
    log,
    notify,
    matchDomains,
    matchHostnames: matchDomains,
    getUrlParts,
    onUrl,
    setDefaultBrowser,
    getBattery,
  };
}
