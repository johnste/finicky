(function (global, factory) {
  typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports) :
  typeof define === 'function' && define.amd ? define(['exports'], factory) :
  (global = global || self, factory(global.fastidious = {}));
}(this, function (exports) { 'use strict';

  function isDefined(value) {
      return typeof value !== "undefined" && value !== null;
  }
  function formatValue(value) {
      if (value instanceof RegExp) {
          return value.toString();
      }
      return JSON.stringify(value);
  }
  function getKeys(object) {
      return Object.keys(object).filter(key => Object.prototype.hasOwnProperty.call(object, key));
  }

  function createValidator(typeName, typeCallback) {
      function isOptional(value, key) {
          if (!isDefined(value)) {
              return undefined;
          }
          const result = typeCallback(value, key);
          if (typeof result === "boolean" && !result) {
              return `Value at ${key}: ${formatValue(value)} is not ${typeName}`;
          }
          else if (Array.isArray(result) && result.length > 0) {
              return result.join(", ");
          }
      }
      function isRequired(value, key) {
          if (!isDefined(value)) {
              return `Expected "${key}" to be ${typeName}`;
          }
          return isOptional(value, key);
      }
      function checkType(value, key) {
          return isOptional(value, key);
      }
      checkType.isRequired = isRequired;
      // Save typeName for nice error messages
      checkType.typeName = typeName;
      return checkType;
  }

  function getErrors(object, schema, prefix = "root.") {
      const schemaKeys = getKeys(schema);
      const errors = [];
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
              result = `Expected a validator at key ${prefix + key}`;
          }
          if (typeof result === "string") {
              errors.push(result);
          }
      });
      // Check for extraneous properties in object
      getKeys(object).forEach(key => {
          if (!schemaKeys.includes(key)) {
              errors.push("extraneous key " + prefix + key);
          }
      });
      return errors;
  }
  const validate = {
      boolean: createValidator("boolean", value => typeof value === "boolean"),
      string: createValidator("string", value => typeof value === "string"),
      number: createValidator("number", value => typeof value === "number" && !Number.isNaN(value)),
      function: createValidator("function", value => typeof value === "function"),
      regex: createValidator("regex", value => value instanceof RegExp),
      value: (expectedValue) => createValidator(expectedValue, value => {
          return value === expectedValue;
      }),
      shape: (schema) => createValidator("shape", (value, key) => {
          if (typeof value !== "object") {
              return false;
          }
          return getErrors(value, schema, key + ".");
      }),
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
          const description = typeCheckers.map(oneOf => oneOf.typeName);
          return createValidator(`oneOf: [${description}]`, (value, key) => {
              const errors = typeCheckers.every(oneOfValidator => typeof oneOfValidator(value, key) === "string");
              return errors ? [`${key}: Value not one of ${description}`] : true;
          });
      }
  };

  exports.getErrors = getErrors;
  exports.validate = validate;

  Object.defineProperty(exports, '__esModule', { value: true });

}));
