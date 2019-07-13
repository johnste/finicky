export function isDefined(value: any) {
  return typeof value !== "undefined";
}

export function formatValue(value: any): string {
  if (value instanceof RegExp) {
    return value.toString();
  } else if (Array.isArray(value)) {
    return `[Array]`;
  } else if (typeof value === "function") {
    return `[Function${value.name ? ` ${value.name}` : ""}]`;
  } else if (value instanceof Date) {
    return `[Date]`;
  } else if (value === null) {
    return "[null]";
  } else if (value === undefined) {
    return "[undefined]";
  }

  return `[${JSON.stringify(value, null, 2)}]`;
}

export function getKeys(object: object) {
  return Object.keys(object).filter(key => Object.prototype.hasOwnProperty.call(object, key));
}

export function enumerate(names: string[], mode: "or" | "and" = "or") {
  if (names.length === 0) {
    return "";
  }

  if (names.length == 1) {
    return names[0];
  }

  const [tail, ...body] = names.reverse();
  return `${body.join(", ")} ${mode} ${tail}`;
}
