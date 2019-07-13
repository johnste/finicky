import typescript from "rollup-plugin-typescript2";
import packagejson from "./package.json";

export default {
  input: "./index.ts",
  plugins: [typescript(/*{ plugin options }*/)],
  output: {
    intro: `/* finicky config api ${packagejson.version} */`,
    file: "../Finicky/Finicky/finickyConfigAPI.js",
    format: "iife",
    name: "finickyConfigApi",
    noInterop: true
  }
  //external: ["react", "styled-components"]
};
