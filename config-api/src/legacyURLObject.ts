/**
 * Legacy FinickyURL type used for backward compatibility
 */
export type LegacyURLObject = {
  username?: string;
  host: string;
  protocol?: string;
  pathname?: string;
  search?: string;
  password?: string;
  port?: number;
  hash?: string;
};

/**
 * Converts a LegacyURLObject to a URL string
 * @param urlObj The LegacyURLObject to convert
 * @returns A URL string representation
 */
export function legacyURLObjectToString(urlObj: LegacyURLObject): string {
  return `${urlObj.protocol ? urlObj.protocol.replace(":", "") : "https"}://${
    urlObj.username
  }${urlObj.password ? ":" + urlObj.password : ""}${
    urlObj.username || urlObj.password ? "@" : ""
  }${urlObj.host}${urlObj.port ? ":" + urlObj.port : ""}${urlObj.pathname}${
    urlObj.search ? "?" + urlObj.search : ""
  }${urlObj.hash ? "#" + urlObj.hash : ""}`;
}

/**
 * Converts a standard URL object to a FinickyURL format
 * @param url The URL object to convert
 * @returns A FinickyURL object
 */
export function URLtoLegacyURLObject(url: URL): LegacyURLObject {
  return {
    username: url.username,
    host: url.hostname,
    protocol: url.protocol.replace(":", ""),
    pathname: url.pathname,
    search: url.search.replace("?", ""),
    password: url.password,
    port: url.port ? parseInt(url.port, 10) : undefined,
    hash: url.hash.replace("#", ""),
  };
}

/**
 * Type guard to check if an object is a LegacyURLObject
 * @param obj The object to check
 * @returns True if the object matches LegacyURLObject structure
 */
export function isLegacyURLObject(obj: unknown): obj is LegacyURLObject {
  if (!obj || typeof obj !== "object") return false;

  const urlObj = obj as Partial<LegacyURLObject>;

  return (
    (typeof urlObj.username === "string" || urlObj.username === undefined) &&
    typeof urlObj.host === "string" &&
    (typeof urlObj.protocol === "string" || urlObj.protocol === undefined) &&
    (typeof urlObj.pathname === "string" || urlObj.pathname === undefined) &&
    (typeof urlObj.search === "string" || urlObj.search === undefined) &&
    (typeof urlObj.password === "string" || urlObj.password === undefined) &&
    (typeof urlObj.port === "number" || urlObj.port === undefined) &&
    (typeof urlObj.hash === "string" || urlObj.hash === undefined)
  );
}
