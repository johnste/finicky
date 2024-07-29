/*!
* File: FinickyHelper.js
* Description: Popup JS. Redirects http(s):// to finicky(s):// for specific URLs
* Author: Alex Torma (https://github.com/torma616/)
* Date: July 24, 2024
* Copyright: ©2024 Alex Torma
* License: MIT
* Version: 1
*
*/
'use strict';
(() => {
  const createElement = ({ tag, ...props } = {}) => {
    const element = document.createElement(tag);
    Object.entries(props).forEach(([key, value]) => {
      if (key === 'children' && Array.isArray(value)) {
        value.forEach(child => element.appendChild(child));
      } else if (key === 'classList' && Array.isArray(value)) {
        value.forEach(cls => element.classList.add(cls))
      } else if (key && value !== undefined) {
        element[key] = value;
      }
    });
    return element;
  };

  const uniqueStringGen = (prefix = 'temp') => {
    let count = 0;
    return () => `${prefix}${++count}`;
  };
  const uniqueString = uniqueStringGen('temp');

  const setUpEnv = (currentBrowser) => {
    const currentBrowserInput = document.querySelector('#current-browser input');
    const writeToStorage = () => FinickyHelper.storage.local.set({ currentBrowser: currentBrowserInput.value })

    if (currentBrowser === '' && currentBrowserInput.value !== '') {
      writeToStorage();
    }

    if (currentBrowser !== '' && currentBrowserInput.value === '') {
      currentBrowserInput.value = currentBrowser;
      writeToStorage();
    }

    if (currentBrowser === '' && currentBrowserInput.value === '') {
      const userInput = prompt('Please put the name of the current browser in the "Current Browser" field.');
      currentBrowserInput.value = currentBrowser = userInput;
      writeToStorage();
    };

    currentBrowserInput.addEventListener('input', writeToStorage)
  }

  const setUpFields = (currentBrowser, patterns, prefixes) => {
    console.log(patterns)
    setUpEnv(currentBrowser);
    document.querySelector('#prefixes').innerHTML = document.getElementById('match-patterns').innerHTML = '';
    patterns.forEach(([browser, patternEntries, type = 'pattern']) => {
      createBrowserInput(type, browser);
      patternEntries.forEach(pattern => createInput(type, browser, pattern));
      ensureEmptyField(type, browser);
    });
    ensureEmptyField('pattern', 'browser', true);

    prefixes.forEach(prefix => createInput('prefix', '', prefix));
    ensureEmptyField('prefix', '');
  };

  const createBrowserInput = (type, browser) => {
    const newInput = createElement({ tag: 'input', placeholder: "Enter Browser", type: "text", classList: ['browser-input'] });
    newInput.addEventListener('input', handleInput);

    const matchWrapper = createElement({
      tag: 'div', classList: [`${browser}-patterns`], children: [
        createElement({
          tag: 'div', classList: ['browsers'], children: [
            createElement({ tag: 'div', className: type, children: [newInput] })
          ]
        }),
        createElement({ tag: 'div', classList: ['patterns'] })
      ]
    });

    document.getElementById('match-patterns').appendChild(matchWrapper);
    newInput.value = !browser.startsWith('temp') ? browser : '';
  };

  const createInput = (type, browser = '', value = '') => {
    const placeholder = `Enter URL ${type.charAt(0).toUpperCase() + type.slice(1)}`;
    const targetDiv = type === 'pattern'
      ? document.querySelector(`.${browser}-patterns .patterns`)
      : document.querySelector('#prefixes');
    const newInput = createElement({ tag: 'input', type: 'text', placeholder, value, classList: [`${type}-input`] });
    newInput.addEventListener('input', handleInput);
    const inputWrapper = createElement({ tag: 'div', className: type, children: [newInput] });
    targetDiv.appendChild(inputWrapper);
  };

  const handleInput = (event) => {
    const inputField = event.target;
    const type = inputField.parentNode.classList.contains('pattern') ? 'pattern' : 'prefix';
    const browser = inputField.closest('[class*="-patterns"]')?.classList[0]?.split('-')?.[0] ?? '';

    inputField.classList.contains('browser-input')
      ? handleBrowserInput(inputField, type, browser)
      : handleTextInput(inputField, type, browser);
    saveMatches();
  };

  const handleBrowserInput = (inputField, type, browser) => {
    const browserInputs = document.querySelectorAll(`.${browser}-patterns input.browser-input`);

    if (browserInputs[browserInputs.length - 1] === inputField) {
      if (inputField.value === '') {
        inputField.addEventListener('blur', () => inputField.closest('[class*="-patterns"]').remove());
      }
      ensureEmptyField(type, 'browser', true);
    } else if (browserInputs.length > 1) {
      inputField.addEventListener('blur', () => inputField.parentElement.remove());
      ensureEmptyField(type, browser, true);
    }
  };

  const handleTextInput = (inputField, type, browser) => {
    const inputFields = type === 'pattern'
      ? document.querySelectorAll(`.${browser}-patterns input`)
      : document.querySelectorAll('#prefixes input');

    if (inputField.value.trim()) {
      if (inputFields[inputFields.length - 1] === inputField) {
        createInput(type, browser);
        ensureEmptyField(type, browser);
      }
    } else if (inputFields.length > 1) {
      inputField.addEventListener('blur', () => inputField.parentElement.remove());
      ensureEmptyField(type, browser);
    }
  };

  const ensureEmptyField = (type, browser, browserDiv = false) => {
    if (browserDiv) {
      browser = browser !== 'browser' ? browser : uniqueString();
      const browserInputFields = document.querySelectorAll('.browsers input');
      if (browserInputFields.length === 0 || browserInputFields[browserInputFields.length - 1].value.trim() !== '') {
        createBrowserInput(type, browser);
        createInput(type, browser);
      }
      return;
    }

    const targetDiv = type === 'pattern'
      ? document.querySelector(`.${browser}-patterns .patterns`)
      : document.querySelector('#prefixes');
    const inputFields = targetDiv.querySelectorAll('input');

    if (inputFields.length === 0 || inputFields[inputFields.length - 1].value.trim()) {
      createInput(type, browser);
    }
  };

  const saveMatches = () => {
    const browsers = [...document.querySelectorAll('.browsers input.browser-input')];
    const prefixInputs = document.querySelectorAll('#prefixes input');

    const patterns = browsers.map(browser => [
      browser.value,
      [...browser.closest('.browsers').parentNode.querySelectorAll('input.pattern-input')].map(field => field.value.trim()).filter(Boolean)
    ]).filter(([_, array]) => array.length > 0);
    const prefixes = Array.from(prefixInputs).map(field => field.value.trim()).filter(Boolean);

    FinickyHelper.storage.local.set({ patterns, prefixes });
  };

  const initialize = () =>
    FinickyHelper.storage.local.get(['currentBrowser', 'patterns', 'prefixes'], ({ currentBrowser = '', patterns = [], prefixes = [] }) =>
      setUpFields(currentBrowser, patterns, prefixes));

  document.addEventListener('DOMContentLoaded', initialize)
})();
