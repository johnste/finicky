import { createRegularExpression } from "../src/utils";

test("Generates regular expression strings", () => {
  let re = createRegularExpression("http://example.com");
  expect(re).toMatchInlineSnapshot(`/\\^http:\\\\/\\\\/example\\\\\\.com\\$/i`);

  re = createRegularExpression("example.com");
  expect(re).toMatchInlineSnapshot(
    `/\\^https\\?:\\\\/\\\\/example\\\\\\.com\\$/i`
  );

  re = createRegularExpression("example.com/*");
  expect(re).toMatchInlineSnapshot(
    `/\\^https\\?:\\\\/\\\\/example\\\\\\.com\\\\/\\.\\*\\$/i`
  );
});

test("Regular expressions", () => {
  let re = createRegularExpression("http://example.com");

  expect(re).toBeInstanceOf(RegExp);

  expect(re.test("http://example.com")).toBe(true);
  expect(re.test("https://example.com")).toBe(false);

  re = createRegularExpression("example.com");
  expect(re.test("http://example.com")).toBe(true);
  expect(re.test("https://example.com")).toBe(true);

  re = createRegularExpression("https://example.com");
  expect(re.test("http://example.com")).toBe(false);
  expect(re.test("https://example.com")).toBe(true);
});

test("Wildcards", () => {
  let re = createRegularExpression("http://example.com/*");

  expect(re.test("http://example.com/abcdefghijklmnopq")).toBe(true);
  expect(re.test("https://example.comabcdefg")).toBe(false);

  re = createRegularExpression("something.*.com*");
  expect(re.test("http://something.example.comst")).toBe(true);
  expect(re.test("https://something.example.com/test?abc=123#124")).toBe(true);
  expect(re.test("https://something.example.net/.com/test?abc=123#124")).toBe(
    true
  );
  expect(re.test("https://somethingexample.net/.com/test?abc=123#124")).toBe(
    false
  );
});
