import { UrlObject, AppType, Options } from "./types";
import { ISchema, IValidator } from "./fastidious/types";
import { getErrors } from "./fastidious/index";

export function createRegularExpression(pattern: string) {
  if (!pattern) {
    return /^$/;
  }

  let result = pattern;

  result = result.replace(/[-[\]\/{}()*+?.,\\^$|#\s]/g, "\\$&");
  result = result.replace(/\\\*/g, ".*");

  if (!pattern.startsWith("http://") && !pattern.startsWith("https://")) {
    result = "https?:\\/\\/" + result;
  }

  return new RegExp("^" + result + "$", "i");
}

export function guessAppType(value: string): AppType {
  if (value === null) {
    return "none";
  }

  if (looksLikeBundleIdentifier(value)) {
    return "bundleId";
  }

  if (looksLikeAbsolutePath(value)) {
    return "appPath";
  }

  return "appName";
}

function looksLikeBundleIdentifier(value: string) {
  // Regular expression to match Uniform Type Identifiers, "Bundle Identifiers"
  // Adapted from https://stackoverflow.com/a/34241710/1698327
  const bundleIdRegex = /^[A-Za-z]{2,6}((?!-)\.[A-Za-z0-9-]{1,63})+$/;
  if (bundleIdRegex.test(value)) {
    return true;
  }
  return false;
}

function looksLikeAbsolutePath(value: string) {
  return value.startsWith("/") || value.startsWith("~");
}

/**
 * Compose a url from a Url Object
 */
export function composeUrl(url: UrlObject) {
  const { protocol, host, pathname = "" } = url;
  let port = url.port ? `:${url.port}` : "";
  let search = url.search ? `?${url.search}` : "";
  let hash = url.hash ? `#${url.hash}` : "";
  let auth = url.username ? `${url.username}` : "";
  auth += url.password ? `:${url.password}` : "";

  return `${protocol}://${auth}${host}${port}${pathname}${search}${hash}`;
}

/**
 * Validate fastidious schema
 */
export function validateSchema(
  value: unknown,
  schema: ISchema | IValidator,
  path = ""
) {
  const errors = getErrors(value, schema, path);
  if (errors.length > 0) {
    throw new Error(
      errors.join("\n") + "\nReceived value: " + JSON.stringify(value, null, 2)
    );
  }
}

/**
 * Create a proxy to warn on accessing deprecated options
 */
export function deprecate<T extends Object>(
  target: T,
  deprecated: Map<keyof T, string>
): T {
  const handler = {
    get: function (target: T, prop: keyof T, receiver?: any) {
      if (deprecated.has(prop)) {
        // @ts-ignore
        finicky.log("⚠️", prop, "is deprecated: ", deprecated.get(prop));
      }

      return Reflect.get(target, prop, receiver);
    },
  };

  return new Proxy(target, handler);
}
