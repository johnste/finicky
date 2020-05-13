let finickyConfigApi = (() => {
  let __assign = Object.assign;
  let __defineProperty = Object.defineProperty;
  let __hasOwnProperty = Object.hasOwnProperty;
  let __modules = {};
  let __commonjs;
  let __require = (id) => {
    let module = __modules[id];
    if (!module) {
      module = __modules[id] = {
        exports: {}
      };
      __commonjs[id](module.exports, module);
    }
    return module.exports;
  };
  let __toModule = (module) => {
    if (module && module.__esModule) {
      return module;
    }
    let result = {};
    for (let key in module) {
      if (__hasOwnProperty.call(module, key)) {
        result[key] = module[key];
      }
    }
    result.default = module;
    return result;
  };
  let __import = (id) => {
    return __toModule(__require(id));
  };
  let __export = (target, all) => {
    __defineProperty(target, "__esModule", {
      value: true
    });
    for (let name in all) {
      __defineProperty(target, name, {
        get: all[name],
        enumerable: true
      });
    }
  };
  __commonjs = {
    8(index2) {
      // src/fastidious/utils.ts
      function isDefined(value) {
        return typeof value !== "undefined";
      }
      function formatValue(value) {
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
        } else if (value === void 0) {
          return "[undefined]";
        }
        return `[${JSON.stringify(value, null, 2)}]`;
      }
      function getKeys(object) {
        return Object.keys(object).filter((key) => Object.prototype.hasOwnProperty.call(object, key));
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

      // src/fastidious/createValidator.ts
      function getTypeName(typeName) {
        if (typeof typeName === "string") {
          return typeName;
        }
        return JSON.stringify(typeName, null, 2);
      }
      function createValidator2(typeName, typeCallback) {
        function isOptional(value, key) {
          if (!isDefined(value)) {
            return void 0;
          }
          const result = typeCallback(value, key);
          if (typeof result === "boolean" && !result) {
            return `Value at ${key}: ${formatValue(value)} is not ${getTypeName(typeName)}`;
          } else if (Array.isArray(result) && result.length > 0) {
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
        checkType.typeName = typeName;
        return checkType;
      }

      // src/fastidious/fastidious.ts
      function getErrors(object, schema, prefix = "root.") {
        if (typeof schema === "function") {
          const result = schema(object, prefix + "value");
          return result ? [result] : [];
        } else if (typeof schema !== "object") {
          return [`Expected an schema that was an object or a function, but received ${typeof object} (path: ${prefix})`];
        }
        const schemaKeys = getKeys(schema);
        const errors = [];
        if (typeof object !== "object" || object === null) {
          errors.push(`Expected an object to validate, but received ${typeof object} (path: ${prefix})`);
        } else {
          schemaKeys.forEach((key) => {
            const propChecker = schema[key];
            let result;
            if (typeof propChecker === "function") {
              result = propChecker(object[key], prefix + key);
            } else if (["string", "number"].includes(typeof propChecker)) {
              result = validate.value(propChecker)(object[key], prefix + key);
            } else {
              result = `Expected a validator at path ${prefix + key}`;
            }
            if (typeof result === "string") {
              errors.push(result);
            }
          });
          getKeys(object).forEach((key) => {
            if (!schemaKeys.includes(key)) {
              errors.push(`unknown key ${key} at ${prefix + key}`);
            }
          });
        }
        return errors;
      }
      const validate = {
        boolean: createValidator2("boolean", (value) => typeof value === "boolean"),
        string: createValidator2("string", (value) => typeof value === "string"),
        number: createValidator2("number", (value) => typeof value === "number" && !Number.isNaN(value)),
        function: (argNames) => {
          if (!Array.isArray(argNames)) {
            if (argNames) {
              argNames = [argNames];
            } else {
              argNames = [];
            }
          }
          const name = `function(${argNames.join(", ")})`;
          return createValidator2(name, (value) => typeof value === "function");
        },
        regex: createValidator2("regex", (value) => value instanceof RegExp),
        value: (expectedValue) => createValidator2(expectedValue, (value) => {
          return value === expectedValue;
        }),
        shape: (schema) => {
          const names = getNameType(schema);
          return createValidator2(names, (value, key) => {
            if (typeof value !== "object" || value === null) {
              return false;
            }
            return getErrors(value, schema, key + ".");
          });
        },
        arrayOf: (validator) => createValidator2("array", (value, key) => {
          if (!Array.isArray(value)) {
            return false;
          }
          return value.reduce((errors, item, index6) => {
            const result = validator(item, `${key}[${index6}]`);
            if (typeof result === "string") {
              return [...errors, result];
            }
            return errors;
          }, []);
        }),
        oneOf: (OneOfs) => {
          const typeCheckers = OneOfs.map((v) => {
            if (["string", "number"].includes(typeof v)) {
              return validate.value(v);
            }
            return v;
          });
          const description = enumerate(typeCheckers.map((oneOf) => getTypeName(oneOf.typeName)));
          return createValidator2(`${description}`, (value, key) => {
            const errors = typeCheckers.every((oneOfValidator) => typeof oneOfValidator(value, key) === "string");
            return errors ? [`${key}: Value not one of ${description}`] : true;
          });
        }
      };
      function getNameType(schema) {
        const names = {};
        const schemaKeys = getKeys(schema);
        schemaKeys.forEach((key) => {
          const property = schema[key];
          if (typeof property === "number" || typeof property === "string") {
            names[key] = typeof property;
          } else {
            names[key] = property.typeName;
          }
        });
        return names;
      }

      // src/fastidious/index.ts

      // src/types.ts
      const urlSchema = {
        url: validate.oneOf([validate.string, validate.shape({
          protocol: validate.string,
          username: validate.string,
          password: validate.string,
          host: validate.string,
          port: validate.oneOf([validate.number, validate.value(null)]),
          pathname: validate.string,
          search: validate.string,
          hash: validate.string
        })]).isRequired
      };
      const browserSchema = validate.oneOf([validate.string, validate.shape({
        name: validate.string.isRequired,
        appType: validate.oneOf(["appName", "appPath", "bundleId"]),
        openInBackground: validate.boolean
      }), validate.function("options"), validate.value(null)]);
      const multipleBrowsersSchema = validate.oneOf([browserSchema, validate.arrayOf(browserSchema.isRequired)]);
      const matchSchema = validate.oneOf([validate.string, validate.function("options"), validate.regex, validate.arrayOf(validate.oneOf([validate.string, validate.function("options"), validate.regex]))]);
      const finickyConfigSchema = {
        defaultBrowser: multipleBrowsersSchema.isRequired,
        options: validate.shape({
          hideIcon: validate.boolean,
          urlShorteners: validate.arrayOf(validate.string),
          checkForUpdate: validate.boolean
        }),
        rewrite: validate.arrayOf(validate.shape({
          match: matchSchema.isRequired,
          url: validate.oneOf([validate.string, validate.function("options")]).isRequired
        }).isRequired),
        handlers: validate.arrayOf(validate.shape({
          match: matchSchema.isRequired,
          url: validate.oneOf([validate.string, validate.function("options")]),
          browser: multipleBrowsersSchema.isRequired
        }))
      };

      // src/utils.ts
      function createRegularExpression(pattern) {
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

      // src/processUrl.ts
      const appDescriptorSchema = {
        name: validate.string,
        appType: validate.oneOf([validate.value("bundleId"), validate.value("appName"), validate.value("appPath"), validate.value("none")]).isRequired,
        openInBackground: validate.boolean
      };
      function processUrl3(config, url, processOptions) {
        if (!processOptions) {
          processOptions = {
            keys: {
              capsLock: false,
              command: false,
              shift: false,
              option: false,
              control: false,
              function: false
            }
          };
        }
        let options = __assign({
          urlString: url,
          url: finicky.getUrlParts(url)
        }, processOptions);
        if (!config) {
          return processBrowserResult("Safari", options);
        }
        options = processUrlRewrites(config, options);
        if (Array.isArray(config.handlers)) {
          for (let handler of config.handlers) {
            if (isMatch(handler.match, options)) {
              if (handler.url) {
                options = rewriteUrl(handler.url, options);
              }
              return processBrowserResult(handler.browser, options);
            }
          }
        }
        return processBrowserResult(config.defaultBrowser, options);
      }
      function validateSchema(value, schema, path = "") {
        const errors = getErrors(value, schema, path);
        if (errors.length > 0) {
          throw new Error(errors.join("\n") + "\nReceived value: " + JSON.stringify(value, null, 2));
        }
      }
      function createUrl(url) {
        const {protocol, host, pathname = ""} = url;
        let port = url.port ? `:${url.port}` : "";
        let search = url.search ? `?${url.search}` : "";
        let hash = url.hash ? `#${url.hash}` : "";
        let auth = url.username ? `${url.username}` : "";
        auth += url.password ? `:${url.password}` : "";
        return `${protocol}://${auth}${host}${port}${pathname}${search}${hash}`;
      }
      function processUrlRewrites(config, options) {
        if (Array.isArray(config.rewrite)) {
          for (let rewrite of config.rewrite) {
            if (isMatch(rewrite.match, options)) {
              options = rewriteUrl(rewrite.url, options);
            }
          }
        }
        return options;
      }
      function rewriteUrl(url, options) {
        let urlResult = resolveUrl(url, options);
        validateSchema({
          url: urlResult
        }, urlSchema);
        if (typeof urlResult === "string") {
          options = __assign(__assign({}, options), {
            url: finicky.getUrlParts(urlResult),
            urlString: urlResult
          });
        } else {
          options = __assign(__assign({}, options), {
            url: urlResult,
            urlString: createUrl(urlResult)
          });
        }
        return options;
      }
      function isMatch(matcher, options) {
        if (!matcher) {
          return false;
        }
        const matchers = Array.isArray(matcher) ? matcher : [matcher];
        return matchers.some((matcher2) => {
          if (matcher2 instanceof RegExp) {
            return matcher2.test(options.urlString);
          } else if (typeof matcher2 === "string") {
            const regex = createRegularExpression(matcher2);
            if (!regex) {
              return false;
            }
            return regex.test(options.urlString);
          } else if (typeof matcher2 === "function") {
            return !!matcher2(options);
          }
          return false;
        });
      }
      function resolveBrowser(result, options) {
        if (typeof result !== "function") {
          return result;
        }
        return result(options);
      }
      function resolveUrl(result, options) {
        if (typeof result === "string") {
          return result;
        } else if (typeof result === "object") {
          return __assign(__assign({}, options.url), result);
        }
        const resolved = result(options);
        if (typeof resolved === "string") {
          return resolved;
        }
        return __assign(__assign({}, options.url), resolved);
      }
      function getAppType(value) {
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
      function processBrowserResult(result, options) {
        let browser = resolveBrowser(result, options);
        if (!Array.isArray(browser)) {
          browser = [browser];
        }
        const browsers = browser.map(createBrowser);
        return {
          browsers,
          url: options.urlString
        };
      }
      function createBrowser(browser) {
        if (typeof browser === "string" || browser === null) {
          browser = {
            name: browser
          };
        }
        if (typeof browser === "object" && !browser.appType) {
          const name = browser.name === null ? "" : browser.name;
          browser = __assign(__assign({}, browser), {
            name,
            appType: getAppType(browser.name)
          });
        }
        validateSchema(browser, appDescriptorSchema);
        return browser;
      }
      function looksLikeBundleIdentifier(value) {
        const bundleIdRegex = /^[A-Za-z]{2,6}((?!-)\.[A-Za-z0-9-]{1,63})+$/;
        if (bundleIdRegex.test(value)) {
          return true;
        }
        return false;
      }
      function looksLikeAbsolutePath(value) {
        return value.startsWith("/") || value.startsWith("~");
      }

      // src/validateConfig.ts
      function validateConfig2(config) {
        const invalid = getErrors(config, finickyConfigSchema, "module.exports.");
        if (invalid.length === 0) {
          return true;
        }
        throw new Error(invalid.join("\n"));
      }

      // src/createAPI.ts
      const url_parse = __import(2 /* url-parse */);
      function createAPI2(options = {}) {
        const log = options.log || ((message) => {
          if (typeof finickyInternalAPI !== "undefined") {
            finickyInternalAPI.log(message);
          } else {
            console.log(`[finicky log] ${message}`);
          }
        });
        const notify = options.notify || ((title, subtitle) => {
          if (typeof finickyInternalAPI !== "undefined") {
            finickyInternalAPI.notify(title, subtitle);
          } else {
            console.log(`[finicky notify] ${title} ${subtitle}`);
          }
        });
        const getBattery = options.getBattery || (() => {
          if (typeof finickyInternalAPI !== "undefined") {
            let status = finickyInternalAPI.getBattery();
            return status;
          } else {
            return void 0;
          }
        });
        const getUrlParts = (urlString) => {
          const url = url_parse.default(urlString);
          const search = url.query;
          return {
            username: url.username,
            host: url.hostname,
            protocol: url.protocol.replace(":", ""),
            pathname: url.pathname,
            search: search.replace("?", ""),
            password: url.password,
            port: url.port ? +url.port : void 0,
            hash: url.hash.replace("#", "")
          };
        };
        const matchDomains = (matchers, ...args) => {
          if (args.length > 0) {
            throw new Error("finicky.matchDomains/matchHostnames only accepts one argument. See https://johnste.github.io/finicky-docs/interfaces/_finickyapi_.finicky.html#matchdomains for more information");
          }
          if (!Array.isArray(matchers)) {
            matchers = [matchers];
          }
          matchers.forEach((matcher) => {
            if (matcher instanceof RegExp || typeof matcher === "string") {
              return;
            }
            throw new Error(`finicky.matchDomains/matchHostnames: Unrecognized hostname "${matcher}"`);
          });
          return function({url}) {
            const domain = url.host;
            return matchers.some((matcher) => {
              if (matcher instanceof RegExp) {
                return matcher.test(domain);
              } else if (typeof matcher === "string") {
                return matcher === domain;
              }
              return false;
            });
          };
        };
        const onUrl = () => {
          log("finicky.onUrl is no longer supported in this version of Finicky, please go to https://github.com/johnste/finicky for updated documentation");
          notify("finicky.onUrl is no longer supported", "Check the Finicky website for updated documentation");
        };
        const setDefaultBrowser = () => {
          log("finicky.setDefaultBrowser is no longer supported in this version of Finicky, please go to https://github.com/johnste/finicky for updated documentation");
          notify("finicky.setDefaultBrowser is no longer supported", "Check the Finicky website for updated documentation");
        };
        return {
          log,
          notify,
          matchDomains,
          matchHostnames: matchDomains,
          getUrlParts,
          onUrl,
          setDefaultBrowser,
          getBattery
        };
      }

      // src/index.ts
      __export(index2, {
        createAPI: () => createAPI2,
        processUrl: () => processUrl3,
        validateConfig: () => validateConfig2
      });
    },

    1(exports, module) {
      // node_modules/requires-port/index.js
      "use strict";
      module.exports = function required(port, protocol) {
        protocol = protocol.split(":")[0];
        port = +port;
        if (!port)
          return false;
        switch (protocol) {
          case "http":
          case "ws":
            return port !== 80;
          case "https":
          case "wss":
            return port !== 443;
          case "ftp":
            return port !== 21;
          case "gopher":
            return port !== 70;
          case "file":
            return false;
        }
        return port !== 0;
      };
    },

    0(exports) {
      // node_modules/querystringify/index.js
      "use strict";
      var has = Object.prototype.hasOwnProperty, undef;
      function decode(input) {
        try {
          return decodeURIComponent(input.replace(/\+/g, " "));
        } catch (e) {
          return null;
        }
      }
      function encode(input) {
        try {
          return encodeURIComponent(input);
        } catch (e) {
          return null;
        }
      }
      function querystring(query) {
        var parser = /([^=?&]+)=?([^&]*)/g, result = {}, part;
        while (part = parser.exec(query)) {
          var key = decode(part[1]), value = decode(part[2]);
          if (key === null || value === null || key in result)
            continue;
          result[key] = value;
        }
        return result;
      }
      function querystringify(obj, prefix) {
        prefix = prefix || "";
        var pairs = [], value, key;
        if ("string" !== typeof prefix)
          prefix = "?";
        for (key in obj) {
          if (has.call(obj, key)) {
            value = obj[key];
            if (!value && (value === null || value === undef || isNaN(value))) {
              value = "";
            }
            key = encodeURIComponent(key);
            value = encodeURIComponent(value);
            if (key === null || value === null)
              continue;
            pairs.push(key + "=" + value);
          }
        }
        return pairs.length ? prefix + pairs.join("&") : "";
      }
      exports.stringify = querystringify;
      exports.parse = querystring;
    },

    2(exports, module) {
      // node_modules/url-parse/index.js
      "use strict";
      var required = __require(1 /* requires-port */), qs = __require(0 /* querystringify */), slashes = /^[A-Za-z][A-Za-z0-9+-.]*:\/\//, protocolre = /^([a-z][a-z0-9.+-]*:)?(\/\/)?([\S\s]*)/i, whitespace = "[\\x09\\x0A\\x0B\\x0C\\x0D\\x20\\xA0\\u1680\\u180E\\u2000\\u2001\\u2002\\u2003\\u2004\\u2005\\u2006\\u2007\\u2008\\u2009\\u200A\\u202F\\u205F\\u3000\\u2028\\u2029\\uFEFF]", left = new RegExp("^" + whitespace + "+");
      function trimLeft(str) {
        return (str ? str : "").toString().replace(left, "");
      }
      var rules = [["#", "hash"], ["?", "query"], function sanitize(address) {
        return address.replace("\\", "/");
      }, ["/", "pathname"], ["@", "auth", 1], [NaN, "host", void 0, 1, 1], [/:(\d+)$/, "port", void 0, 1], [NaN, "hostname", void 0, 1, 1]];
      var ignore = {
        hash: 1,
        query: 1
      };
      function lolcation(loc) {
        var globalVar;
        if (typeof window !== "undefined")
          globalVar = window;
        else if (typeof global !== "undefined")
          globalVar = global;
        else if (typeof self !== "undefined")
          globalVar = self;
        else
          globalVar = {};
        var location = globalVar.location || {};
        loc = loc || location;
        var finaldestination = {}, type = typeof loc, key;
        if ("blob:" === loc.protocol) {
          finaldestination = new Url(unescape(loc.pathname), {});
        } else if ("string" === type) {
          finaldestination = new Url(loc, {});
          for (key in ignore)
            delete finaldestination[key];
        } else if ("object" === type) {
          for (key in loc) {
            if (key in ignore)
              continue;
            finaldestination[key] = loc[key];
          }
          if (finaldestination.slashes === void 0) {
            finaldestination.slashes = slashes.test(loc.href);
          }
        }
        return finaldestination;
      }
      function extractProtocol(address) {
        address = trimLeft(address);
        var match = protocolre.exec(address);
        return {
          protocol: match[1] ? match[1].toLowerCase() : "",
          slashes: !!match[2],
          rest: match[3]
        };
      }
      function resolve(relative, base) {
        if (relative === "")
          return base;
        var path = (base || "/").split("/").slice(0, -1).concat(relative.split("/")), i = path.length, last = path[i - 1], unshift = false, up = 0;
        while (i--) {
          if (path[i] === ".") {
            path.splice(i, 1);
          } else if (path[i] === "..") {
            path.splice(i, 1);
            up++;
          } else if (up) {
            if (i === 0)
              unshift = true;
            path.splice(i, 1);
            up--;
          }
        }
        if (unshift)
          path.unshift("");
        if (last === "." || last === "..")
          path.push("");
        return path.join("/");
      }
      function Url(address, location, parser) {
        address = trimLeft(address);
        if (!(this instanceof Url)) {
          return new Url(address, location, parser);
        }
        var relative, extracted, parse, instruction, index, key, instructions = rules.slice(), type = typeof location, url = this, i = 0;
        if ("object" !== type && "string" !== type) {
          parser = location;
          location = null;
        }
        if (parser && "function" !== typeof parser)
          parser = qs.parse;
        location = lolcation(location);
        extracted = extractProtocol(address || "");
        relative = !extracted.protocol && !extracted.slashes;
        url.slashes = extracted.slashes || relative && location.slashes;
        url.protocol = extracted.protocol || location.protocol || "";
        address = extracted.rest;
        if (!extracted.slashes)
          instructions[3] = [/(.*)/, "pathname"];
        for (; i < instructions.length; i++) {
          instruction = instructions[i];
          if (typeof instruction === "function") {
            address = instruction(address);
            continue;
          }
          parse = instruction[0];
          key = instruction[1];
          if (parse !== parse) {
            url[key] = address;
          } else if ("string" === typeof parse) {
            if (~(index = address.indexOf(parse))) {
              if ("number" === typeof instruction[2]) {
                url[key] = address.slice(0, index);
                address = address.slice(index + instruction[2]);
              } else {
                url[key] = address.slice(index);
                address = address.slice(0, index);
              }
            }
          } else if (index = parse.exec(address)) {
            url[key] = index[1];
            address = address.slice(0, index.index);
          }
          url[key] = url[key] || (relative && instruction[3] ? location[key] || "" : "");
          if (instruction[4])
            url[key] = url[key].toLowerCase();
        }
        if (parser)
          url.query = parser(url.query);
        if (relative && location.slashes && url.pathname.charAt(0) !== "/" && (url.pathname !== "" || location.pathname !== "")) {
          url.pathname = resolve(url.pathname, location.pathname);
        }
        if (!required(url.port, url.protocol)) {
          url.host = url.hostname;
          url.port = "";
        }
        url.username = url.password = "";
        if (url.auth) {
          instruction = url.auth.split(":");
          url.username = instruction[0] || "";
          url.password = instruction[1] || "";
        }
        url.origin = url.protocol && url.host && url.protocol !== "file:" ? url.protocol + "//" + url.host : "null";
        url.href = url.toString();
      }
      function set(part, value, fn) {
        var url = this;
        switch (part) {
          case "query":
            if ("string" === typeof value && value.length) {
              value = (fn || qs.parse)(value);
            }
            url[part] = value;
            break;
          case "port":
            url[part] = value;
            if (!required(value, url.protocol)) {
              url.host = url.hostname;
              url[part] = "";
            } else if (value) {
              url.host = url.hostname + ":" + value;
            }
            break;
          case "hostname":
            url[part] = value;
            if (url.port)
              value += ":" + url.port;
            url.host = value;
            break;
          case "host":
            url[part] = value;
            if (/:\d+$/.test(value)) {
              value = value.split(":");
              url.port = value.pop();
              url.hostname = value.join(":");
            } else {
              url.hostname = value;
              url.port = "";
            }
            break;
          case "protocol":
            url.protocol = value.toLowerCase();
            url.slashes = !fn;
            break;
          case "pathname":
          case "hash":
            if (value) {
              var char = part === "pathname" ? "/" : "#";
              url[part] = value.charAt(0) !== char ? char + value : value;
            } else {
              url[part] = value;
            }
            break;
          default:
            url[part] = value;
        }
        for (var i = 0; i < rules.length; i++) {
          var ins = rules[i];
          if (ins[4])
            url[ins[1]] = url[ins[1]].toLowerCase();
        }
        url.origin = url.protocol && url.host && url.protocol !== "file:" ? url.protocol + "//" + url.host : "null";
        url.href = url.toString();
        return url;
      }
      function toString(stringify) {
        if (!stringify || "function" !== typeof stringify)
          stringify = qs.stringify;
        var query, url = this, protocol = url.protocol;
        if (protocol && protocol.charAt(protocol.length - 1) !== ":")
          protocol += ":";
        var result = protocol + (url.slashes ? "//" : "");
        if (url.username) {
          result += url.username;
          if (url.password)
            result += ":" + url.password;
          result += "@";
        }
        result += url.host + url.pathname;
        query = "object" === typeof url.query ? stringify(url.query) : url.query;
        if (query)
          result += "?" !== query.charAt(0) ? "?" + query : query;
        if (url.hash)
          result += url.hash;
        return result;
      }
      Url.prototype = {
        set,
        toString
      };
      Url.extractProtocol = extractProtocol;
      Url.location = lolcation;
      Url.trimLeft = trimLeft;
      Url.qs = qs;
      module.exports = Url;
    }
  };
  return __require(8);
})();
