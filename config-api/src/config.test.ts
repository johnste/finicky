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
        { match: "open.spotify.com*", open: "Spotify" },
        { match: ["github.com*", /amazon/], open: "Google Chrome" },
        {
          match: ["figma.com*", "sketch.com*"],
          open: { name: "Google Chrome", profile: "Design" },
        },
        {
          match: (url: URL) => url.pathname.includes("docs"),
          open: "Google Chrome",
        },
      ],
    };

    const cases = [
      { url: "https://example.com", expected: "Firefox" },
      { url: "https://open.spotify.com/track/123", expected: "Spotify" },
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
        expect(result).toMatchObject(
          typeof expected === "string" ? { name: expected } : expected
        );
      });
    });

    it("works with null opener", () => {
      const result = openUrl("https://example.com", 1234, null, handlerConfig);
      expect(result).toMatchObject({ name: "Firefox" });
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
    ];

    cases.forEach(({ input, expectedUrl }) => {
      it(`rewrites ${input}`, () => {
        const result = openUrl(input, mockProcessInfo, rewriteConfig);
        expect(result).toMatchObject({ name: "Safari", url: expectedUrl });
      });
    });
  });

  describe("wildcards", () => {
    const wildcardConfig = {
      defaultBrowser: "Safari",
      handlers: [
        { match: "*.example.com*", open: "Chrome" },
        { match: ["*.google.*", "mail.*.com*"], open: "Chrome" },
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
          expect(result).toMatchObject({ name: expected });
        });
      });
    });
  });

  describe("edge cases", () => {
    const edgeCaseConfig = {
      defaultBrowser: "Safari",
      handlers: [
        { match: "", open: "Chrome" },
        { match: "https://*", open: "Firefox" },
        { match: "*?param=value*", open: "Edge" },
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
        expect(result).toMatchObject({ name: expected });
      });
    });
  });
});
