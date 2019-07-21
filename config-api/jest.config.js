module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",

  testPathIgnorePatterns: ["/node_modules/", "src/fastidious/"],
  globals: {
    "ts-jest": {
      diagnostics: true
    }
  }
};
