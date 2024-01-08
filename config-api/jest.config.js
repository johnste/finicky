module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",

  testPathIgnorePatterns: ["/node_modules/", "src/fastidious/"],
  transform: {
    "^.+\\.ts$": ["ts-jest"],
  },
};
