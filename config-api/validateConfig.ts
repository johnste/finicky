import { FinickyConfig, finickyConfigSchema } from "./types";
import { getErrors } from "./fastidious/index";

declare const module:
  | {
      exports?: FinickyConfig;
    }
  | undefined;

export function validateConfig() {
  if (!module) {
    throw new Error("module is not defined");
  }

  if (!module.exports) {
    throw new Error("module.exports is not defined");
  }

  const invalid = getErrors(
    module.exports,
    finickyConfigSchema,
    "module.exports."
  );

  if (invalid.length === 0) {
    return true;
  }

  throw new Error(invalid.join("\n"));
}
