import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { FinickyURL } from "./FinickyURL";
import { ProcessInfo } from "./configSchema";
import * as legacyURLObjectModule from "./legacyURLObject";

describe("FinickyURL", () => {
  // Setup console.warn spy
  let consoleWarnSpy: any;

  beforeEach(() => {
    // Mock console.warn to track deprecation warnings
    consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    // Restore console.warn after each test
    consoleWarnSpy.mockRestore();
  });

  describe("constructor", () => {
    it("should create a FinickyURL instance with a valid URL", () => {
      const url = new FinickyURL("https://example.com");
      expect(url).toBeInstanceOf(FinickyURL);
      expect(url).toBeInstanceOf(URL);
      expect(url.href).toBe("https://example.com/");
    });

    it("should create a FinickyURL instance with opener information", () => {
      const opener: ProcessInfo = {
        name: "TestApp",
        bundleId: "com.test.app",
        path: "/Applications/TestApp.app",
      };
      const url = new FinickyURL("https://example.com", opener);
      expect(url).toBeInstanceOf(FinickyURL);
      expect(url.href).toBe("https://example.com/");
    });

    it("should throw an error for invalid URLs", () => {
      expect(() => new FinickyURL("invalid-url")).toThrow();
    });
  });

  describe("urlString property", () => {
    it("should return the href and show deprecation warning", () => {
      const url = new FinickyURL(
        "https://example.com/path?query=value#fragment"
      );

      expect(url.urlString).toBe(
        "https://example.com/path?query=value#fragment"
      );
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        expect.stringContaining('Accessing legacy property "urlString"')
      );
    });
  });

  describe("url property", () => {
    it("should return a LegacyURLObject and show deprecation warning", () => {
      const url = new FinickyURL(
        "https://user:pass@example.com:8080/path?query=value#fragment"
      );
      const spy = vi.spyOn(legacyURLObjectModule, "URLtoLegacyURLObject");

      const legacyUrl = url.url;

      expect(consoleWarnSpy).toHaveBeenCalledWith(
        expect.stringContaining('Accessing legacy property "url"')
      );
      expect(spy).toHaveBeenCalledWith(url);

      // Verify the legacy URL object properties
      expect(legacyUrl.host).toBe("example.com");
      expect(legacyUrl.protocol).toBe("https");
      expect(legacyUrl.pathname).toBe("/path");
      expect(legacyUrl.search).toBe("query=value");
      expect(legacyUrl.username).toBe("user");
      expect(legacyUrl.password).toBe("pass");
      expect(legacyUrl.port).toBe(8080);
      expect(legacyUrl.hash).toBe("fragment");
    });
  });

  describe("opener property", () => {
    it("should return the opener and show deprecation warning", () => {
      const opener: ProcessInfo = {
        name: "TestApp",
        bundleId: "com.test.app",
        path: "/Applications/TestApp.app",
      };
      const url = new FinickyURL("https://example.com", opener);

      expect(url.opener).toEqual(opener);
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        expect.stringContaining('Accessing legacy property "opener"')
      );
    });

    it("should return null when no opener is provided", () => {
      const url = new FinickyURL("https://example.com");

      expect(url.opener).toBeNull();
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        expect.stringContaining('Accessing legacy property "opener"')
      );
    });
  });

  describe("keys property", () => {
    it("should throw an error with a message about using finicky.getModifierKeys()", () => {
      const url = new FinickyURL("https://example.com");

      expect(() => url.keys).toThrow(
        'Accessing legacy property "keys" that is no longer supported, please use finicky.getModifierKeys() instead.'
      );
    });
  });

  describe("URL inheritance", () => {
    it("should inherit all URL properties", () => {
      const url = new FinickyURL(
        "https://user:pass@example.com:8080/path?query=value#fragment"
      );

      expect(url.href).toBe(
        "https://user:pass@example.com:8080/path?query=value#fragment"
      );
      expect(url.protocol).toBe("https:");
      expect(url.hostname).toBe("example.com");
      expect(url.host).toBe("example.com:8080");
      expect(url.pathname).toBe("/path");
      expect(url.search).toBe("?query=value");
      expect(url.hash).toBe("#fragment");
      expect(url.username).toBe("user");
      expect(url.password).toBe("pass");
      expect(url.port).toBe("8080");
      expect(url.origin).toBe("https://example.com:8080");
    });

    it("should inherit URL methods", () => {
      const url = new FinickyURL("https://example.com/path");

      url.searchParams.append("query", "value");
      expect(url.href).toBe("https://example.com/path?query=value");

      url.pathname = "/newpath";
      expect(url.href).toBe("https://example.com/newpath?query=value");
    });
  });
});
