(function (global, factory) {
  typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports) :
  typeof define === 'function' && define.amd ? define(['exports'], factory) :
  (global = global || self, factory(global.fastidious = {}));
}(this, function (exports) { 'use strict';

  /* fastidious 1.0.12 - https://github.com/johnste/fastidious */

  function isDefined(value) {
      return typeof value !== "undefined";
  }
  function formatValue(value) {
      if (value instanceof RegExp) {
          return value.toString();
      }
      else if (Array.isArray(value)) {
          return `[Array]`;
      }
      else if (typeof value === "function") {
          return `[Function${value.name ? ` ${value.name}` : ""}]`;
      }
      else if (value instanceof Date) {
          return `[Date]`;
      }
      else if (value === null) {
          return "[null]";
      }
      else if (value === undefined) {
          return "[undefined]";
      }
      return `[${JSON.stringify(value, null, 2)}]`;
  }
  function getKeys(object) {
      return Object.keys(object).filter(key => Object.prototype.hasOwnProperty.call(object, key));
  }
  function enumerate(names, mode = "or") {
      if (names.length === 0) {
          return "";
      }
      if (names.length == 1) {
          return names[0];
      }
      const [tail, ...body] = names.reverse();
      return `${body.join(", ")} ${mode} ${tail}`;
  }

  function getTypeName(typeName) {
      if (typeof typeName === "string") {
          return typeName;
      }
      return JSON.stringify(typeName, null, 2);
  }
  function createValidator(typeName, typeCallback) {
      function isOptional(value, key) {
          if (!isDefined(value)) {
              return undefined;
          }
          const result = typeCallback(value, key);
          if (typeof result === "boolean" && !result) {
              return `Value at ${key}: ${formatValue(value)} is not ${getTypeName(typeName)}`;
          }
          else if (Array.isArray(result) && result.length > 0) {
              return result.join("\n");
          }
      }
      function isRequired(value, key) {
          if (!isDefined(value)) {
              return `Expected "${key}" to be ${getTypeName(typeName)}`;
          }
          return isOptional(value, key);
      }
      isRequired.typeName = typeName;
      function checkType(value, key) {
          return isOptional(value, key);
      }
      checkType.isRequired = isRequired;
      // Save typeName for nice error messages
      checkType.typeName = typeName;
      return checkType;
  }

  function getErrors(object, schema, prefix = "root.") {
      // If schema is a function we're testing a single validator
      if (typeof schema === "function") {
          const result = schema(object, prefix + "value");
          return result ? [result] : [];
      }
      else if (typeof schema !== "object") {
          return [
              `Expected an schema that was an object or a function, but received ${typeof object} (path: ${prefix})`
          ];
      }
      const schemaKeys = getKeys(schema);
      const errors = [];
      if (typeof object !== "object" || object === null) {
          errors.push(`Expected an object to validate, but received ${typeof object} (path: ${prefix})`);
      }
      else {
          // Validate each property in schema
          schemaKeys.forEach(key => {
              const propChecker = schema[key];
              let result;
              if (typeof propChecker === "function") {
                  result = propChecker(object[key], prefix + key);
              }
              else if (["string", "number"].includes(typeof propChecker)) {
                  result = validate.value(propChecker)(object[key], prefix + key);
              }
              else {
                  result = `Expected a validator at path ${prefix + key}`;
              }
              if (typeof result === "string") {
                  errors.push(result);
              }
          });
          // Check for extraneous properties in object
          getKeys(object).forEach(key => {
              if (!schemaKeys.includes(key)) {
                  errors.push(`unknown key ${key} at ${prefix + key}`);
              }
          });
      }
      return errors;
  }
  const validate = {
      boolean: createValidator("boolean", value => typeof value === "boolean"),
      string: createValidator("string", value => typeof value === "string"),
      number: createValidator("number", value => typeof value === "number" && !Number.isNaN(value)),
      function: (argNames) => {
          if (!Array.isArray(argNames)) {
              if (argNames) {
                  argNames = [argNames];
              }
              else {
                  argNames = [];
              }
          }
          const name = `function(${argNames.join(", ")})`;
          return createValidator(name, value => typeof value === "function");
      },
      regex: createValidator("regex", value => value instanceof RegExp),
      value: (expectedValue) => createValidator(expectedValue, value => {
          return value === expectedValue;
      }),
      shape: (schema) => {
          const names = getNameType(schema);
          return createValidator(names, (value, key) => {
              if (typeof value !== "object" || value === null) {
                  return false;
              }
              return getErrors(value, schema, key + ".");
          });
      },
      arrayOf: (validator) => createValidator("array", (value, key) => {
          if (!Array.isArray(value)) {
              return false;
          }
          return value.reduce((errors, item, index) => {
              const result = validator(item, `${key}[${index}]`);
              if (typeof result === "string") {
                  return [...errors, result];
              }
              return errors;
          }, []);
      }),
      oneOf: (OneOfs) => {
          const typeCheckers = OneOfs.map(v => {
              if (["string", "number"].includes(typeof v)) {
                  return validate.value(v);
              }
              return v;
          });
          const description = enumerate(typeCheckers.map(oneOf => getTypeName(oneOf.typeName)));
          return createValidator(`${description}`, (value, key) => {
              const errors = typeCheckers.every(oneOfValidator => typeof oneOfValidator(value, key) === "string");
              return errors ? [`${key}: Value not one of ${description}`] : true;
          });
      }
  };
  function getNameType(schema) {
      const names = {};
      const schemaKeys = getKeys(schema);
      schemaKeys.forEach(key => {
          const property = schema[key];
          if (typeof property === "number" || typeof property === "string") {
              names[key] = typeof property;
          }
          else {
              names[key] = property.typeName;
          }
      });
      return names;
  }

  exports.getErrors = getErrors;
  exports.validate = validate;

  Object.defineProperty(exports, '__esModule', { value: true });

}));
