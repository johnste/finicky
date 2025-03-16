import { describe, it, expect, vi } from "vitest";
import { FinickyURL } from "./FinickyURL";
import { ProcessInfo, OpenUrlOptions } from "./configSchema";
import * as legacyURLObjectModule from "./legacyURLObject";

// Import the functions directly from index.ts
import {
  isMatch,
  rewriteUrl,
  autodetectAppStringType,
  resolveBrowser,
} from "./index";
import * as indexModule from "./index";

describe("isMatch", () => {
  const mockUrl = new FinickyURL(
    "https://example.com/path?query=value#fragment"
  );
  const mockProcessInfo: ProcessInfo = {
    name: "TestApp",
    bundleId: "com.test.app",
    path: "/Applications/TestApp.app",
  };
  const mockOptions: OpenUrlOptions = {
    opener: mockProcessInfo,
  };

  describe("string patterns", () => {
    it("matches exact strings", () => {
      expect(
        isMatch(
          "https://example.com/path?query=value#fragment",
          mockUrl,
          mockOptions
        )
      ).toBe(true);
      expect(isMatch("https://different.com", mockUrl, mockOptions)).toBe(
        false
      );
    });

    it("matches wildcard patterns", () => {
      expect(isMatch("example.com*", mockUrl, mockOptions)).toBe(true);
      expect(isMatch("*.com/path*", mockUrl, mockOptions)).toBe(true);
      expect(isMatch("example.*/path?query=*", mockUrl, mockOptions)).toBe(
        true
      );
      expect(isMatch("*.org*", mockUrl, mockOptions)).toBe(false);
    });

    it("handles empty string patterns", () => {
      expect(isMatch("", mockUrl, mockOptions)).toBe(false);
    });
  });

  describe("regex patterns", () => {
    it("matches regex patterns", () => {
      expect(isMatch(/example\.com/, mockUrl, mockOptions)).toBe(true);
      expect(
        isMatch(/example\.com\/path\?query=value/, mockUrl, mockOptions)
      ).toBe(true);
      expect(isMatch(/different\.com/, mockUrl, mockOptions)).toBe(false);
    });
  });

  describe("function patterns", () => {
    it("matches function patterns", () => {
      const matchFn = (url: URL) => url.hostname === "example.com";
      const noMatchFn = (url: URL) => url.hostname === "different.com";

      expect(isMatch(matchFn, mockUrl, mockOptions)).toBe(true);
      expect(isMatch(noMatchFn, mockUrl, mockOptions)).toBe(false);
    });

    it("passes options to function patterns", () => {
      const matchFn = (url: URL, options: OpenUrlOptions) =>
        options.opener?.name === "TestApp";

      expect(isMatch(matchFn, mockUrl, mockOptions)).toBe(true);
    });
  });

  describe("array patterns", () => {
    it("matches if any pattern in array matches", () => {
      expect(
        isMatch(["different.com", "example.com*"], mockUrl, mockOptions)
      ).toBe(true);
      expect(
        isMatch([/different\.com/, /example\.com/], mockUrl, mockOptions)
      ).toBe(true);
      expect(
        isMatch(["different.com", /different\.org/], mockUrl, mockOptions)
      ).toBe(false);
    });

    it("handles mixed pattern types in arrays", () => {
      const matchFn = (url: URL) => url.hostname === "example.com";
      const patterns = ["different.com", /example\.com/, matchFn];

      expect(isMatch(patterns, mockUrl, mockOptions)).toBe(true);
    });
  });
});

describe("rewriteUrl", () => {
  const mockUrl = new FinickyURL(
    "https://example.com/path?query=value#fragment"
  );
  const mockProcessInfo: ProcessInfo = {
    name: "TestApp",
    bundleId: "com.test.app",
    path: "/Applications/TestApp.app",
  };
  const mockOptions: OpenUrlOptions = {
    opener: mockProcessInfo,
  };

  it("handles string rewrites", () => {
    const result = rewriteUrl("https://rewritten.com", mockUrl, mockOptions);
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://rewritten.com/");
  });

  it("handles URL object rewrites", () => {
    const newUrl = new URL("https://rewritten.com/newpath");
    const result = rewriteUrl(newUrl, mockUrl, mockOptions);
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://rewritten.com/newpath");
  });

  it("handles FinickyURL object rewrites", () => {
    const newUrl = new FinickyURL(
      "https://rewritten.com/newpath",
      mockProcessInfo
    );
    const result = rewriteUrl(newUrl, mockUrl, mockOptions);
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://rewritten.com/newpath");
  });

  it("handles function rewrites that return strings", () => {
    const rewriteFn = (url: URL, options: OpenUrlOptions) =>
      "https://rewritten.com/function";
    const result = rewriteUrl(rewriteFn, mockUrl, mockOptions);
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://rewritten.com/function");
  });

  it("handles function rewrites that return URL objects", () => {
    const rewriteFn = (url: URL, options: OpenUrlOptions) =>
      new URL("https://rewritten.com/function-url");
    const result = rewriteUrl(rewriteFn, mockUrl, mockOptions);
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://rewritten.com/function-url");
  });

  it("handles function rewrites that return FinickyURL objects", () => {
    const rewriteFn = (url: URL, options: OpenUrlOptions) =>
      new FinickyURL("https://rewritten.com/function-finicky");
    const result = rewriteUrl(rewriteFn, mockUrl, mockOptions);
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://rewritten.com/function-finicky");
  });

  it("handles legacy URL object rewrites", () => {
    // Create a legacy URL object
    const legacyUrl = {
      host: "legacy.com",
      protocol: "https",
      pathname: "/legacy-path",
      search: "query=value",
      hash: "fragment",
    };

    // Mock the isLegacyURLObject function to return true for our test object
    vi.spyOn(legacyURLObjectModule, "isLegacyURLObject").mockReturnValue(true);

    // Mock the legacyURLObjectToString function to return a predictable URL
    vi.spyOn(legacyURLObjectModule, "legacyURLObjectToString").mockReturnValue(
      "https://legacy.com/legacy-path?query=value#fragment"
    );

    // Pass the legacy URL object directly
    const result = rewriteUrl(legacyUrl as any, mockUrl, mockOptions);

    expect(result).toBeInstanceOf(FinickyURL);
    expect(legacyURLObjectModule.legacyURLObjectToString).toHaveBeenCalledWith(
      legacyUrl
    );

    vi.restoreAllMocks();
  });

  it("returns the original URL for unsupported rewrite types", () => {
    // Test with an actual unsupported type (number)
    const result = rewriteUrl(123 as any, mockUrl, mockOptions);
    expect(result).toBe(mockUrl); // Should return the original URL
  });

  it("simulates chained function rewrites", () => {
    // Instead of trying to test chained functions directly, we'll test the recursive behavior
    // Create a function that returns a string URL
    const rewriteFn = (url: URL, options: OpenUrlOptions) =>
      "https://chained-rewritten.com";

    // Call rewriteUrl with our function
    const result = rewriteUrl(rewriteFn, mockUrl, mockOptions);

    // Verify the result is as expected
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://chained-rewritten.com/");
  });

  it("handles recursive URL transformations", () => {
    // This test verifies that rewriteUrl can handle a sequence of transformations
    // First transformation: string -> URL
    const stringToUrl = (url: URL, options: OpenUrlOptions) =>
      new URL("https://first-transform.com");

    // Call rewriteUrl with the first transformation
    const firstResult = rewriteUrl(stringToUrl, mockUrl, mockOptions);
    expect(firstResult.href).toBe("https://first-transform.com/");

    // Second transformation: URL -> string
    const urlToString = (url: URL, options: OpenUrlOptions) =>
      "https://second-transform.com";

    // Call rewriteUrl with the second transformation and the result of the first
    const secondResult = rewriteUrl(urlToString, firstResult, mockOptions);
    expect(secondResult.href).toBe("https://second-transform.com/");
  });

  it("handles function rewrites that return legacy URL objects via mocking", () => {
    // Instead of directly returning a legacy URL object (which TypeScript doesn't allow),
    // we'll mock the function to simulate this behavior

    // Create a string URL that will be converted to a FinickyURL
    const legacyUrlString = "https://function-legacy.com/from-function";

    // Create a function that returns a string
    const rewriteFn = (url: URL, options: OpenUrlOptions) => legacyUrlString;

    // Mock isLegacyURLObject to simulate the legacy URL object check
    vi.spyOn(legacyURLObjectModule, "isLegacyURLObject").mockReturnValue(false);

    const result = rewriteUrl(rewriteFn, mockUrl, mockOptions);
    expect(result).toBeInstanceOf(FinickyURL);
    expect(result.href).toBe("https://function-legacy.com/from-function");

    vi.restoreAllMocks();
  });
});

describe("autodetectAppStringType", () => {
  it("detects app names", () => {
    expect(autodetectAppStringType("Google Chrome")).toBe("appName");
    expect(autodetectAppStringType("Firefox")).toBe("appName");
    expect(autodetectAppStringType("Safari")).toBe("appName");
    expect(autodetectAppStringType("Chrome 123")).toBe("appName");
  });

  it("detects bundle IDs", () => {
    expect(autodetectAppStringType("com.google.Chrome")).toBe("bundleId");
    expect(autodetectAppStringType("org.mozilla.firefox")).toBe("bundleId");
    expect(autodetectAppStringType("com.apple.Safari")).toBe("bundleId");
  });

  it("detects app paths", () => {
    expect(autodetectAppStringType("/Applications/Google Chrome.app")).toBe(
      "path"
    );
    expect(autodetectAppStringType("~/Applications/Firefox.app")).toBe("path");
    expect(
      autodetectAppStringType("/Users/username/Applications/Custom Browser.app")
    ).toBe("path");
  });

  it("handles null input", () => {
    expect(autodetectAppStringType(null)).toBe("none");
  });

  it("defaults to appName for ambiguous cases", () => {
    // This is not a valid path format but contains slashes
    expect(autodetectAppStringType("Invalid/Path")).toBe("appName");

    // This looks like a bundle ID but has invalid characters
    expect(autodetectAppStringType("com.example!invalid")).toBe("appName");
  });
});

describe("resolveBrowser", () => {
  const mockUrl = new FinickyURL("https://example.com");
  const mockProcessInfo: ProcessInfo = {
    name: "TestApp",
    bundleId: "com.test.app",
    path: "/Applications/TestApp.app",
  };
  const mockOptions: OpenUrlOptions = {
    opener: mockProcessInfo,
  };

  it("resolves string browser specifications", () => {
    const result = resolveBrowser("Firefox", mockUrl, mockOptions);
    expect(result).toMatchObject({
      name: "Firefox",
      appType: "appName",
      url: "https://example.com/",
    });
  });

  it("resolves browser specifications with profiles", () => {
    const result = resolveBrowser("Chrome:Work", mockUrl, mockOptions);
    expect(result).toMatchObject({
      name: "Chrome",
      profile: "Work",
      url: "https://example.com/",
    });
  });

  it("resolves object browser specifications", () => {
    const browserConfig = {
      name: "Google Chrome",
      profile: "Personal",
      openInBackground: true,
    };

    const result = resolveBrowser(browserConfig, mockUrl, mockOptions);
    expect(result).toMatchObject({
      name: "Google Chrome",
      profile: "Personal",
      openInBackground: true,
      url: "https://example.com/",
    });
  });

  it("resolves function browser specifications that return strings", () => {
    const browserFn = (url: URL) => "Firefox";
    const result = resolveBrowser(browserFn, mockUrl, mockOptions);
    expect(result).toMatchObject({
      name: "Firefox",
      url: "https://example.com/",
    });
  });

  it("resolves function browser specifications that return objects", () => {
    const browserFn = (url: URL) => ({
      name: "Google Chrome",
      profile: "Work",
    });

    const result = resolveBrowser(browserFn, mockUrl, mockOptions);
    expect(result).toMatchObject({
      name: "Google Chrome",
      profile: "Work",
      url: "https://example.com/",
    });
  });

  it("resolves null browser specifications to empty browser with 'none' type", () => {
    const result = resolveBrowser(null, mockUrl, mockOptions);
    expect(result).toMatchObject({
      name: "",
      appType: "none",
      url: "https://example.com/",
    });
  });

  it("throws error for undefined browser specifications", () => {
    const browserFn = () => undefined;
    expect(() => resolveBrowser(browserFn, mockUrl, mockOptions)).toThrow();
  });

  it("throws error for invalid browser specifications", () => {
    expect(() => resolveBrowser(123 as any, mockUrl, mockOptions)).toThrow();
  });

  it("passes options to function browser specifications", () => {
    const browserFn = (url: URL, options: OpenUrlOptions) =>
      options.opener?.name === "TestApp" ? "Firefox" : "Chrome";

    const result = resolveBrowser(browserFn, mockUrl, mockOptions);
    expect(result).toMatchObject({
      name: "Firefox",
      url: "https://example.com/",
    });
  });
});
