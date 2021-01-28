import { validate as v, getErrors as f } from "./index";

describe("test", () => {
  test("empty schema", () => {
    expect(f({}, {})).toHaveLength(0);
  });

  test("Optional", () => {
    const schema = {
      a: v.boolean,
    };

    expect(f({ a: true }, schema)).toHaveLength(0);
    expect(f({ a: undefined }, schema)).toHaveLength(0);
    expect(f({}, schema)).toHaveLength(0);
  });

  test("Required", () => {
    const schema = {
      a: v.boolean.isRequired,
    };

    expect(f({ a: true }, schema)).toHaveLength(0);
    expect(f({ a: undefined }, schema)).toHaveLength(1);
    expect(f({}, schema)).toHaveLength(1);
  });

  test("Extraneous properties", () => {
    const schema = {
      a: v.boolean,
    };

    expect(f({ b: true, c: "string" }, schema)).toHaveLength(2);
  });

  test("Boolean", () => {
    const schema = {
      a: v.boolean.isRequired,
    };

    expect(f({ a: true }, schema)).toHaveLength(0);
    expect(f({ a: false }, schema)).toHaveLength(0);
    expect(f({ a: "truthy value" }, schema)).toHaveLength(1);
    expect(f({ a: null }, schema)).toHaveLength(1);
    expect(f({ a: undefined }, schema)).toHaveLength(1);
  });

  test("Number", () => {
    const schema = {
      a: v.number.isRequired,
    };

    expect(f({ a: 5 }, schema)).toHaveLength(0);
    expect(f({ a: -413 }, schema)).toHaveLength(0);
    expect(f({ a: 0 }, schema)).toHaveLength(0);
    expect(f({ a: Math.PI }, schema)).toHaveLength(0);
    expect(f({ a: Number.MAX_VALUE }, schema)).toHaveLength(0);
    expect(f({ a: Number.POSITIVE_INFINITY }, schema)).toHaveLength(0);
    expect(f({ a: Number.MIN_VALUE }, schema)).toHaveLength(0);
    expect(f({ a: Number.NaN }, schema)).toHaveLength(1);
    expect(f({ a: "truthy value" }, schema)).toHaveLength(1);
    expect(f({ a: null }, schema)).toHaveLength(1);
    expect(f({ a: undefined }, schema)).toHaveLength(1);
  });

  test("String", () => {
    const schema = {
      a: v.string.isRequired,
    };

    expect(f({ a: "test" }, schema)).toHaveLength(0);
    expect(f({ a: "0" }, schema)).toHaveLength(0);
    expect(f({ a: "" }, schema)).toHaveLength(0);
    expect(f({ a: {} }, schema)).toHaveLength(1);
    expect(f({ a: [] }, schema)).toHaveLength(1);
  });

  test("Function", () => {
    const schema = {
      a: v.function("test").isRequired,
    };

    expect(f({ a() {} }, schema)).toHaveLength(0);
    expect(f({ a: (x: any) => x }, schema)).toHaveLength(0);
    expect(f({ a: {} }, schema)).toHaveLength(1);
    expect(f({ a: false }, schema)).toHaveLength(1);
  });

  test("RegExp", () => {
    const schema = {
      a: v.regex.isRequired,
    };

    expect(f({ a: /test/ }, schema)).toHaveLength(0);
    expect(f({ a: new RegExp("test") }, schema)).toHaveLength(0);
    expect(f({ a: {} }, schema)).toHaveLength(1);
    expect(f({ a: false }, schema)).toHaveLength(1);
  });

  describe("Value", () => {
    test("validator", () => {
      const schema = {
        a: v.value("value").isRequired,
      };

      expect(f({ a: "value" }, schema)).toHaveLength(0);
      expect(f({ a: "not value" }, schema)).toHaveLength(1);
      expect(f({ a: undefined }, schema)).toHaveLength(1);
    });

    test("validator", () => {
      const value: any[] = [];
      const schema = {
        a: v.value(value).isRequired,
      };

      expect(f({ a: value }, schema)).toHaveLength(0);
      expect(f({ a: [] }, schema)).toHaveLength(1);
      expect(f({ a: value.slice() }, schema)).toHaveLength(1);
    });

    test("null value", () => {
      const schema = {
        a: v.value(null).isRequired,
      };

      expect(f({ a: null }, schema)).toHaveLength(0);
    });
  });

  test("OneOf", () => {
    const schema = {
      a: v.oneOf([v.boolean, v.string]).isRequired,
    };

    expect(f({ a: "test" }, schema)).toHaveLength(0);
    expect(f({ a: true }, schema)).toHaveLength(0);
    expect(f({ a: undefined }, schema)).toHaveLength(1);
    expect(f({ a: /regex/ }, schema)).toHaveLength(1);
    expect(f({ a() {} }, schema)).toHaveLength(1);
  });

  test("OneOf null", () => {
    const schema = {
      value: v.oneOf([
        v.string,
        v.shape({
          name: v.string.isRequired,
          appType: v.oneOf(["appName", "bundleId", "appPath"]),
          openInBackground: v.boolean,
          args: v.arrayOf(v.string),
        }),
        v.function("options"),
        v.value(null),
      ]).isRequired,
    };
    expect(f({ value: null }, schema)).toHaveLength(0);
  });

  describe("ArrayOf", () => {
    test("optional", () => {
      const schema = {
        a: v.arrayOf(v.boolean).isRequired,
      };

      expect(f({ a: [true, false] }, schema)).toHaveLength(0);
      expect(f({ a: [true, undefined] }, schema)).toHaveLength(0);
      expect(f({ a: [true, "test"] }, schema)).toHaveLength(1);
    });

    test("required", () => {
      const schema = {
        a: v.arrayOf(v.boolean.isRequired).isRequired,
      };

      expect(f({ a: [true, false] }, schema)).toHaveLength(0);
      expect(f({ a: [true, undefined] }, schema)).toHaveLength(1);
    });
  });

  test("shape", () => {
    const schema = {
      a: v.shape({ bool: v.boolean.isRequired, string: v.string }).isRequired,
    };

    expect(f({ a: { bool: true, string: "string" } }, schema)).toHaveLength(0);
    expect(f({ a: { bool: true } }, schema)).toHaveLength(0);
    expect(f({ a: { string: "string" } }, schema)).toHaveLength(1);
  });

  test("shape 2", () => {
    const schema = {
      a: v.shape({ bool: v.boolean.isRequired, string: v.string.isRequired })
        .isRequired,
    };

    expect(f({ a: {} }, schema)).toHaveLength(1);
    expect(f({ a: true }, schema)).toHaveLength(1);
  });
});

describe("Complex", () => {
  test("valid", () => {
    const schema = {
      app: v.oneOf([
        v.string,
        v.shape({
          name: v.string.isRequired,
          appType: v.oneOf(["appName", "bundleId", "appPath"]),
          openInBackground: v.boolean,
          args: v.arrayOf(v.string),
        }),
      ]).isRequired,
    };

    expect(f({ app: "string" }, schema)).toHaveLength(0);
    expect(
      f(
        {
          app: {
            name: "string",
            appType: "appName",
          },
        },
        schema
      )
    ).toHaveLength(0);
  });

  test("invalid output", () => {
    const schema = {
      lapp: v.shape({ korv: v.boolean.isRequired }).isRequired,
      app: v.oneOf([
        v.string,
        v.shape({
          name: v.string.isRequired,
          appType: v.oneOf(["appName", "bundleId", "appPath"]),
          openInBackground: v.boolean,
          args: v.arrayOf(v.string),
          anotherShape: v.shape({
            boool: v.boolean,
          }),
        }),
      ]).isRequired,
    };

    expect(
      f(
        {
          app: /error/,
        },
        schema
      )
    ).toMatchSnapshot();
  });

  test("Serialize functions to readable description", () => {
    const schema = {
      lapp: v.boolean,
    };

    expect(
      f(
        {
          lapp: () => {},
        },
        schema
      )
    ).toMatchInlineSnapshot(`
                  Array [
                    "Value at root.lapp: [Function lapp] is not boolean",
                  ]
            `);
  });

  test("Serialize expected objects to readable description", () => {
    const schema = {
      test: v.shape({ key: v.boolean }).isRequired,
    };

    expect(
      f(
        {
          test: "true",
        },
        schema
      )
    ).toMatchInlineSnapshot(`
      Array [
        "Value at root.test: [\\"true\\"] is not {
        \\"key\\": \\"boolean\\"
      }",
      ]
    `);
  });
});

describe("Single values", () => {
  test("single valid values", () => {
    expect(f(true, v.boolean)).toHaveLength(0);
    expect(f(false, v.boolean)).toHaveLength(0);
    expect(f(/regex/, v.regex)).toHaveLength(0);
    expect(f(4, v.number)).toHaveLength(0);
    expect(f(() => {}, v.function("arg"))).toHaveLength(0);
    expect(f(4, v.value(4))).toHaveLength(0);
    expect(f([4], v.arrayOf(v.number))).toHaveLength(0);
  });

  test("single invalid values", () => {
    expect(f("test", v.boolean)).toHaveLength(1);
    expect(f({}, v.boolean)).toHaveLength(1);
    expect(f(/regex/, v.boolean)).toHaveLength(1);
    expect(f(/regex/, v.number)).toHaveLength(1);
  });

  test("invalid schemas", () => {
    // @ts-ignore
    expect(f(undefined, "test")).toHaveLength(1);
    // @ts-ignore
    expect(f(undefined, undefined)).toHaveLength(1);
    // @ts-ignore
    expect(f({ test: true }, false)).toHaveLength(1);
  });
});

test("test x", () => {
  const browserSchema = v.oneOf([
    v.string,
    v.shape({
      name: v.string.isRequired,
      appType: v.oneOf(["appName", "bundleId", "appPath"]),
      openInBackground: v.boolean,
      args: v.arrayOf(v.string),
    }),
    v.function("options"),
    v.value(null),
  ]);

  const multipleBrowsersSchema = v.oneOf([
    browserSchema,
    v.arrayOf(browserSchema.isRequired),
  ]);

  const schema = {
    defaultBrowser: multipleBrowsersSchema.isRequired,
  };

  expect(
    f(
      {
        defaultBrowser: ["fest", "test"],
      },
      schema
    )
  ).toHaveLength(0);
});
