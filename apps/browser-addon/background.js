const menuItemId = "finicky-open-url";

// Service workers don't have access to window object, use chrome directly
const browser = chrome;

// Create context menu when service worker starts
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: menuItemId,
    title: "Open with Finicky",
    contexts: ["link"],
  });
});

// Handle context menu clicks
chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId !== menuItemId) {
    return;
  }

  console.log("Finicky Browser Extension: Opening link in Finicky", info.linkUrl);

  chrome.tabs.update(tab.id, { url: "finicky://open/" + btoa(info.linkUrl) });
});
