import { UrlObject } from "./../dist/src/types.d";
import { processUrl } from "../src/processUrl";
import { createAPI } from "../src/createAPI";
import { Url, UrlFunction } from "../src/types";

describe("Rewrite", () => {
  beforeAll(() => {
    // @ts-ignore
    global.finicky = createAPI();
  });

  const doRewrite = (urlResult: Url | UrlFunction) => {
    const exampleUrl = "https://test.example";

    return processUrl(
      {
        defaultBrowser: "test",
        rewrite: [{ match: () => true, url: urlResult }]
      },
      exampleUrl
    );
  };

  test("String", () => {
    const url = "https://test.example";
    const result = doRewrite(url);
    expect(result.url).toBe(url);
  });

  test("Function", () => {
    const url = "https://test.example";
    const result = doRewrite(() => url);
    expect(result.url).toBe(url);
  });

  test("Function arguments", () => {
    const result = doRewrite(({ urlString }) => urlString + "?ok");
    expect(result.url).toBe("https://test.example?ok");
  });

  test("Function arguments", () => {
    const result = doRewrite(
      ({ urlString, url }) => urlString + "?" + url.protocol
    );
    expect(result.url).toBe("https://test.example?https");
  });

  test("Object result ", () => {
    const result = doRewrite(({ url }) => ({
      ...url,
      host: "test2.example"
    }));

    expect(result.url).toBe("https://test2.example");
  });
});
