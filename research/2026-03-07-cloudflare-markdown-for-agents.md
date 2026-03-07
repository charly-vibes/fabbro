# Cloudflare Markdown for Agents — Research for Fabbro Web

**Date:** 2026-03-07  
**Topic:** Using Cloudflare's "Markdown for Agents" feature for fabbro's web version

## What Is Markdown for Agents?

Cloudflare's **Markdown for Agents** (launched Feb 12, 2026) is an edge feature that automatically converts HTML pages to Markdown on-the-fly using HTTP content negotiation. When a client sends `Accept: text/markdown`, Cloudflare intercepts the origin HTML response and converts it to clean Markdown before serving.

**Key stats:** 80% token reduction (16,180 HTML tokens → 3,150 Markdown tokens for a typical blog post).

## How It Works

1. Client sends request with `Accept: text/markdown` header
2. Cloudflare fetches original HTML from origin
3. Edge converts HTML → Markdown automatically
4. Response includes:
   - `Content-Type: text/markdown; charset=utf-8`
   - `x-markdown-tokens: <count>` (estimated token count)
   - `Content-Signal: ai-train=yes, search=yes, ai-input=yes`
   - `Vary: accept`

### Usage (Client-Side)

```bash
curl https://example.com/page/ -H "Accept: text/markdown"
```

```javascript
const r = await fetch(url, {
  headers: { Accept: "text/markdown, text/html" }
});
const tokenCount = r.headers.get("x-markdown-tokens");
const markdown = await r.text();
```

### Enabling (Server-Side)

- **Dashboard:** Zone → AI Crawl Control → toggle "Markdown for Agents"
- **API:** `PATCH /client/v4/zones/{zone_tag}/settings/content_converter` with `{"value": "on"}`
- **Per-path:** Configuration Rules with expressions like `starts_with(http.request.uri.path, "/blog/")`
- **Plans:** Pro, Business, Enterprise (no extra cost, Beta)

## Other Cloudflare Markdown APIs

| Method | Use Case | How |
|--------|----------|-----|
| **Markdown for Agents** | Real-time edge conversion for enabled zones | `Accept: text/markdown` header |
| **Workers AI `toMarkdown()`** | Convert arbitrary documents (PDF, DOCX, images, CSV, HTML) | `env.AI.toMarkdown(files)` binding or REST API |
| **Browser Rendering `/markdown`** | Convert JS-heavy/SPA pages (renders in headless browser first) | REST API: `POST /browser-rendering/markdown` |

### Workers AI `toMarkdown()` — Most Versatile

```javascript
// Worker binding
const results = await env.AI.toMarkdown([
  { name: "doc.pdf", blob: new Blob([buffer], { type: "application/pdf" }) },
  { name: "page.html", blob: new Blob([html], { type: "text/html" }) },
]);
// result: [{ name, mimeType, format: "markdown", tokens, data }]
```

```bash
# REST API
curl https://api.cloudflare.com/client/v4/accounts/{ACCOUNT_ID}/ai/tomarkdown \
  -H 'Authorization: Bearer {API_TOKEN}' \
  -F "files=@document.pdf"
```

Supported formats: PDF, images (JPEG/PNG/WebP/SVG), HTML, XML, Office docs (XLSX/DOCX), ODF, CSV, Apple Numbers.

### Browser Rendering `/markdown`

```bash
curl -X POST 'https://api.cloudflare.com/client/v4/accounts/{id}/browser-rendering/markdown' \
  -H 'Authorization: Bearer {token}' \
  -d '{"url": "https://example.com"}'
```

Options: `rejectRequestPattern`, `gotoOptions.waitUntil: "networkidle0"` for SPAs, custom `userAgent`.

## Relevance to Fabbro Web

### Current State

Fabbro's web app (`web/`) is a vanilla JS SPA that:
- Fetches content from URLs via `fetch.js`
- Renders text line-by-line in a plain-text code viewer (`viewer.js`) — there is no HTML rendering or stripping
- Allows annotation/review of text content
- Supports: GitHub raw files, plain text URLs, pasted text, file drops (.md, .txt, .docx, code files)
- **Already sends `Accept: text/markdown, text/html`** in `fetchMarkdown()` (fetch.js line 42)

Fabbro's docs site (`site/`) uses Hugo with the hugo-book theme, hosted on GitHub Pages at `charly-vibes.github.io/fabbro/docs/`.

Fabbro is explicitly a **local-first tool with no server components** (per project philosophy in AGENTS.md).

### What Fabbro Already Does

The `fetchMarkdown()` function already requests markdown via content negotiation:

```javascript
res = await fetch(url, {
  headers: { Accept: "text/markdown, text/html" },
});
```

This means fabbro already benefits from Markdown for Agents on any Cloudflare-enabled site that has the feature toggled on. No code changes needed for the basic integration.

### Key Constraint: CORS

The biggest practical limitation for URL fetching in a browser SPA is **CORS (Cross-Origin Resource Sharing)**, not content format. Most websites block cross-origin `fetch()` requests from browsers. Fabbro already handles this with a specific error message in `fetchMarkdown()`. Cloudflare's Markdown for Agents does not change CORS behavior — if a site blocks cross-origin requests, fabbro still can't fetch it regardless of the `Accept` header.

The effective scope where Markdown for Agents helps fabbro is narrow: URLs that are both (a) CORS-accessible and (b) behind Cloudflare with the feature enabled. GitHub URLs bypass this entirely since fabbro uses the GitHub API for raw content.

### Potential Improvements (Local-First Compatible)

#### 1. **Detect and surface markdown responses**

Since fabbro already sends the header but doesn't check what came back, we could:

- Check `Content-Type` response header for `text/markdown` vs `text/html`
- Read `x-markdown-tokens` when present and display estimated token count in the UI
- Warn the user when raw HTML was returned (since the viewer shows it as plain text, which is unreadable)

This is purely client-side and aligns with local-first philosophy.

#### 2. **Client-side HTML-to-text fallback**

When a non-Cloudflare site returns HTML, fabbro currently displays raw HTML tags in the viewer. A lightweight client-side HTML-to-text conversion (e.g., using DOMParser + textContent extraction) could improve this without adding server dependencies.

#### 3. **Make fabbro docs agent-friendly (out of scope today)**

If fabbro's docs site moved to Cloudflare Pages (currently GitHub Pages), enabling Markdown for Agents would let AI agents consume documentation efficiently. This would require moving to a custom domain through Cloudflare.

### Out of Scope (Conflicts with Local-First Philosophy)

The following Cloudflare APIs require server-side credentials and would introduce a hosted backend component, which conflicts with fabbro's explicit "no server components" design:

- **Workers AI `toMarkdown()`** — requires Cloudflare account + API token, runs server-side
- **Browser Rendering `/markdown`** — requires API credentials, runs headless browser on Cloudflare infra
- **Any "fetch proxy" Worker** — introduces a hosted dependency, billing, abuse surface, and privacy concerns

These would only become relevant if the project philosophy explicitly changes to allow hosted components.

## Limitations

- Markdown for Agents only converts HTML (not PDFs, images, etc.)
- Origin response cannot exceed 2 MB
- Only available on Cloudflare Pro+ plans (for site owners enabling it)
- **CORS is the dominant barrier** for browser-based SPA fetching — Markdown for Agents doesn't help if the target blocks cross-origin requests
- If a site returns HTML (no Markdown for Agents), fabbro's viewer displays raw HTML tags with no conversion
- `toMarkdown()` and Browser Rendering APIs require server-side credentials (not usable from a pure client-side app)

## Why `fetchMarkdown` Fails on GitHub Pages (CORS Deep Dive)

### The Problem

When deployed to GitHub Pages (`charly-vibes.github.io`), `fetchMarkdown()` fails for almost all arbitrary URLs. The `Accept: text/markdown` header is sent correctly, but the browser blocks the response due to CORS.

### Verified Behavior (tested 2026-03-07)

| Target | `Access-Control-Allow-Origin` | Markdown served? | Works from GH Pages? |
|--------|-------------------------------|-------------------|----------------------|
| GitHub API (`api.github.com`) | `*` | N/A (raw content) | ✅ Yes |
| `raw.githubusercontent.com` | `*` | N/A (raw text) | ✅ Yes |
| `blog.cloudflare.com` | `https://dash.cloudflare.com` | ✅ Yes (on GET) | ❌ No — CORS blocks |
| `developers.cloudflare.com` | `*` | ❌ No (returns HTML) | ⚠️ CORS OK, but no markdown |
| Most arbitrary websites | None | N/A | ❌ No — CORS blocks |

**Key finding:** Even Cloudflare's own blog serves markdown correctly but restricts CORS to `dash.cloudflare.com`, so a browser on GitHub Pages cannot read the response.

### Why GitHub URLs Work

`fetchGitHub()` uses `api.github.com` which returns `Access-Control-Allow-Origin: *`, allowing any browser origin. This path completely bypasses the CORS problem.

### Root Cause

CORS is a **server-side** policy. The target website must explicitly allow cross-origin reads via `Access-Control-Allow-Origin`. Most websites either:
- Send no CORS headers (browser blocks)
- Restrict to specific origins (e.g., Cloudflare restricts to their own dashboard)
- Return `*` (rare for non-API endpoints)

Fabbro's `fetchMarkdown()` correctly catches this as a `TypeError` and shows a helpful error message. The code is correct — the limitation is inherent to browser SPAs.

### Solutions Evaluated

#### Option A: Cloudflare Pages + Functions (Recommended)

Move fabbro's web deployment from GitHub Pages to **Cloudflare Pages**. Add a single serverless Function as a CORS proxy:

```
web/
  functions/
    api/
      fetch.ts    ← Cloudflare Pages Function
```

```typescript
// functions/api/fetch.ts
export const onRequest: PagesFunction = async (context) => {
  const url = new URL(context.request.url).searchParams.get("url");
  if (!url) return new Response("Missing url param", { status: 400 });

  const res = await fetch(url, {
    headers: { Accept: "text/markdown, text/html" },
  });
  const body = await res.text();

  return new Response(body, {
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Content-Type": res.headers.get("Content-Type") || "text/plain",
      "X-Markdown-Tokens": res.headers.get("x-markdown-tokens") || "",
    },
  });
};
```

Client-side `fetchMarkdown()` would call `/api/fetch?url=...` instead of the target directly.

**Pros:**
- Free tier: 100K requests/day, static assets unlimited
- Same deployment pipeline (git push → auto deploy)
- No external dependencies, no API keys to manage
- The proxy runs on Cloudflare's edge (same infra as Markdown for Agents)
- Can also leverage `toMarkdown()` for HTML→Markdown conversion server-side

**Cons:**
- Introduces a server component (conflicts with strict local-first philosophy)
- Fabbro becomes dependent on Cloudflare infrastructure
- Proxy could be abused (needs origin validation + rate limiting)

**Tension with local-first:** This is a _minimal_ server component — a thin proxy that doesn't store data. All session state remains client-side in IndexedDB. The proxy only fetches content on behalf of the user. This is comparable to how the GitHub API already acts as a third-party dependency.

#### Option B: Self-Hosted Cloudflare Worker (Standalone CORS Proxy)

Deploy a standalone Worker at e.g. `fabbro-proxy.workers.dev`:

```javascript
export default {
  async fetch(request) {
    const url = new URL(request.url).searchParams.get("url");
    if (!url) return new Response("Missing url", { status: 400 });

    if (request.method === "OPTIONS") {
      return new Response(null, {
        status: 204,
        headers: {
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, OPTIONS",
          "Access-Control-Allow-Headers": "Accept",
          "Access-Control-Max-Age": "86400",
        },
      });
    }

    const res = await fetch(url, {
      headers: { Accept: "text/markdown, text/html" },
    });

    const response = new Response(res.body, {
      status: res.status,
      headers: {
        "Access-Control-Allow-Origin": "*",
        "Content-Type": res.headers.get("Content-Type") || "text/plain",
      },
    });
    return response;
  },
};
```

**Pros:** Keeps GitHub Pages deployment, free Worker tier (100K req/day)
**Cons:** Two separate deployments to manage, still a server dependency

#### Option C: Third-Party CORS Proxies

Use services like `corsproxy.io`, `allorigins.win`, or `corsfix.com`:

```javascript
const content = await fetch(
  `https://corsproxy.io/?url=${encodeURIComponent(url)}`
).then(r => r.text());
```

**Pros:** Zero deployment, zero cost, immediate
**Cons:**
- Security risk: third party reads all traffic
- Unreliable (free services get shut down)
- Can't pass `Accept: text/markdown` through most proxies
- Never suitable for production

#### Option D: Keep Current Behavior (Status Quo)

Accept that arbitrary URL fetching only works for:
- GitHub file URLs (via API)
- The rare CORS-enabled site
- Pasted text / file drops

**Pros:** Zero complexity, fully local-first
**Cons:** Limited URL fetching capability in the web version

### Recommendation

**Option A (Cloudflare Pages + Functions)** is the best fit because:

1. It's the minimum viable server component — a stateless proxy, not a backend
2. Free tier is generous (100K req/day)
3. Single deployment (Pages serves static + functions together)
4. Can opportunistically benefit from Markdown for Agents on the same network
5. GitHub Pages → Cloudflare Pages migration is straightforward

If strict local-first purity is paramount, **Option D** is the honest choice — accept the limitation and focus on the paths that work (GitHub URLs, paste, file drops).

## References

- [Blog announcement](https://blog.cloudflare.com/markdown-for-agents/)
- [Developer docs](https://developers.cloudflare.com/fundamentals/reference/markdown-for-agents/)
- [Workers AI toMarkdown()](https://developers.cloudflare.com/workers-ai/features/markdown-conversion/)
- [Browser Rendering /markdown](https://developers.cloudflare.com/browser-rendering/rest-api/markdown-endpoint/)
- [Cloudflare Pages Functions](https://developers.cloudflare.com/pages/functions/)
- [Cloudflare Pages CORS example](https://developers.cloudflare.com/pages/functions/examples/cors-headers/)
- [Building a CORS proxy with CF Workers](https://rednafi.com/javascript/cors_proxy_with_cloudflare_workers/)
- [CORS proxy security risks](https://httptoolkit.com/blog/cors-proxies/)
