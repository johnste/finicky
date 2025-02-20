type Matcher = string | RegExp

export function matchHostnames(matchers: Matcher | Matcher[]) {

    matchers = Array.isArray(matchers) ? matchers : [matchers];
  
    matchers.forEach((matcher) => {
      if (matcher instanceof RegExp || typeof matcher === "string") {
        return;
      }
      throw new Error(
        `finicky.matchHostnames: Unrecognized hostname type "${typeof matcher}"`
      );
    });
  
    return function(url: URL) {      
      const hostname = url.hostname;
      
      if (!hostname) {        
        console.warn('finicky.matchHostnames: No hostname', url.href);
        return false;
      }

      return (matchers as Matcher[]).some((matcher) => {        
        if (matcher instanceof RegExp) {
          return matcher.test(hostname);
        } else if (typeof matcher === "string") {
          return matcher === hostname;
        }
  
        return false;
      });
    };
  }

