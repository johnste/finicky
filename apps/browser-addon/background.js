const menuItemId = "finicky-open-url";

let browser = typeof window !== "undefined" && typeof window.browser !== "undefined" ? window.browser : chrome;


browser.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId !== menuItemId) {
    return;
  }

  console.log("Finicky Browser Extension: Opening link in Finicky", info.linkUrl);

  browser.tabs.update(tab.id, { url: "finicky://open/" + btoa(info.linkUrl) });
});

try {
  chrome.contextMenus.create({
    id: menuItemId,
    title: "Open with Finicky",
    contexts: ["link"],
  });
} catch (ex) {}
