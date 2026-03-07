# Plan: Cloudflare Worker CORS Proxy for URL Fetching

**Date:** 2026-03-07  
**Research:** research/2026-03-07-cloudflare-markdown-for-agents.md

## Summary

Deploy a standalone Cloudflare Worker as a CORS proxy so the fabbro web app can fetch arbitrary URLs (not just GitHub). The Worker requests content with `Accept: text/markdown`, forwards the response body and key headers back to the browser with permissive CORS headers, and applies basic security controls. The client-side `fetchMarkdown()` is updated to route through the proxy when direct fetch fails.

## Context

The fabbro web SPA currently cannot fetch most arbitrary URLs because browsers enforce CORS — only GitHub API URLs (which return `Access-Control-Allow-Origin: *`) work reliably. The research document evaluated four options and recommended Cloudflare Pages + Functions (Option A). This plan implements **Option B (standalone Worker)** instead because:

- It keeps the existing GitHub Pages deployment untouched — no migration risk
- The Worker is a single file, independently deployable, and free-tier (100K req/day)
- It's the smallest possible change to unlock URL fetching
- If the project later migrates to Cloudflare Pages, the Worker logic can be inlined as a Pages Function

### Tension with local-first philosophy

This introduces a hosted proxy. However:
- The proxy is **stateless** — it fetches on behalf of the user and returns the response. No data is stored.
- The proxy is **optional** — GitHub URLs and paste/file-drop work without it. Direct CORS-friendly URLs still work directly.
- The GitHub API is already an external dependency in the same spirit.

## Design

### Worker behavior

```
GET https://fabbro-proxy.{account}.workers.dev/?url={encoded_url}

1. Validate `url` param exists and is a valid HTTP(S) URL
2. Reject non-HTTP schemes, private IPs, localhost
3. Fetch the target with `Accept: text/markdown, text/html`
4. Return body with CORS headers + forwarded Content-Type + x-markdown-tokens
```

### CORS preflight

```
OPTIONS → 204 with Access-Control-Allow-Origin: *
```

### Security controls

| Control | Implementation |
|---------|---------------|
| Origin allowlist | Check `Origin` header against configured allowed origins (e.g., `charly-vibes.github.io`, `localhost`) |
| URL scheme validation | Only allow `https://` targets (reject `http://`, `file://`, `ftp://`, etc.) |
| Private IP blocking | Reject targets resolving to `10.x`, `172.16-31.x`, `192.168.x`, `127.x`, `::1` |
| Response size limit | Abort if response body exceeds 2 MB |
| Rate limiting | Cloudflare Worker free tier has implicit 100K/day; consider adding per-IP rate limiting later |

### Client-side changes

Update `fetchMarkdown()` in `web/fetch.js` to:

1. Try direct `fetch()` first (current behavior)
2. On CORS `TypeError`, retry through the proxy: `fetch(PROXY_URL + "?url=" + encodeURIComponent(url))`
3. Surface `x-markdown-tokens` count and `Content-Type` in the returned result object

### Response metadata

Extend the `fetchContent()` return value to include:

```javascript
{
  content: "...",
  source: "...",
  filename: "...",
  contentType: "text/markdown" | "text/html" | "text/plain",
  markdownTokens: 3150 | null,   // from x-markdown-tokens header
  proxied: true | false,
}
```

The UI can later use `contentType` to warn when raw HTML was returned and `markdownTokens` to display estimated token count.

## Implementation

### Phase 1: Cloudflare Worker (new repo/directory)

**Files created:**
- `worker/wrangler.toml` — Worker configuration
- `worker/src/index.ts` — Worker handler
- `worker/package.json` — dependencies (wrangler only)
- `worker/tsconfig.json` — TypeScript config
- `worker/README.md` — setup & deployment instructions

**Steps:**

1. Scaffold `worker/` directory with wrangler config
2. Implement the fetch handler:
   - Parse `url` query param
   - Validate URL scheme (https only)
   - Handle OPTIONS preflight
   - Check Origin against allowlist (configurable via `wrangler.toml` vars)
   - Fetch target with `Accept: text/markdown, text/html`
   - Cap response body at 2 MB
   - Return response with CORS headers, forwarded `Content-Type`, `x-markdown-tokens`
3. Add error handling (invalid URL → 400, target error → 502, oversized → 413)

### Phase 2: Client-side proxy fallback

**Files changed:**
- `web/fetch.js` — add proxy fallback + metadata

**Steps:**

1. Add `PROXY_URL` constant (can be overridden for local dev)
2. In `fetchMarkdown()`, catch CORS `TypeError` and retry via proxy
3. Read `Content-Type` and `x-markdown-tokens` from response headers
4. Extend return object from `fetchContent()` with `contentType`, `markdownTokens`, `proxied`

### Phase 3: UI feedback (optional, separate plan)

Surface the new metadata in the toolbar/status bar:
- Show "via proxy" indicator when proxied
- Show estimated token count when available
- Warn when raw HTML was returned

This phase is out of scope for this plan.

## Verification

- `wrangler dev` → Worker starts locally
- `curl 'http://localhost:8787/?url=https://blog.cloudflare.com/markdown-for-agents/' -H 'Origin: http://localhost'` → returns markdown content with CORS headers
- `curl 'http://localhost:8787/?url=file:///etc/passwd'` → 400 error
- `curl 'http://localhost:8787/?url=http://192.168.1.1'` → 400 error  
- `curl 'http://localhost:8787/'` → 400 "Missing url param"
- In the web app: paste a non-GitHub URL → content loads via proxy fallback
- In the web app: paste a GitHub URL → still uses GitHub API directly (no regression)
- Paste text / drop file → still works (no regression)

## Out of scope

- Migrating from GitHub Pages to Cloudflare Pages
- Server-side `toMarkdown()` or Browser Rendering API integration
- HTML-to-text client-side fallback (separate enhancement)
- Per-IP rate limiting beyond Cloudflare's free tier limits
- Custom domain for the Worker
