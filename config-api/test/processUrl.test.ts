import { Matcher, Handler, Rewriter, UrlObject } from "./../src/types";
import { processUrl } from "../src/processUrl";
import { createAPI } from "../src/createAPI";
import { Url, UrlFunction } from "../src/types";

const createRewriteConfig = ({
  urlResult = "https://test.changed",
  match = () => true,
}: {
  urlResult?: Partial<UrlObject> | Url | UrlFunction;
  match?: Matcher;
} = {}) => {
  return {
    defaultBrowser: "test",
    rewrite: [{ match, url: urlResult }],
  };
};

const processOptions = {
  opener: {
    pid: 1337,
    path: "/dev/null",
    name: "Finicky",
    bundleId: "net.kassett.Finicky",
  },
};

describe("Rewrites", () => {
  describe("Rewrite matcher", () => {
    beforeAll(() => {
      // @ts-ignore
      global.finickyInternalAPI = {
        getKeys: () => ({
          shift: false,
          option: false,
          command: false,
          control: false,
          capsLock: false,
          function: false,
        }),
      };

      // @ts-ignore
      global.finicky = createAPI();
    });

    test("function that returns true", () => {
      const config = createRewriteConfig({ match: () => true });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test.changed");
    });

    test("function that returns false", () => {
      const config = createRewriteConfig({ match: () => false });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).not.toBe("https://test.changed");
    });

    test("match regular expression", () => {
      const config = createRewriteConfig({ match: /test\.example/ });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test.changed");
    });

    test("match string", () => {
      const config = createRewriteConfig({ match: "https://test.example" });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test.changed");
    });

    test("match wildcard pattern", () => {
      const config = createRewriteConfig({ match: "https://test.example/*" });
      const result = processUrl(
        config,
        "https://test.example" + "/path?query=123#anchor",
        processOptions
      );
      expect(result.url).toBe("https://test.changed");
    });
  });

  describe("Rewrite url", () => {
    beforeAll(() => {
      // @ts-ignore
      global.finicky = createAPI();
    });

    test("String", () => {
      const config = createRewriteConfig({ urlResult: "https://test.changed" });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test.changed");
    });

    test("Function", () => {
      const config = createRewriteConfig({
        urlResult: () => "https://test.changed",
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test.changed");
    });

    test("Function arguments", () => {
      const config = createRewriteConfig({
        urlResult: ({ urlString }) => urlString + "?ok",
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test.example?ok");
    });

    test("Function argument object", () => {
      const config = createRewriteConfig({
        urlResult: ({ urlString, url }) => urlString + "?" + url.protocol,
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test.example?https");
    });

    test("Object result ", () => {
      const config = createRewriteConfig({
        urlResult: ({ url }) => ({
          ...url,
          host: "test2.example",
        }),
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.url).toBe("https://test2.example");
    });
  });

  describe("Rewrite partial url", () => {
    beforeAll(() => {
      // @ts-ignore
      global.finicky = createAPI();
    });

    test("Protocol change", () => {
      const config = createRewriteConfig({ urlResult: { protocol: "ftp" } });
      const result = processUrl(config, "http://example.com", processOptions);
      expect(result.url).toBe("ftp://example.com");
    });

    test("Hostname change", () => {
      const config = createRewriteConfig({
        urlResult: { host: "example.org" },
      });
      const result = processUrl(config, "http://example.com", processOptions);
      expect(result.url).toBe("http://example.org");
    });

    test("Multiple change", () => {
      const config = createRewriteConfig({
        urlResult: { hash: "anchor", port: 1234, pathname: "/a/path" },
      });
      const result = processUrl(config, "http://example.com", processOptions);
      expect(result.url).toBe("http://example.com:1234/a/path#anchor");
    });
  });
});

const EXAMPLE_BROWSER = "Default Browser";
const CHANGED_BROWSER = "Custom Browser";
const EXAMPLE_BUNDLEID = "bundle.id";

const createConfig = ({
  rewrite = [],
  handlers = [],
}: {
  rewrite?: Rewriter[];
  handlers?: Handler[];
} = {}) => {
  return {
    defaultBrowser: EXAMPLE_BROWSER,
    rewrite,
    handlers,
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
        handlers: [{ browser: CHANGED_BROWSER, match: () => true }],
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
    });

    test("function that returns false", () => {
      const config = createConfig({
        handlers: [{ browser: CHANGED_BROWSER, match: () => false }],
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.browsers[0].name).not.toBe(CHANGED_BROWSER);
    });

    test("match regular expression", () => {
      const config = createConfig({
        handlers: [{ browser: CHANGED_BROWSER, match: /test\.example/ }],
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
    });

    test("match string", () => {
      const config = createConfig({
        handlers: [{ browser: CHANGED_BROWSER, match: "https://test.example" }],
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
    });

    test("match wildcard pattern", () => {
      const config = createConfig({
        handlers: [
          { browser: CHANGED_BROWSER, match: "https://test.example/*" },
        ],
      });
      const result = processUrl(
        config,
        "https://test.example" + "/path?query=123#anchor",
        processOptions
      );
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
    });
  });

  describe("Browser", () => {
    describe("Browser name", () => {
      test("string", () => {
        const config = createConfig({
          handlers: [{ browser: CHANGED_BROWSER, match: () => true }],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName",
        });
      });

      test("string function", () => {
        const config = createConfig({
          handlers: [{ browser: () => CHANGED_BROWSER, match: () => true }],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName",
        });
      });

      test("object", () => {
        const config = createConfig({
          handlers: [{ browser: { name: CHANGED_BROWSER }, match: () => true }],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName",
        });
      });

      test("object function", () => {
        const config = createConfig({
          handlers: [
            { browser: () => ({ name: CHANGED_BROWSER }), match: () => true },
          ],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: CHANGED_BROWSER,
          appType: "appName",
        });
      });
    });

    describe("Bundle id", () => {
      test("string", () => {
        const config = createConfig({
          handlers: [{ browser: EXAMPLE_BUNDLEID, match: () => true }],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId",
        });
      });

      test("string function", () => {
        const config = createConfig({
          handlers: [{ browser: () => EXAMPLE_BUNDLEID, match: () => true }],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId",
        });
      });

      test("object", () => {
        const config = createConfig({
          handlers: [
            { browser: { name: EXAMPLE_BUNDLEID }, match: () => true },
          ],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId",
        });
      });

      test("object function", () => {
        const config = createConfig({
          handlers: [
            { browser: () => ({ name: EXAMPLE_BUNDLEID }), match: () => true },
          ],
        });
        const result = processUrl(
          config,
          "https://test.example",
          processOptions
        );
        expect(result.browsers[0]).toEqual({
          name: EXAMPLE_BUNDLEID,
          appType: "bundleId",
        });
      });
    });
  });

  describe("Handlers with browser and url rewrite", () => {
    test("that matches and rewrites the url", () => {
      const config = createConfig({
        handlers: [
          {
            browser: CHANGED_BROWSER,
            url: "https://test.changed",
            match: () => true,
          },
        ],
      });
      const result = processUrl(config, "https://test.example", processOptions);
      expect(result.browsers[0].name).toBe(CHANGED_BROWSER);
      expect(result.url).toBe("https://test.changed");
    });
  });
});
