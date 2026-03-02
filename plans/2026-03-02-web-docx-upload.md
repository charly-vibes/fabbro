# Plan: Web DOCX File Upload

**Issue:** fabbro-4gq  
**Spec:** specs/09_web_docx_upload.feature  
**Date:** 2026-03-02

## Summary

Add `.docx` file support to the fabbro web app's drag-and-drop zone. When a user drops a Word document, we extract the raw text client-side using mammoth.js and start a review session with it.

## Context

The webapp currently accepts text-based files (`.md`, `.txt`, `.fem`, code files) via a drop zone on the landing page. Files are read with `FileReader.readAsText()`. Word `.docx` files are binary (ZIP archives containing XML), so they require a library to extract text.

### Library choice: mammoth.js

- **`mammoth.extractRawText({arrayBuffer})`** — returns plain text, which is exactly what fabbro needs (not HTML)
- 6.1k GitHub stars, mature, BSD-2-Clause license
- Works client-side with `ArrayBuffer` input
- Standalone browser build available (`mammoth.browser.min.js`, ~150KB)
- No dependencies at runtime

### Dependencies

| Library | Version | License | Location |
|---------|---------|---------|----------|
| mammoth.js | 1.x (pin latest) | BSD-2-Clause | `web/vendor/mammoth.browser.min.js` |

The library is vendored into the repository to preserve fabbro's local-first, offline-capable philosophy. No CDN dependency.

### mammoth.extractRawText behavior

mammoth separates each paragraph with **two newlines** (a blank line between paragraphs). This is acceptable for fabbro's review use case — it provides visual separation between paragraphs in the viewer.

## Implementation

**Files changed:**
- `web/vendor/mammoth.browser.min.js` — vendored library (new)
- `web/app.html` — load mammoth via `<script>` tag
- `web/app.js` — drop handler changes

**Steps:**

1. **Vendor mammoth.js**: Download `mammoth.browser.min.js` and place in `web/vendor/`.

2. **Load mammoth in app.html**: Add a `<script>` tag before the module script:
   ```html
   <script src="vendor/mammoth.browser.min.js"></script>
   ```
   This sets `window.mammoth` as a global. mammoth's browser build is UMD/IIFE, not ESM, so it cannot be loaded via `import()`.

3. **Handle .doc before the acceptedExts guard** in the drop handler:
   - Check if extension is `.doc` *before* the `acceptedExts` check
   - Show: `"Legacy .doc files are not supported. Please save as .docx."`
   - Return early

4. **Add `.docx` to the `acceptedExts` Set**.

5. **Update drop zone label**: Change from `"Drop a .md, .txt, .fem, or code file here"` to `"Drop a .md, .txt, .docx, .fem, or code file here"`.

6. **Handle .docx in the drop handler**: Add a branch alongside the existing `.fem` handler:
   - When extension is `.docx`, read file as `ArrayBuffer` (not text)
   - Call `window.mammoth.extractRawText({arrayBuffer})` to get plain text
   - Normalize line endings (`\r\n` → `\n`) and trim trailing whitespace
   - Check if result is empty/whitespace-only → show `"This document appears to be empty."`
   - Otherwise set `session.content` and call `startSession()`

7. **Handle extraction errors**: Wrap mammoth call in try/catch, show: `"Could not read .docx file. The file may be corrupt."`

## Verification

- Drop a `.docx` file → text appears in editor, paragraphs separated by blank lines
- Drop a `.doc` file → specific error message (before acceptedExts check)
- Drop a corrupt `.docx` → graceful error message
- Drop an empty `.docx` → "This document appears to be empty."
- Drop zone label mentions `.docx`
- Existing file types (`.md`, `.txt`, `.fem`, code) continue to work unchanged
- Open the app offline with vendored mammoth → .docx support works

## Out of scope

- Rich text / formatting preservation (headings as `#`, bold as `**`, etc.)
- Image extraction from .docx
- Table extraction from .docx
- Export back to .docx format
