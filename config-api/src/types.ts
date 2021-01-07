import { validate } from "./fastidious/index";

/**
 * Finicky Configuration Reference.
 */

/**
 * This represents the full `.finicky.js` `module.exports` object.
 *
 * Example:
 *
 * ```js
 *  module.exports = {
 *    defaultBrowser: "Google Chrome",
 *    options: {
 *      hideIcon: false
 *    },
 *    handlers: [{
 *      match: finicky.matchHostnames("example.com'),
 *      browser: "Firefox"
 *    }]
 *  }
 * ```
 */
export interface FinickyConfig {
  /** The default browser or app to open for urls where no other handler
   * matches.
   */
  defaultBrowser: Browser | BrowserFunction | Array<Browser | BrowserFunction>;
  options?: {
    /** Whether or not to hide the finicky icon in the menu bar */
    hideIcon?: boolean;
    /** An array of domain names to replace the built in list of url
     * shortener domains. Note that using this option replaces the list
     * completely.
     *
     * Example:
     * ```js
     *  module.exports = {
     *    options: {
     *      urlShorteners: ["another-url-shortener-service.com"]
     *    }
     *  }
     * ```
     */
    urlShorteners?: string[];
  };
  /** An array of Rewriters that can change the url being opened */
  rewrite?: Rewriter[];
  /** An array of Handlers to select which browser to open for urls */
  handlers?: Handler[];
}

export type BrowserResult =
  | Browser
  | BrowserFunction
  | Array<Browser | BrowserFunction>;

/**
 * A handler contains a matcher and a browser. If the matcher matches when opening a url, the browser in the handler will be opened.
 */
export interface Handler {
  match: Matcher | Matcher[];
  url?: PartialUrl | UrlFunction;
  browser: BrowserResult;
}

/**
 * A rewriter contains a matcher and a url. If the matcher matches when opening a url, the final url will be changed to whatever the url property is.
 */
export interface Rewriter {
  match: Matcher | Matcher[];
  url: PartialUrl | UrlFunction;
}

/**
 * Matches urls (or other properties) to decide if to change the url or open a browser.
 *
 * If the matcher is a string, if the url equals the string exactly, it will match.
 * If the matcher is a regular expression, if it matches any part of the url, it will match.
 * If the matcher is a [[MatcherFunction]], it will match if the function returns `true`
 *
 */
export type Matcher = string | RegExp | MatcherFunction;
export type MatcherFunction = (options: Options) => boolean;

/**
 * Represents a browser or app that finicky could start
 */
export type Browser = string | BrowserObject;

/**
 * Represents a browser or app to open
 */
export interface BrowserObject {
  name: string;
  appType?: "appName" | "bundleId" | "appPath" | "none";
  openInBackground?: boolean;
  profile?: string;
  args?: string[];
  passUrlAsArg?: boolean,
}

/**
 * A function that returns a browser to open
 */
type BrowserFunction = (options: Options) => Browser;

/**
 * Represents a url that will be handled by finicky
 */
export type Url = string | UrlObject;

/**
 * Represents a url that will be handled by finicky
 */
export type PartialUrl = string | Partial<UrlObject>;

/**
 * An object that represents a url
 */
export interface UrlObject {
  protocol: string;
  username?: string;
  password?: string;
  host: string;
  port?: number;
  pathname?: string;
  search?: string;
  hash?: string;
}

/**
 * A function that returns a url
 */
export type UrlFunction = (options: Options) => PartialUrl;

/**
 * Options sent as the argument to [[ProcessUrl]]
 */
export interface ProcessOptions {
  /** If opened in from an app, this string contains the bundle identifier from that app */
  sourceBundleIdentifier?: string;
  /** The state of keyboard state. E.g. shift === true if pressed. */
  keys: {
    shift: boolean;
    option: boolean;
    command: boolean;
    control: boolean;
    capsLock: boolean;
    function: boolean;
  };
}

/**
 * Options sent as the argument to [[MatcherFunction]], [[BrowserFunction]] and [[UrlFunction]]
 */
export interface Options extends ProcessOptions {
  /** The url being opened */
  urlString: string;
  /** The url being opened as an object */
  url: UrlObject;
}

const urlObjectSchema = {
  protocol: validate.string.isRequired,
  username: validate.string,
  password: validate.string,
  host: validate.string.isRequired,
  port: validate.oneOf([validate.number, validate.value(null)]),
  pathname: validate.string,
  search: validate.string,
  hash: validate.string
}

const partialUrlSchema = {
  ...urlObjectSchema,
  protocol: validate.string,
  host: validate.string,
};


export const urlSchema = {
  url: validate.oneOf([
    validate.string,
    validate.shape(urlObjectSchema)
  ]).isRequired
};



const browserSchema = validate.oneOf([
  validate.string,
  validate.shape({
    name: validate.string.isRequired,
    appType: validate.oneOf(["appName", "appPath", "bundleId"]),
    openInBackground: validate.boolean,
    profile: validate.string,
    args: validate.arrayOf(validate.string),
    passUrlAsArg: validate.boolean,
  }),
  validate.function("options"),
  validate.value(null)
]);

const multipleBrowsersSchema = validate.oneOf([
  browserSchema,
  validate.arrayOf(browserSchema.isRequired)
]);

const matchSchema = validate.oneOf([
  validate.string,
  validate.function("options"),
  validate.regex,
  validate.arrayOf(
    validate.oneOf([
      validate.string,
      validate.function("options"),
      validate.regex
    ])
  )
]);

export const finickyConfigSchema = {
  defaultBrowser: multipleBrowsersSchema.isRequired,
  options: validate.shape({
    hideIcon: validate.boolean,
    urlShorteners: validate.arrayOf(validate.string),
    checkForUpdate: validate.boolean
  }),
  rewrite: validate.arrayOf(
    validate.shape({
      match: matchSchema.isRequired,
      url: validate.oneOf([validate.string, validate.shape(partialUrlSchema), validate.function("options")])
        .isRequired
    }).isRequired
  ),
  handlers: validate.arrayOf(
    validate.shape({
      match: matchSchema.isRequired,
      url: validate.oneOf([validate.string, validate.shape(partialUrlSchema), validate.function("options")]),
      browser: multipleBrowsersSchema.isRequired
    })
  )
};
