/*!
* File: FinickyHelper_Worker.js
* Description: Finicky Helper Service Worker. Redirects http(s):// to finicky(s):// for specific URLs
* Author: Alex Torma (https://github.com/torma616/)
* Date: July 24, 2024
* Copyright: © 2024 Alex Torma
* License: MIT
* Version: 1
*
*/
const FinickyHelper = (() => typeof chrome !== 'undefined' ? chrome : browser)();

const detectBrowser = () => {
  const userAgent = navigator.userAgent;

  if (typeof chrome !== 'undefined') {
    if (navigator.userAgentData) {
      const brands = navigator.userAgentData.brands;
      const brandMap = {
        'Google Chrome': 'chrome',
        'Microsoft Edge': 'edge',
        'Chromium': brands.length === 2 ? 'chromium' : null
      };

      for (const { brand } of brands) {
        if (brandMap[brand]) return brandMap[brand];
      }
    } else {
      if (userAgent.includes('Edg')) return 'edge';
      if (userAgent.includes('Chrome')) return 'chrome';
    }
  } else if (typeof browser !== 'undefined') {
    return 'firefox';
  }

  return 'Unknown';
};
