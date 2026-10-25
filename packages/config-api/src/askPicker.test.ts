import { describe, it, expect } from "vitest";

import { createBrowserConfig } from "./index";

describe("createBrowserConfig with picker options", () => {
  it("defaults options to an empty list", () => {
    expect(createBrowserConfig("Google Chrome").options).toEqual([]);
    expect(createBrowserConfig(null).options).toEqual([]);
  });

  it("keeps the ask app type and its own title", () => {
    const config = createBrowserConfig({
      name: "ask",
      appType: "ask",
      title: "Where should this go?",
      options: ["Safari"],
    });

    expect(config.appType).toBe("ask");
    expect(config.title).toBe("Where should this go?");
  });

  it("normalizes object options, autodetecting the app type", () => {
    const config = createBrowserConfig({
      name: "ask",
      appType: "ask",
      options: [
        { name: "Google Chrome", profile: "Profile 2", title: "Work" },
        { name: "com.google.Chrome", profile: "Default" },
      ],
    });

    expect(config.options).toEqual([
      {
        name: "Google Chrome",
        appType: "appName",
        openInBackground: undefined,
        profile: "Profile 2",
        args: [],
        title: "Work",
      },
      {
        name: "com.google.Chrome",
        appType: "bundleId",
        openInBackground: undefined,
        profile: "Default",
        args: [],
        title: "",
      },
    ]);
  });

  it("accepts the name:profile string shorthand for an option", () => {
    const config = createBrowserConfig({
      name: "ask",
      appType: "ask",
      options: ["Google Chrome:Profile 2"],
    });

    expect(config.options).toEqual([
      {
        name: "Google Chrome",
        appType: "appName",
        openInBackground: undefined,
        profile: "Profile 2",
        args: [],
        title: "",
      },
    ]);
  });

  it("leaves an explicit option app type alone", () => {
    const config = createBrowserConfig({
      name: "ask",
      appType: "ask",
      options: [{ name: "Google Chrome", appType: "path" }],
    });

    expect(config.options[0].appType).toBe("path");
  });
});
