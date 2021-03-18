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

    console.log("mousedowned something", anchor, isIntercepting);
  }
);

window.addEventListener(
  "click", 
  function (event) {
    const anchor = capture(event);
    if (!anchor) return;

    console.log("clicked something", anchor, isIntercepting);

    try {
      const url = new URL(anchor.href, document.baseURI).href;
      console.log("opening url in finicky", url);
      window.location = "finicky://" + url;
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