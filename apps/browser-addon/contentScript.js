let isIntercepting = false;

window.addEventListener("keydown", (event) => {
  if (event.key.toLowerCase() === "alt") {
    isIntercepting = true;
  }
});

window.addEventListener("keyup", (event) => {
  if (event.key.toLowerCase() === "alt") {
    isIntercepting = false;
  }
});

window.addEventListener("blur", (event) => {
  isIntercepting = false;
});

function getAnchor(element) {
  do {
    if (element.tagName?.toLowerCase() === "a") {
      return element;
    }

    element = element.parentNode;
  } while (element.parentNode);

  return undefined;
}

window.addEventListener(
  "mousedown", 
  function (event) {
    const anchor = getAnchor(event.target);
    if (!anchor) return;
  }
);

window.addEventListener(
  "click", 
  function (event) {
    const anchor = capture(event);
    if (!anchor) return;

    try {
      const url = new URL(anchor.href, document.baseURI).href;          
      window.location = "finicky://open/" + btoa(url);
    } catch (ex) {
      console.error("Finicky Browser Extension Error", ex);
    }
  },
  true
);

function capture(event) {
  if (!isIntercepting) return;    

  const anchor = getAnchor(event.target);

  if (!anchor?.hasAttribute("href")) return;

  event.preventDefault();
  event.stopImmediatePropagation();

  return anchor;
}