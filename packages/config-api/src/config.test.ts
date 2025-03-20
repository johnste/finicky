import { describe, it, expect, vi } from "vitest";
import { openUrl } from "./index";
import { Config, ProcessInfo } from "./configSchema";

describe("openUrl", () => {
  const mockProcessInfo: ProcessInfo = {
    name: "TestApp",
    bundleId: "com.test.app",
    path: "/Applications/TestApp.app",
  };

  it("validation", () => {
    expect(() =>
      openUrl("https://example.com", mockProcessInfo, {
        defaultBrowser: 123,
      })
    ).toThrow("Invalid config");
  });

  describe("handlers", () => {
    const handlerConfig = {
      defaultBrowser: "Firefox",
      handlers: [
        { match: "browser.spotify.com*", browser: "Spotify" },
        { match: ["github.com*", /amazon/], browser: "Google Chrome" },
        {
          match: ["figma.com*", "sketch.com*"],
          browser: { name: "Google Chrome", profile: "Design" },
        },
        {
          match: (url: URL) => url.pathname.includes("docs"),
          browser: "Google Chrome",
        },
      ],
    };

    const cases = [
      { url: "https://example.com", expected: "Firefox" },
      { url: "https://browser.spotify.com/track/123", expected: "Spotify" },
      { url: "https://github.com/some-repo", expected: "Google Chrome" },
      { url: "https://amazon.com/product", expected: "Google Chrome" },
      {
        url: "https://figma.com/file/123",
        expected: { name: "Google Chrome", profile: "Design" },
      },
      {
        url: "https://sketch.com/dashboard",
        expected: { name: "Google Chrome", profile: "Design" },
      },
      { url: "https://example.com/docs", expected: "Google Chrome" },
    ];

    cases.forEach(({ url, expected }) => {
      it(`handles ${url}`, () => {
        const result = openUrl(url, mockProcessInfo, handlerConfig);
        expect(result.browser).toMatchObject(
          typeof expected === "string" ? { name: expected } : expected
        );
      });
    });

    it("works with null opener", () => {
      const result = openUrl("https://example.com", null, handlerConfig);
      expect(result.browser).toMatchObject({ name: "Firefox" });
    });
  });

  describe("rewrites", () => {
    const rewriteConfig = {
      defaultBrowser: "Safari",
      rewrite: [
        {
          match: "https://slack-redir.net/link?url=*",
          url: (url: URL) => url.href,
        },
        {
          match: /^https:\/\/t\.co\/.+/,
          url: (url: URL) => url.href.replace("t.co", "twitter.com"),
        },
        {
          match: (url: URL) => url.href.includes("shortened"),
          url: (url: URL) => url.href + "/expanded",
        },
        {
          match: (url: URL) => url.href.includes("use-url-object"),
          url: (url: URL) => new URL("https://example.com"),
        },
      ],
    };

    const cases = [
      {
        input: "https://slack-redir.net/link?url=https://example.com",
        expectedUrl: "https://slack-redir.net/link?url=https://example.com",
      },
      {
        input: "https://t.co/abc123",
        expectedUrl: "https://twitter.com/abc123",
      },
      {
        input: "https://my.shortened.url/123",
        expectedUrl: "https://my.shortened.url/123/expanded",
      },
      {
        input: "https://www.use-url-object.com",
        expectedUrl: "https://example.com/",
      },
    ];

    cases.forEach(({ input, expectedUrl }) => {
      it(`rewrites ${input}`, () => {
        const result = openUrl(input, mockProcessInfo, rewriteConfig);
        expect(result.browser).toMatchObject({
          name: "Safari",
          url: expectedUrl,
        });
      });
    });
  });

  describe("wildcards", () => {
    const wildcardConfig = {
      defaultBrowser: "Safari",
      handlers: [
        { match: "*.example.com*", browser: "Chrome" },
        { match: ["*.google.*", "mail.*.com*"], browser: "Chrome" },
      ],
    };

    const cases = [
      {
        urls: ["https://sub1.example.com", "https://sub2.example.com/path"],
        expected: "Chrome",
      },
      {
        urls: [
          "https://mail.google.com",
          "https://docs.google.co.uk",
          "https://mail.yahoo.com",
        ],
        expected: "Chrome",
      },
      {
        urls: ["https://example.com", "https://google.com"],
        expected: "Safari",
      },
    ];

    cases.forEach(({ urls, expected }) => {
      urls.forEach((url) => {
        it(`${expected} handles ${url}`, () => {
          const result = openUrl(url, mockProcessInfo, wildcardConfig);
          expect(result.browser).toMatchObject({ name: expected });
        });
      });
    });
  });

  describe("edge cases", () => {
    const edgeCaseConfig = {
      defaultBrowser: "Safari",
      handlers: [
        { match: "", browser: "Chrome" },
        { match: "https://*", browser: "Firefox" },
        { match: "*?param=value*", browser: "Edge" },
      ],
    };

    const cases = [
      { url: "http://example.com", expected: "Safari" },
      { url: "https://example.com", expected: "Firefox" },
      { url: "http://example.com/?param=value&other=123", expected: "Edge" },
    ];

    cases.forEach(({ url, expected }) => {
      it(`${expected} handles ${url}`, () => {
        const result = openUrl(url, mockProcessInfo, edgeCaseConfig);
        expect(result.browser).toMatchObject({ name: expected });
      });
    });
  });
});
