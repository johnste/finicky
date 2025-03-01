import { z } from "zod";
import { generateConfigSchema } from "./generateSchema";

export const NativeUrlSchema = z.instanceof(URL);
export const RegexpSchema = z.instanceof(RegExp);

const {
  ProcessInfoSchema,
  OpenUrlOptionsSchema,
  UrlPatternSchema,
  UrlTransformFnSchema,
  MatcherFnSchema,
  UrlMatcherSchema,
  UrlMatcherPatternSchema,
  BrowserConfigStrictSchema,
  BrowserConfigSchema,
  BrowserResolverFnSchema,
  BrowserPatternSchema,
  RewriteRuleSchema,
  HandlerRuleSchema,
  ConfigOptionsSchema,
  ConfigSchema,
  SimpleConfigSchema,
  appTypes,
} = generateConfigSchema(NativeUrlSchema, RegexpSchema);

export {
  ConfigSchema,
  SimpleConfigSchema,
  BrowserPatternSchema,
  BrowserConfigStrictSchema,
};

export type AppType = (typeof appTypes)[number];

export type ProcessInfo = z.infer<typeof ProcessInfoSchema>;
export type OpenUrlOptions = z.infer<typeof OpenUrlOptionsSchema>;

export type UrlPattern = z.infer<typeof UrlPatternSchema>;
export type UrlTransformFn = z.infer<typeof UrlTransformFnSchema>;

export type MatcherFn = z.infer<typeof MatcherFnSchema>;
export type UrlMatcher = z.infer<typeof UrlMatcherSchema>;
export type UrlMatcherPattern = z.infer<typeof UrlMatcherPatternSchema>;
export type BrowserConfigStrict = z.infer<typeof BrowserConfigStrictSchema>;

export type BrowserConfig = z.infer<typeof BrowserConfigSchema>;
export type BrowserResolverFn = z.infer<typeof BrowserResolverFnSchema>;
export type BrowserPattern = z.infer<typeof BrowserPatternSchema>;

export type RewriteRule = z.infer<typeof RewriteRuleSchema>;
export type HandlerRule = z.infer<typeof HandlerRuleSchema>;

export type ConfigOptions = z.infer<typeof ConfigOptionsSchema>;
export type Config = z.infer<typeof ConfigSchema>;
export type SimpleConfig = z.infer<typeof SimpleConfigSchema>;
