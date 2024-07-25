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
'use strict';

const initialize = () => api.runtime.onInstalled.addListener((details) => {
  const reason = details.reason
  switch (reason) {
    case 'install':
    case 'update':
      console.log('Thanks for using Finicky Helper')
      api.storage.local.set({
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
  api.tabs.update(details.tabId, { url: newUrl });
};

const listenForNav = () => chrome.webNavigation.onBeforeNavigate.addListener(details => {
  api.storage.local.get(['patterns', 'prefixes'], ({ patterns = [], prefixes = [] })  => {
  const url = new URL(details.url);
    if (checkUrlAgainstPatterns(url.href, patterns, prefixes)) {
      handleNavigation(details, url);
    }
  });
}, { url: [{ urlMatches: 'http://*/*' }, { urlMatches: 'https://*/*' }] }
);

const api = typeof chrome !== undefined ? chrome : browser
initialize();
listenForNav();