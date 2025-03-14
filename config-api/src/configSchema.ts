import { z } from "zod";

const NativeUrlSchema = z
  .instanceof(URL)
  .overrideTypeNode((ts) =>
    ts.factory.createTypeReferenceNode(ts.factory.createIdentifier("URL"))
  );

const RegexpSchema = z
  .instanceof(RegExp)
  .overrideTypeNode((ts) =>
    ts.factory.createTypeReferenceNode(ts.factory.createIdentifier("RegExp"))
  );

// ===== Process & URL Options =====
const ProcessInfoSchema = z
  .object({
    name: z.string(),
    bundleId: z.string(),
    path: z.string(),
  })
  .identifier("ProcessInfo");

export type ProcessInfo = z.infer<typeof ProcessInfoSchema>;

const OpenUrlOptionsSchema = z
  .object({
    opener: ProcessInfoSchema.nullable(),
  })
  .identifier("OpenUrlOptions");

export type OpenUrlOptions = z.infer<typeof OpenUrlOptionsSchema>;

// ===== URL Schemas =====

const UrlTransformerSchema = z
  .function(z.tuple([NativeUrlSchema, OpenUrlOptionsSchema]))
  .returns(z.union([z.string(), NativeUrlSchema]))
  .identifier("UrlTransformer");

const UrlTransformSpecificationSchema = z
  .union([z.string(), NativeUrlSchema, UrlTransformerSchema])
  .identifier("UrlTransformSpecification");

// ===== Matcher Schemas =====
const UrlMatcherFunctionSchema = z
  .function(z.tuple([NativeUrlSchema, OpenUrlOptionsSchema]))
  .returns(z.boolean())
  .identifier("UrlMatcherFunction");

const UrlMatcherSchema = z
  .union([z.string(), RegexpSchema, UrlMatcherFunctionSchema])
  .identifier("UrlMatcher");

const UrlMatcherPatternSchema = z.union([
  UrlMatcherSchema,
  z.array(UrlMatcherSchema),
]);

// ===== Browser Schemas =====

const appTypes = ["appName", "bundleId", "path", "none"] as const;

export type AppType = (typeof appTypes)[number];

const BrowserConfigSchema = z
  .object({
    name: z.string(),
    appType: z.enum(appTypes).optional(),
    openInBackground: z.boolean().optional(),
    profile: z.string().optional(),
    args: z.array(z.string()).optional(),
  })
  .identifier("BrowserConfig");

export const BrowserConfigStrictSchema = z.object({
  name: z.string(),
  appType: z.enum(appTypes),
  openInBackground: z.boolean(),
  profile: z.string(),
  args: z.array(z.string()),
  url: z.string(),
});

const BrowserResolverSchema = z
  .function(z.tuple([NativeUrlSchema, OpenUrlOptionsSchema]))
  .returns(z.union([z.string(), BrowserConfigSchema]))
  .identifier("BrowserResolver");

export const BrowserSpecificationSchema = z
  .union([z.null(), z.string(), BrowserConfigSchema, BrowserResolverSchema])
  .identifier("BrowserSpecification");

// ===== Rule Schemas =====
const RewriteRuleSchema = z.object({
  match: UrlMatcherPatternSchema,
  url: UrlTransformSpecificationSchema,
});

const HandlerRuleSchema = z.object({
  match: UrlMatcherPatternSchema,
  browser: BrowserSpecificationSchema,
});

// ===== Configuration Schemas =====
const ConfigOptionsSchema = z
  .object({
    urlShorteners: z.array(z.string()).optional(),
    logRequests: z.boolean().optional(),
    checkForUpdates: z.boolean().optional(),
  })
  .identifier("ConfigOptions");

/**
 * @internal - don't export this schema as a type
 */
export const ConfigSchema = z.object({
  defaultBrowser: BrowserSpecificationSchema,
  options: ConfigOptionsSchema.optional(),
  rewrite: z.array(RewriteRuleSchema).optional(),
  handlers: z.array(HandlerRuleSchema).optional(),
});

export type UrlTransformSpecification = z.infer<
  typeof UrlTransformSpecificationSchema
>;
export type UrlTransformer = z.infer<typeof UrlTransformerSchema>;

export type UrlMatcherFunction = z.infer<typeof UrlMatcherFunctionSchema>;
export type UrlMatcher = z.infer<typeof UrlMatcherSchema>;
export type UrlMatcherPattern = z.infer<typeof UrlMatcherPatternSchema>;
export type BrowserConfigStrict = z.infer<typeof BrowserConfigStrictSchema>;

export type BrowserConfig = z.infer<typeof BrowserConfigSchema>;
export type BrowserResolver = z.infer<typeof BrowserResolverSchema>;
export type BrowserSpecification = z.infer<typeof BrowserSpecificationSchema>;

export type RewriteRule = z.infer<typeof RewriteRuleSchema>;
export type HandlerRule = z.infer<typeof HandlerRuleSchema>;

export type ConfigOptions = z.infer<typeof ConfigOptionsSchema>;
export type Config = z.infer<typeof ConfigSchema>;
