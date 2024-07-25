/*!
 * File: FinickyHelper.js
 * Description: Popup JS. Redirects http(s):// to finicky(s):// for specific URLs
 * Author: Alex Torma (https://github.com/torma616/)
 * Date: July 24, 2024
 * Copyright: © 2024 Alex Torma
 * License: MIT
 * Version: 1
 *
 */
'use strict';

const createElement = ({ tag, ...props } = {}) => {
  const element = document.createElement(tag);
  Object.entries(props).forEach(([key, value]) => {
    if (key && value !== undefined) {
      element[key] = value;
    }
  });
  return element;
};

const getDivs = () => ({
  patternsDiv: document.getElementById('patterns'),
  prefixesDiv: document.getElementById('prefixes')
});

const getInputs = () => ({
  patternInputs: document.querySelectorAll('#patterns input'),
  prefixInputs: document.querySelectorAll('#prefixes input')
});

const clearDivs = (patternsDiv, prefixesDiv) => {
  patternsDiv.innerHTML = '';
  prefixesDiv.innerHTML = '';
};


const initializeInputFields = () => {
  const { patternsDiv, prefixesDiv } = getDivs();
  clearDivs(patternsDiv, prefixesDiv);

  api.storage.local.get(['patterns', 'prefixes'], ({ patterns = [], prefixes = [] }) => {
    // Helper function to initialize fields and ensure one empty field
    const initializeFields = (type, items) => {
      items.forEach(item => createInputField(type, item));
      ensureOneEmptyInputField(type);
    };

    initializeFields('pattern', patterns);
    initializeFields('prefix', prefixes);
  });
};

const createInputField = (type, value = '') => {
  const capitalizeFirstLetter = (string) => string.charAt(0).toUpperCase() + string.slice(1);
  
  const { patternsDiv, prefixesDiv } = getDivs();
  const placeholder = `Enter URL ${capitalizeFirstLetter(type)}`;
  
  const newInput = createElement({ tag: 'input', type: 'text', placeholder, value });
  const inputWrapper = createElement({ tag: 'div', className: type });
  
  newInput.addEventListener('input', handleInput);
  inputWrapper.appendChild(newInput);
  
  const targetDiv = type === 'pattern' ? patternsDiv : prefixesDiv;
  targetDiv.appendChild(inputWrapper);
};

const handleInput = (event) => {
  const inputField = event.target;
  const type = inputField.parentNode.classList.contains('pattern') ? 'pattern' : 'prefix';
  
  const { patternInputs, prefixInputs } = getInputs();
  
  if (inputField.value.trim()) {
    if ((type === 'pattern' && patternInputs[patternInputs.length - 1] === inputField) ||
        (type === 'prefix' && prefixInputs[prefixInputs.length - 1] === inputField)) {
      createInputField(type);
    }
  } else if ((type === 'pattern' && patternInputs.length > 1) ||
             (type === 'prefix' && prefixInputs.length > 1)) {
    inputField.parentElement.remove();
  }
  
  ensureOneEmptyInputField(type);
  savePatterns();
};

const ensureOneEmptyInputField = (type) => {
  const { patternsDiv, prefixesDiv } = getDivs();
  const parentDiv = type === 'pattern' ? patternsDiv : prefixesDiv;
  const inputFields = parentDiv.querySelectorAll('input');
  
  if (inputFields.length === 0 || inputFields[inputFields.length - 1].value.trim()) {
    createInputField(type);
  }
};

const savePatterns = () => {
  const { patternInputs, prefixInputs } = getInputs();
  
  const patterns = Array.from(patternInputs).map(field => field.value.trim()).filter(Boolean);
  const prefixes = Array.from(prefixInputs).map(field => field.value.trim()).filter(Boolean);
  
  api.storage.local.set({ patterns, prefixes });
};

const api = typeof chrome !== undefined ? chrome : browser
document.addEventListener('DOMContentLoaded', () => {
  initializeInputFields()
});
