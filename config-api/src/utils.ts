export function createRegularExpression(pattern: string) {
    if (!pattern) {
      return undefined;
    }
  
    let result = pattern;
  
    result = result.replace(/[-[\]\/{}()*+?.,\\^$|#\s]/g, "\\$&");    
    result = result.replace(/\\\*/g, ".*");
  
    if (
      !pattern.startsWith("http://") &&
      !pattern.startsWith("https://")
    ) {
      result = "https?:\\/\\/" + result;
    }
  
    return new RegExp("^" + result + "$", 'i');
  }