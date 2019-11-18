import urlParse from "url-parse";
import path from "path";
import chalk from "chalk";

import { processUrl } from "./processUrl";
import { validateConfig } from "./validateConfig";
import { createAPI } from "./createAPI";

const [, , ...args] = process.argv;

const finicky = createAPI({
  log(message) {
    console.log(chalk`{dim [log]} ${message}`);
  },
  notify(title, subtitle) {
    console.log(chalk`{dim [notification]} {bold ${title}} ${subtitle}`);
  }
});

function errorMessage(message: string, exception: Error | string) {
  console.log(
    chalk`{red.bold Error:} ${message} {red ${exception.toString()}}`
  );
}

// @ts-ignore
global.finicky = finicky;

const configPath = path.resolve(
  process.cwd(),
  args[1] || "./.finicky.example.js"
);

const url = args[0] || "https://example.com/test";

const protocol = urlParse(url).protocol.replace(":", "");

console.log(chalk`Opening configuration file {dim ${configPath}}`);

let config;

try {
  config = require(configPath);
} catch (ex) {
  errorMessage("Couldn't open configuration file. ", ex);
  process.exit(1);
}

try {
  validateConfig(config);
} catch (ex) {
  errorMessage("Couldn't validate configuration. ", ex);
  process.exit(1);
}

try {
  const result = processUrl(config, url);
  console.log(chalk`Result:
{green ${JSON.stringify(result, null, 2)}}`);
} catch (ex) {
  errorMessage("Couldn't validate result. ", ex);
  process.exit(1);
}
