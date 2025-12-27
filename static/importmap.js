// Import map setup - shared between index.html and key.html
// This file centralizes dependency version management for ES modules
//
// To add a new dependency:
// 1. Add it to the importMap object below
// 2. Update the CSP script hash in handlers/index.go if the inline script content changes
//
// To update a version:
// 1. Change the URL in the importMap
// 2. Test that all imports still work
const importMap = {
    "imports": {
        "preact": "https://esm.sh/preact@10.28.1",
        "preact/hooks": "https://esm.sh/preact@10.28.1/hooks"
    }
};

// Add import map to document head
const script = document.createElement('script');
script.type = 'importmap';
script.textContent = JSON.stringify(importMap, null, 4);
document.head.appendChild(script);