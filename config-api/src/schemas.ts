import { validate } from "./fastidious/index";

/**
 * Finicky Configuration Schemas – used to validate the configuration
 */
const urlObjectSchema = {
  protocol: validate.string.isRequired,
  username: validate.string,
  password: validate.string,
  host: validate.string.isRequired,
  port: validate.oneOf([validate.number, validate.value(null)]),
  pathname: validate.string,
  search: validate.string,
  hash: validate.string,
};

const partialUrlSchema = {
  ...urlObjectSchema,
  protocol: validate.string,
  host: validate.string,
};

export const urlSchema = {
  url: validate.oneOf([validate.string, validate.shape(urlObjectSchema)])
    .isRequired,
};

const browserSchema = validate.oneOf([
  validate.string,
  validate.shape({
    name: validate.string.isRequired,
    appType: validate.oneOf(["appName", "appPath", "bundleId"]),
    openInBackground: validate.boolean,
    profile: validate.string,
    args: validate.arrayOf(validate.string),
  }),
  validate.function("options"),
  validate.value(null),
]);

const multipleBrowsersSchema = validate.oneOf([
  browserSchema,
  validate.arrayOf(browserSchema.isRequired),
]);

const matchSchema = validate.oneOf([
  validate.string,
  validate.function("options"),
  validate.regex,
  validate.arrayOf(
    validate.oneOf([
      validate.string,
      validate.function("options"),
      validate.regex,
    ])
  ),
]);

export const finickyConfigSchema = {
  defaultBrowser: multipleBrowsersSchema.isRequired,
  options: validate.shape({
    hideIcon: validate.boolean,
    urlShorteners: validate.oneOf([
      validate.arrayOf(validate.string),
      validate.function("hostnames"),
    ]),
    checkForUpdate: validate.boolean,
    logRequests: validate.boolean,
  }),
  rewrite: validate.arrayOf(
    validate.shape({
      match: matchSchema.isRequired,
      url: validate.oneOf([
        validate.string,
        validate.shape(partialUrlSchema),
        validate.function("options"),
      ]).isRequired,
    }).isRequired
  ),
  handlers: validate.arrayOf(
    validate.shape({
      match: matchSchema.isRequired,
      url: validate.oneOf([
        validate.string,
        validate.shape(partialUrlSchema),
        validate.function("options"),
      ]),
      browser: multipleBrowsersSchema.isRequired,
    })
  ),
};

export const appDescriptorSchema = {
  name: validate.string,
  appType: validate.oneOf([
    validate.value("bundleId"),
    validate.value("appName"),
    validate.value("appPath"),
    validate.value("none"),
  ]).isRequired,
  openInBackground: validate.boolean,
  profile: validate.string,
  args: validate.arrayOf(validate.string),
};
