import { ProcessInfo } from "./configSchema";
import { LegacyURLObject } from "./legacyURLObject";
import { URLtoLegacyURLObject } from "./legacyURLObject";

/**
 * FinickyURL class that extends URL to maintain backward compatibility
 * with legacy properties while providing deprecation warnings.
 */
export class FinickyURL extends URL {
  private _opener: ProcessInfo | null;

  constructor(url: string, opener: ProcessInfo | null = null) {
    super(url);
    this._opener = opener;
  }

  get urlString(): string {
    console.warn(
      'Accessing legacy property "urlString" that is no longer supported. This first argument to the function is a URL instance, you should be able to use its href property directly instead. See https://developer.mozilla.org/en-US/docs/Web/API/URL for reference.'
    );
    return this.href;
  }

  get url(): LegacyURLObject {
    console.warn(
      'Accessing legacy property "url" that is no longer supported. This first argument to the function is a URL instance, a standard interface for URLs. https://developer.mozilla.org/en-US/docs/Web/API/URL'
    );
    return URLtoLegacyURLObject(this);
  }

  get opener(): ProcessInfo | null {
    console.warn(
      'Accessing legacy property "opener" that is no longer supported.'
    );
    return this._opener;
  }

  get keys() {
    throw new Error(
      'Accessing legacy property "keys" that is no longer supported, please use finicky.getModifierKeys() instead.'
    );
  }
}
