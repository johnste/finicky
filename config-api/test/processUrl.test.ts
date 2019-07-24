import { Matcher, BrowserResult, Handler, Rewriter } from "./../src/types";
import { UrlObject } from "./../dist/src/types.d";
import { processUrl } from "../src/processUrl";
import { createAPI } from "../src/createAPI";
import { Url, UrlFunction } from "../src/types";

const EXAMPLE_URL = "https://test.example";
const CHANGED_URL = "https://test.changed";

const createRewriteConfig = ({
  urlResult = CHANGED_URL,
  match = () => true
}: { urlResult?: Url | UrlFunction; match?: Matcher } = {}) => {
  return {
    defaultBrowser: "test",
    rewrite: [{ match, url: urlResult }]
  };
};

describe("Rewrites", () => {
  describe("Rewrite matcher", () => {
    beforeAll(() => {
      // @ts-ignore
      global.finicky = createAPI();
    });

    test("function that returns true", () => {
      const config = createRewriteConfig({ match: () => true });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe(CHANGED_URL);
    });

    test("function that returns false", () => {
      const config = createRewriteConfig({ match: () => false });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).not.toBe(CHANGED_URL);
    });

    test("match regular expression", () => {
      const config = createRewriteConfig({ match: /test\.example/ });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe(CHANGED_URL);
    });

    test("match string", () => {
      const config = createRewriteConfig({ match: EXAMPLE_URL });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe(CHANGED_URL);
    });
  });

  describe("Rewrite url", () => {
    beforeAll(() => {
      // @ts-ignore
      global.finicky = createAPI();
    });

    test("String", () => {
      const config = createRewriteConfig({ urlResult: CHANGED_URL });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe(CHANGED_URL);
    });

    test("Function", () => {
      const config = createRewriteConfig({ urlResult: () => CHANGED_URL });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe(CHANGED_URL);
    });

    test("Function arguments", () => {
      const config = createRewriteConfig({
        urlResult: ({ urlString }) => urlString + "?ok"
      });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe("https://test.example?ok");
    });

    test("Function argument object", () => {
      const config = createRewriteConfig({
        urlResult: ({ urlString, url }) => urlString + "?" + url.protocol
      });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe("https://test.example?https");
    });

    test("Object result ", () => {
      const config = createRewriteConfig({
        urlResult: ({ url }) => ({
          ...url,
          host: "test2.example"
        })
      });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.url).toBe("https://test2.example");
    });
  });
});

const EXAMPLE_BROWSER = "Default Browser";
const CHANGED_BROWSER = "Custom Browser";
const EXAMPLE_BUNDLEID = "bundle.id";

const createConfig = ({
  rewrite = [],
  handlers = []
}: {
  rewrite?: Rewriter[];
  handlers?: Handler[];
} = {}) => {
  return {
    defaultBrowser: EXAMPLE_BROWSER,
    rewrite,
    handlers
  };
};

describe("Handlers", () => {
  describe("Matcher", () => {
    beforeAll(() => {
      // @ts-ignore
      global.finicky = createAPI();
    });

    test("function that returns true", () => {
      const config = createConfig({
        handlers: [{ browser: CHANGED_BROWSER, match: () => true }]
      });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
    });

    test("function that returns false", () => {
      const config = createConfig({
        handlers: [{ browser: CHANGED_BROWSER, match: () => false }]
      });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.browsers[0].name).not.toBe(CHANGED_BROWSER);
    });

    test("match regular expression", () => {
      const config = createConfig({
        handlers: [{ browser: CHANGED_BROWSER, match: /test\.example/ }]
      });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
    });

    test("match string", () => {
      const config = createConfig({
        handlers: [{ browser: CHANGED_BROWSER, match: EXAMPLE_URL }]
      });
      const result = processUrl(config, EXAMPLE_URL);
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
    });
  });

  describe("Browser", () => {
    describe("Browser name", () => {
      test("string", () => {
        const config = createConfig({
          handlers: [{ browser: CHANGED_BROWSER, match: () => true }]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName"
        });
      });

      test("string function", () => {
        const config = createConfig({
          handlers: [{ browser: () => CHANGED_BROWSER, match: () => true }]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName"
        });
      });

      test("object", () => {
        const config = createConfig({
          handlers: [{ browser: { name: CHANGED_BROWSER }, match: () => true }]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName"
        });
      });

      test("object function", () => {
        const config = createConfig({
          handlers: [
            { browser: () => ({ name: CHANGED_BROWSER }), match: () => true }
          ]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName"
        });
      });
    });

    describe("Bundle id", () => {
      test("string", () => {
        const config = createConfig({
          handlers: [{ browser: EXAMPLE_BUNDLEID, match: () => true }]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId"
        });
      });

      test("string function", () => {
        const config = createConfig({
          handlers: [{ browser: () => EXAMPLE_BUNDLEID, match: () => true }]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId"
        });
      });

      test("object", () => {
        const config = createConfig({
          handlers: [{ browser: { name: EXAMPLE_BUNDLEID }, match: () => true }]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId"
        });
      });

      test("object function", () => {
        const config = createConfig({
          handlers: [
            { browser: () => ({ name: EXAMPLE_BUNDLEID }), match: () => true }
          ]
        });
        const result = processUrl(config, EXAMPLE_URL);
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId"
        });
      });
    });
  });
});
