import { validateConfig } from "../src/validateConfig";
import { validate } from "@babel/types";

function makeValidate(config: any) {
  return function validate() {
    return validateConfig(config);
  };
}

describe("test", () => {
  test("empty config", () => {
    const config = {};
    expect(makeValidate(config)).toThrow();
  });

  test("invalid config", () => {
    const config = "!! invalid config !!";
    expect(makeValidate(config)).toThrow();
  });

  test("valid minimal config", () => {
    const config = {
      defaultBrowser: "Browser"
    };

    expect(validateConfig(config)).toBe(true);
  });

  test("invalid minimal config", () => {
    const config = {
      defaultBrowser: "Browser",
      invalidKey: "This key is invalid"
    };

    expect(makeValidate(config)).toThrow();
  });

  test("invalid minimal config", () => {
    const config = {
      defaultBrowser: "Browser",
      defaultBrowserss: "Browser"
    };

    expect(makeValidate(config)).toThrow();
  });
});
