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
