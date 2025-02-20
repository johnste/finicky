import { z } from "zod";

export function generateConfigSchema(
  NativeUrlSchema: z.ZodTypeAny,
  RegexpSchema: z.ZodTypeAny
) {
  // ===== Process & URL Options =====
  const ProcessInfoSchema = z.object({
    name: z.string(),
    bundleId: z.string(),
    path: z.string(),
  });

  const OpenUrlOptionsSchema = z.object({
    opener: z.union([ProcessInfoSchema, z.null()]).optional(),
  });

  // ===== URL Schemas =====

  const UrlTransformFnSchema = z
    .function()
    .args(NativeUrlSchema, OpenUrlOptionsSchema)
    .returns(z.union([z.string(), NativeUrlSchema]));

  const UrlPatternSchema = z.union([
    z.string(),
    RegexpSchema,
    UrlTransformFnSchema,
  ]);

  // ===== Matcher Schemas =====
  const MatcherFnSchema = z
    .function()
    .args(NativeUrlSchema, OpenUrlOptionsSchema)
    .returns(z.boolean());

  const UrlMatcherSchema = z.union([z.string(), RegexpSchema, MatcherFnSchema]);

  const UrlMatcherPatternSchema = z.union([
    UrlMatcherSchema,
    z.array(UrlMatcherSchema),
  ]);

  // ===== Browser Schemas =====

  const appTypes = ["appName", "bundleId", "path", "none"] as const;

  const BrowserConfigSchema = z.object({
    name: z.string(),
    appType: z.enum(appTypes).optional(),
    openInBackground: z.boolean().optional(),
    profile: z.string().optional(),
    args: z.array(z.string()).optional(),
  });

  const BrowserConfigStrictSchema = BrowserConfigSchema.extend({
    appType: z.enum(appTypes),
    openInBackground: z.boolean(),
    profile: z.string(),
    args: z.array(z.string()),
    url: z.string(),
  });

  const BrowserResolverFnSchema = z
    .function()
    .args(NativeUrlSchema, OpenUrlOptionsSchema)
    .returns(z.union([z.string(), BrowserConfigSchema]));

  const BrowserPatternSchema = z.union([
    z.null(),
    z.string(),
    BrowserConfigSchema,
    BrowserResolverFnSchema,
  ]);

  // ===== Rule Schemas =====
  const RewriteRuleSchema = z.object({
    match: UrlMatcherPatternSchema,
    url: UrlPatternSchema,
  });

  const HandlerRuleSchema = z.object({
    match: UrlMatcherPatternSchema,
    browser: BrowserPatternSchema,
  });

  // ===== Configuration Schemas =====
  const ConfigOptionsSchema = z
    .object({
      urlShorteners: z.array(z.string()).optional(),
      logRequests: z.boolean().optional(),
      checkForUpdates: z.boolean().optional(),
    })
    .optional();

  /**
   * @internal - don't export this schema as a type
   */
  const ConfigSchema = z.object({
    defaultBrowser: BrowserPatternSchema,
    options: ConfigOptionsSchema,
    rewrite: z.array(RewriteRuleSchema).optional(),
    handlers: z.array(HandlerRuleSchema).optional(),
  });

  const SimpleConfigSchema = z.record(z.string(), z.string());

  return {
    appTypes,
    ConfigSchema,
    SimpleConfigSchema,
    ConfigOptionsSchema,
    RewriteRuleSchema,
    HandlerRuleSchema,
    BrowserConfigSchema,
    BrowserConfigStrictSchema,
    BrowserResolverFnSchema,
    BrowserPatternSchema,
    UrlMatcherPatternSchema,
    UrlPatternSchema,
    UrlTransformFnSchema,
    UrlMatcherSchema,
    MatcherFnSchema,
    OpenUrlOptionsSchema,
    ProcessInfoSchema,
  };
}
