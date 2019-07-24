import { isDefined, formatValue } from "./utils";
import { TypeCallback, ICheckType, NameType } from "./types";

export function getTypeName(typeName: NameType) {
  if (typeof typeName === "string") {
    return typeName;
  }
  return JSON.stringify(typeName, null, 2);
}

export function createValidator(typeName: NameType, typeCallback: TypeCallback): ICheckType {
  function isOptional(value: any, key: string) {
    if (!isDefined(value)) {
      return undefined;
    }

    const result = typeCallback(value, key);
    if (typeof result === "boolean" && !result) {
      return `Value at ${key}: ${formatValue(value)} is not ${getTypeName(typeName)}`;
    } else if (Array.isArray(result) && result.length > 0) {
      return result.join("\n");
    }
  }

  function isRequired(value: any, key: string) {
    if (!isDefined(value)) {
      return `Expected "${key}" to be ${getTypeName(typeName)}`;
    }
    return isOptional(value, key);
  }

  isRequired.typeName = typeName;

  function checkType(value: any, key: string) {
    return isOptional(value, key);
  }

  checkType.isRequired = isRequired;
  // Save typeName for nice error messages
  checkType.typeName = typeName;

  return checkType;
}
