import { describe, it, expect, vi } from "vitest";
import { openUrl } from "./index";
import { Config, ProcessInfo } from "./configSchema";

describe("openUrl", () => {
  const mockProcessInfo: ProcessInfo = {
    name: "TestApp",
    bundleId: "com.test.app",
    path: "/Applications/TestApp.app",
  };

  const baseConfig = {
    defaultBrowser: "Firefox",
    options: {},
    rewrite: [
      {
        match: "https://slack-redir.net/link?url=*",
        url: (url: URL) => url.href,
      },
    ],
    handlers: [
      {
        match: "open.spotify.com",
        open: "Spotify",
      },
      {
        match: ["github.com", /amazon/],
        open: "Google Chrome",
      },
    ],
  };

  // Complex config with varied matching options
  const complexConfig: Config = {
    defaultBrowser: "Safari",
    options: {
      logRequests: true,
    },
    rewrite: [
      {
        match: /^https:\/\/t\.co\/.+/,
        url: (url: URL) => url.href.replace("t.co", "twitter.com"),
      },
      {
        match: (url: URL) => url.href.includes("shortened"),
        url: (url: URL) => url.href + "/expanded",
      },
      {
        match: ["example.com", "test.com", "sample.org"],
        url: (url: URL) => url.href + "/new-path",
      },
    ],
    handlers: [
      {
        match: ["figma.com", "sketch.com"],
        open: {
          name: "Google Chrome",
          profile: "Design",
          openInBackground: false,
        },
      },
      {
        match: (url: URL) => url.pathname.includes("docs"),
        open: (url: URL) => "Google Chrome",
      },
    ],
  };

  it("should throw error for invalid config", () => {
    const invalidConfig = {
      defaultBrowser: 123, // Invalid type for defaultBrowser
    };

    expect(() =>
      openUrl("https://example.com", 1234, mockProcessInfo, invalidConfig)
    ).toThrow("Invalid config");
  });

  it("should use default browser for unmatched URLs", () => {
    const url = "https://example.com";
    const result = openUrl(url, 1234, mockProcessInfo, baseConfig);
    expect(result).toEqual({
      name: "Firefox",
      appType: "bundleId",
      openInBackground: false,
      profile: "",
      args: [],
      url: "https://example.com/",
    });
  });

  it("should handle Spotify URLs correctly", () => {
    const url = "https://open.spotify.com/track/123";
    const result = openUrl(url, 1234, mockProcessInfo, baseConfig);
    expect(result).toEqual({
      name: "Spotify",
      appType: "bundleId",
      openInBackground: false,
      profile: "",
      args: [],
      url: "https://open.spotify.com/track/123",
    });
  });

  it("should handle multiple matches in handlers array", () => {
    const url = "https://github.com/some-repo";
    const result = openUrl(url, 1234, mockProcessInfo, baseConfig);
    expect(result).toEqual({
      name: "Google Chrome",
      appType: "name",
      openInBackground: false,
      profile: "",
      args: [],
      url: "https://github.com/some-repo",
    });
  });

  it("should apply URL rewrites before handling", () => {
    const url = "https://slack-redir.net/link?url=https://example.com";
    const result = openUrl(url, 1234, mockProcessInfo, baseConfig);
    expect(result).toEqual({
      name: "Firefox",
      appType: "bundleId",
      openInBackground: false,
      profile: "",
      args: [],
      url: "https://slack-redir.net/link?url=https://example.com",
    });
  });

  it("should handle regex matches in handlers", () => {
    const url = "https://amazon.com/product";
    const result = openUrl(url, 1234, mockProcessInfo, baseConfig);
    expect(result).toEqual({
      name: "Google Chrome",
      appType: "name",
      openInBackground: false,
      profile: "",
      args: [],
      url: "https://amazon.com/product",
    });
  });

  it("should work with null opener", () => {
    const url = "https://example.com";
    const result = openUrl(url, 1234, null, baseConfig);
    expect(result).toEqual({
      name: "Firefox",
      appType: "bundleId",
      openInBackground: false,
      profile: "",
      args: [],
      url: "https://example.com/",
    });
  });

  describe("Complex matching scenarios", () => {
    it("should handle array of mixed matchers (string, regex, function)", () => {
      const urls = [
        "https://figma.com/file/123",
        "https://notion.so/workspace",
        "https://my.workspace.com/dashboard",
        "http://localhost:3000",
      ];

      urls.forEach((url) => {
        const result = openUrl(url, 1234, mockProcessInfo, complexConfig);
        expect(result).toEqual({
          name: "Google Chrome",
          appType: "name",
          openInBackground: false,
          profile: "Design",
          args: [],
          url: new URL(url).href,
        });
      });
    });

    it("should handle regex-based rewrites", () => {
      const url = "https://t.co/abc123";
      const result = openUrl(url, 1234, mockProcessInfo, complexConfig);
      expect(result).toEqual({
        name: "Safari",
        appType: "bundleId",
        openInBackground: false,
        profile: "",
        args: [],
        url: "https://twitter.com/abc123",
      });
    });

    it("should handle function-based matches", () => {
      const url = "https://my.shortened.url/123";
      const result = openUrl(url, 1234, mockProcessInfo, complexConfig);
      expect(result).toEqual({
        name: "Safari",
        appType: "bundleId",
        openInBackground: false,
        profile: "",
        args: [],
        url: url + "/expanded",
      });
    });

    it("should handle browser objects with profiles and options", () => {
      const url = "https://figma.com/file/123";
      const result = openUrl(url, 1234, mockProcessInfo, complexConfig);
      expect(result).toEqual({
        name: "Google Chrome",
        appType: "name",
        openInBackground: false,
        profile: "Design",
        args: [],
        url: url,
      });
    });

    it("should handle function-based browser selection", () => {
      const urls = [
        "https://example.com/doc.docx",
        "https://docs.google.com/document/123",
        "https://sheets.google.com/spreadsheet/456",
      ];

      urls.forEach((url) => {
        const result = openUrl(url, 1234, mockProcessInfo, complexConfig);
        expect(result).toEqual({
          name: "Google Chrome",
          appType: "name",
          openInBackground: false,
          profile: "",
          args: [],
          url: new URL(url).href,
        });
      });
    });

    it("should handle multiple rewrite rules in sequence", () => {
      const url = "https://tas.com/shortened/123";
      const result = openUrl(url, 1234, mockProcessInfo, complexConfig);
      expect(result).toEqual({
        name: "Safari",
        appType: "bundleId",
        openInBackground: false,
        profile: "",
        args: [],
        url: url + "/expanded",
      });
    });

    describe("Wildcard matching", () => {
      const wildcardConfig = {
        defaultBrowser: "Safari",
        handlers: [
          {
            match: "*.example.com",
            open: { name: "Chrome", appType: "bundleId" },
          },
          {
            match: "https://*.github.io/*",
            open: { name: "Firefox", appType: "bundleId" },
          },
          {
            match: ["*.google.*", "mail.*.com"],
            open: { name: "Chrome", appType: "bundleId" },
          },
          {
            match: "test\\*literal", // Testing escaped asterisk
            open: { name: "Safari", appType: "bundleId" },
          },
        ],
      };

      it("should match domain wildcards", () => {
        const urls = [
          "https://sub1.example.com",
          "https://sub2.example.com/path",
          "http://another.example.com",
        ];

        urls.forEach((url) => {
          const result = openUrl(url, 1234, mockProcessInfo, wildcardConfig);
          expect(result).toEqual({
            name: "Chrome",
            appType: "bundleId",
            openInBackground: false,
            profile: "",
            args: [],
            url: new URL(url).href,
          });
        });
      });

      it("should match path wildcards", () => {
        const urls = [
          "https://user1.github.io/repo",
          "https://org.github.io/docs/api",
        ];

        urls.forEach((url) => {
          const result = openUrl(url, 1234, mockProcessInfo, wildcardConfig);
          expect(result).toEqual({
            name: "Firefox",
            appType: "bundleId",
            openInBackground: false,
            profile: "",
            args: [],
            url: new URL(url).href,
          });
        });
      });

      it("should not match non-matching wildcards", () => {
        const urls = [
          "https://example.com", // No subdomain
          "https://github.io", // No subdomain and no path
          "http://github.io/test", // No subdomain
        ];

        urls.forEach((url) => {
          const result = openUrl(url, 1234, mockProcessInfo, wildcardConfig);
          expect(result).toEqual({
            name: "Safari",
            appType: "bundleId",
            openInBackground: false,
            profile: "",
            args: [],
            url: new URL(url).href,
          });
        });
      });

      it("should match multiple wildcards", () => {
        const urls = [
          "https://mail.google.com",
          "https://mail.yahoo.com",
          "https://docs.google.co.uk",
        ];

        urls.forEach((url) => {
          const result = openUrl(url, 1234, mockProcessInfo, wildcardConfig);
          expect(result).toEqual({
            name: "Chrome",
            appType: "bundleId",
            openInBackground: false,
            profile: "",
            args: [],
            url: new URL(url).href,
          });
        });
      });

      it("should handle escaped asterisks literally", () => {
        const url = "https://test*literal";
        const result = openUrl(url, 1234, mockProcessInfo, wildcardConfig);
        expect(result).toEqual({
          name: "Safari",
          appType: "bundleId",
          openInBackground: false,
          profile: "",
          args: [],
          url: url.endsWith("/") ? url : url + "/",
        });
      });
    });

    describe("String matching edge cases", () => {
      const edgeCaseConfig = {
        defaultBrowser: "Safari",
        handlers: [
          {
            match: "", // Empty string
            open: { name: "Chrome", appType: "bundleId" },
          },
          {
            match: "https://", // Protocol only
            open: { name: "Firefox", appType: "bundleId" },
          },
          {
            match: "?param=value", // Query parameter
            open: { name: "Edge", appType: "bundleId" },
          },
        ],
      };

      it("should handle empty string matches", () => {
        const url = "https://example.com";
        const result = openUrl(url, 1234, mockProcessInfo, edgeCaseConfig);
        expect(result).toEqual({
          name: "Chrome",
          appType: "bundleId",
          openInBackground: false,
          profile: "",
          args: [],
          url: new URL(url).href,
        });
      });

      it("should match protocol only", () => {
        const url = "https://example.com";
        const result = openUrl(url, 1234, mockProcessInfo, edgeCaseConfig);
        expect(result).toEqual({
          name: "Firefox",
          appType: "bundleId",
          openInBackground: false,
          profile: "",
          args: [],
          url: new URL(url).href,
        });
      });

      it("should match query parameters", () => {
        const url = "https://example.com?param=value&other=123";
        const result = openUrl(url, 1234, mockProcessInfo, edgeCaseConfig);
        expect(result).toEqual({
          name: "Edge",
          appType: "bundleId",
          openInBackground: false,
          profile: "",
          args: [],
          url: url.replace("?", "/?"),
        });
      });
    });
  });
});
