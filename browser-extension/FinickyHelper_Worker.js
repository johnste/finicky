/*
* File: FinickyHelper_Worker.js
* Description: Finicky Helper Service Worker. Redirects http(s):// to finicky(s):// for specific URLs
* Author: Alex Torma (https://github.com/torma616/)
* Date: July 24, 2024
* Copyright: © 2024 Alex Torma
* License: MIT
* Version: 1
*
*/
'use strict';
(() => {
  const setupExtension = () => FinickyHelper.runtime.onInstalled.addListener((details) => {
    const reason = details.reason
    switch (reason) {
      case 'install':
      case 'update':
        console.log('Thanks for using Finicky Helper')
        FinickyHelper.storage.local.set({
          patterns: [
          ],
          prefixes: [
          ]
        });
        break;

      default:
        console.log(reason, 'Other install events within the browser')
        break;
    }
  })

  const checkUrlAgainstPatterns = (url, patterns, prefixes) => {
    const protocolPattern = /^https?:\/\//i;
    const strippedUrl = url.replace(protocolPattern, '');
    prefixes.unshift('');

    const patternRegexes = patterns.flatMap(pattern => {
      return prefixes.map(prefix => {
        const prefixedPattern = `${prefix}${pattern.replace(protocolPattern, '')}`;
        const regexPattern = `^${prefixedPattern.replace(/\./g, '\\.').replace(/\*/g, '.*')}$`;
        return new RegExp(regexPattern, 'i');
      });
    });
    return patternRegexes.some(regex => regex.test(strippedUrl));
  };

  const handleNavigation = (details, url) => {
    const scheme = url.protocol === 'https:' ? 'finickys' : 'finicky';
    const newUrl = `${scheme}://${url.hostname}${url.pathname}${url.search}${url.hash}`;
    FinickyHelper.tabs.update(details.tabId, { url: newUrl });
  };

  const listenForNav = () => FinickyHelper.webNavigation.onBeforeNavigate.addListener(details => {
    FinickyHelper.storage.local.get(['patterns', 'prefixes'], ({ patterns = [], prefixes = [] }) => {
      const url = new URL(details.url);
      const currentBrowser = detectBrowser();

      const filteredPatterns = patterns
        .filter(([browser]) => browser !== currentBrowser) // Exclude patterns for the current browser
        .flatMap(([, patternList]) => patternList);

      if (checkUrlAgainstPatterns(url.href, filteredPatterns, prefixes)) {
        handleNavigation(details, url);
      }
    });
  }, { url: [{ urlMatches: 'http://*/*' }, { urlMatches: 'https://*/*' }] }
  );

  const initialize = () => {
    importScripts('FinickyHelper_Shared.js');
    setupExtension();
    listenForNav();
  }

  initialize();
})();
