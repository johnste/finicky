let isIntercepting = false;
console.log("👇 hiya")

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
  "mousedown", // doubleclick instead?
  function (event) {
    if (!isIntercepting) {
      return; 
    }
   
    event.preventDefault();
    event.stopImmediatePropagation();
    
    const anchor = getAnchor(event.target);

    
    if (!anchor?.hasAttribute("href")) {
      return;
    }
    console.log('mousedowned something', anchor, isIntercepting)
    event.preventDefault();
    event.stopImmediatePropagation();
    
  })


window.addEventListener(
  "click", // doubleclick instead?
  function (event) {
    if (!isIntercepting) {
      return; 
    }
    event.preventDefault();
    event.stopImmediatePropagation();
    
    const anchor = getAnchor(event.target);

    
    if (!anchor?.hasAttribute("href")) {
      return;
    }
    console.log('clicked something', anchor, isIntercepting)

    try {
      const url = new URL(anchor.href, document.baseURI).href;
      console.log("opening url in finicky", url);      
      window.location = "finicky://" + url
    } catch (ex) {
      console.log(ex);
    }

    event.preventDefault();
    event.stopImmediatePropagation();
  },
  true
);
