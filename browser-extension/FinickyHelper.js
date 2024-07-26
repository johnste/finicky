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

  const setUpFields = (patterns, prefixes) => {
    document.querySelector('#prefixes').innerHTML = document.getElementById('match-patterns').innerHTML = '';
    patterns.forEach(([browser, patternEntries, type = 'pattern']) => {
      createSelect(type, browser);
      patternEntries.forEach(pattern => createInput(type, browser, pattern));
      ensureEmptyField(type, browser);
    });
    ensureEmptyField('pattern', 'make-selection', true);

    prefixes.forEach(prefix => createInput('prefix', '', prefix));
    ensureEmptyField('prefix', '');
  };

  const createOptions = (options) => options.map(opt => createElement({ tag: 'option', ...opt }));

  const createSelect = (type, browser) => {
    const newSelect = createElement({
      tag: 'select', children: createOptions([
        { value: 'make-selection', innerHTML: 'Choose A Browser' },
        { value: 'chrome', innerHTML: 'Google Chrome' },
        { value: 'edge', innerHTML: 'Microsoft Edge' },
        { value: 'chromium', innerHTML: 'Chromium' }
      ])
    });
    newSelect.addEventListener('input', handleInput);

    const matchWrapper = createElement({
      tag: 'div', classList: [`${browser}-patterns`], children: [
        createElement({
          tag: 'div', classList: ['browsers'], children: [
            createElement({ tag: 'div', className: type, children: [newSelect] })
          ]
        }),
        createElement({ tag: 'div', classList: ['patterns'] })
      ]
    });

    document.getElementById('match-patterns').appendChild(matchWrapper);
    newSelect.value = ['chrome', 'chromium', 'edge'].includes(browser) ? browser : 'make-selection';
  };

  const createInput = (type, browser = '', value = '') => {
    const placeholder = `Enter URL ${type.charAt(0).toUpperCase() + type.slice(1)}`;
    const targetDiv = type === 'pattern'
      ? document.querySelector(`.${browser}-patterns .patterns`)
      : document.querySelector('#prefixes');
    const newInput = createElement({ tag: 'input', type: 'text', placeholder, value });
    newInput.addEventListener('input', handleInput);
    const inputWrapper = createElement({ tag: 'div', className: type, children: [newInput] });
    targetDiv.appendChild(inputWrapper);
  };

  const handleInput = (event) => {
    const inputField = event.target;
    const type = inputField.parentNode.classList.contains('pattern') ? 'pattern' : 'prefix';
    const browser = inputField.closest('[class*="-patterns"]')?.classList[0]?.split('-')?.[0] ?? '';

    (inputField.tagName === 'SELECT' && inputField.tagName !== 'INPUT')
      ? handleSelectInput(inputField, type, browser)
      : handleTextInput(inputField, type, browser);
    saveMatches();
  };

  const handleSelectInput = (inputField, type, browser) => {
    const browserSelects = document.querySelectorAll(`.${browser}-patterns select`);
    const lastSelect = browserSelects[browserSelects.length - 1];

    if (lastSelect === inputField) {
      if (inputField.value === 'make-selection') {
        inputField.closest('[class*="-patterns"]').remove();
      }
      ensureEmptyField(type, 'make-selection', true);
    } else if (browserSelects.length > 1) {
      inputField.parentElement.remove();
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
      inputField.parentElement.remove();
      ensureEmptyField(type, browser);
    }
  };

  const ensureEmptyField = (type, browser, browserDiv = false) => {
    if (browserDiv) {
      browser = browser !== 'make-selection' ? browser : uniqueString();
      const selectFields = document.querySelectorAll('.browsers select');
      if (selectFields.length === 0 || selectFields[selectFields.length - 1].value.trim() !== 'make-selection') {
        createSelect(type, browser);
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
    const browsers = [...document.querySelectorAll('.browsers select')];
    const prefixInputs = document.querySelectorAll('#prefixes input');

    const patterns = browsers.map(browser => [
      browser.value,
      [...browser.closest('.browsers').parentNode.querySelectorAll('input')].map(field => field.value.trim()).filter(Boolean)
    ]).filter(([_, array]) => array.length > 0);
    const prefixes = Array.from(prefixInputs).map(field => field.value.trim()).filter(Boolean);

    FinickyHelper.storage.local.set({ patterns, prefixes });
  };

  const initialize = () =>
    FinickyHelper.storage.local.get(['patterns', 'prefixes'], ({ patterns = [], prefixes = [] }) =>
      setUpFields(patterns, prefixes));

  document.addEventListener('DOMContentLoaded', initialize)
})();
