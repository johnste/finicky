import { FinickyConfig, finickyConfigSchema } from "./types";
import { getErrors } from "./fastidious/index";

export function validateConfig(config: FinickyConfig) {
  const invalid = getErrors(config, finickyConfigSchema, "module.exports.");

  if (invalid.length === 0) {
    return true;
  }

  throw new Error(invalid.join("\n"));
}
