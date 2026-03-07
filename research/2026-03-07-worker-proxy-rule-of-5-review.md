# Research: Rule of 5 Review — Cloudflare Worker CORS Proxy Plan

**Date:** 2026-03-07  
**Status:** Complete  
**Methodology:** Rule of 5 iterative review (5 passes: Architecture & Philosophy, Security & Abuse, Technical Feasibility, Operations & Maintenance, Scope & Prioritization)  
**Convergence:** Converged after pass 4; pass 5 reinforced existing themes with no new CRITICAL issues  
**Plan under review:** `plans/2026-03-07-cloudflare-worker-cors-proxy.md`

---

## Executive Summary

The plan is technically workable but **materially conflicts with fabbro's "no server components" local-first philosophy** and is **under-secured as written** — it's close to an open proxy. The plan frames a policy change as a small feature.

### Top 3 Findings

1. **ARCH-001 (CRITICAL):** Introduces an always-on server dependency into a product that explicitly has "no server components"
2. **SEC-001 (CRITICAL):** Origin allowlist + permissive CORS is not an abuse control — non-browser clients ignore it entirely, making this an open fetch relay
3. **SEC-002 (CRITICAL):** Private IP blocking as specified is incomplete and likely bypassable (DNS rebinding, redirect chains)

### Net Recommendation

**Decide policy first, then implementation.** Either:
- **Stay strict local-first** → no hosted proxy; ship client-only improvements (content-type detection, HTML-to-text fallback) — these deliver 60–80% of the UX win
- **Allow a proxy** → make it **BYO/self-hosted and off-by-default**; ship the Worker code + docs but don't run a hosted instance

---

## Issue Summary

| Severity | Count | Categories |
|----------|-------|------------|
| CRITICAL | 4 | Architecture (1), Security (2), Operations (1) |
| HIGH | 5 | Architecture (1), Security (2), Technical (2) |
| MEDIUM | 5 | Architecture (1), Security (1), Technical (2), Scope (1) |
| LOW | 1 | Technical |

---

## PASS 1: Architecture & Philosophy Alignment

*Does this conflict with local-first? Is the proxy justified?*

### ARCH-001: Introduces an always-on server dependency into a "no server components" product (CRITICAL)

**Location:** Plan Summary + "Tension with local-first philosophy"

**Evidence:** AGENTS.md is explicit: "There are **no server components**." The research document itself flags "Any fetch proxy Worker" as out-of-scope / conflicting with local-first philosophy (research L135–L143).

**Impact:**
- fabbro web becomes dependent on a maintained Cloudflare account + infra
- Introduces privacy expectations ("my content goes through your server") even if stateless
- Creates an implicit SLA and failure mode that didn't exist

**Recommendation:** Reframe the plan to preserve philosophy:
- Make the Worker **Bring-Your-Own Proxy**: ship Worker code + docs, but do not run a hosted instance
- In the SPA, support `PROXY_URL` override but default to no proxy
- Add explicit UX copy when a proxy is configured: "Requests will be sent to \<proxy\>"

---

### ARCH-002: "Proxy is optional" is true technically but false in user value terms (HIGH)

**Location:** "proxy is optional" argument

**Evidence:** For the core feature ("fetch arbitrary URLs"), the proxy becomes the primary path; direct fetch works only for GitHub URLs and rare CORS-enabled sites (research L109–L183).

**Impact:** Users will experience this as "the app needs a server to do URL fetching," undermining local-first messaging.

**Recommendation:** Treat proxy support as an advanced integration, not a default capability. Position: "fabbro can fetch GitHub + pasted + dropped files locally; arbitrary URL fetch requires your own proxy."

---

### ARCH-003: Plan justification prioritizes "avoid migration" over architectural coherence (MEDIUM)

**Location:** Option B rationale

**Evidence:** Research recommended Option A (Pages+Functions) and stated strict local-first ⇒ Option D "honest choice" (research L303–L314). The plan chooses B mostly to avoid moving off GitHub Pages.

**Recommendation:** Update decision framing: first decide whether any hosted proxy is acceptable, then choose Worker vs Pages.

---

## PASS 2: Security & Abuse

*Open proxy risks, SSRF, abuse surface*

### SEC-001: Origin allowlist + permissive CORS is not an abuse control (CRITICAL)

**Location:** Security controls table, preflight section

**Evidence:** CORS/Origin checks only influence browser behavior. Anyone can `curl https://worker/?url=...` with no Origin or a spoofed Origin.

**Impact:** The Worker is effectively a public fetch relay:
- Can be used to proxy scraping, ToS-violating access, or illegal content retrieval
- Can be used for bandwidth abuse and to burn free-tier quotas
- Can amplify traffic to third parties from your account

**Recommendation:**
- Require an application-level token — do not ship a shared token in the public SPA
- If you insist on no auth, revert to BYO/self-host and accept that each user secures their own instance
- At minimum: reject requests with missing/invalid Origin *and* require `Sec-Fetch-Site: cross-site`

---

### SEC-002: Private IP blocking as specified is incomplete and likely bypassable (CRITICAL)

**Location:** Security controls table, validation steps

**Evidence:** In Workers, you typically only have the hostname. Trustworthy DNS resolution + rebinding-safe checks across redirects is non-trivial.

**Impact:** If implemented naïvely (only block IP-literals like `http://192.168…`), it still allows:
- DNS rebinding / domains that resolve to private ranges
- Redirect chains to private hosts

**Recommendation:**
- Block IP-literals in private ranges, `localhost`, `.local`, non-443 ports
- Set `redirect: "manual"` and enforce redirect policy
- Document that private-IP blocking is best-effort

---

### SEC-003: Preflight section conflicts with origin allowlist (HIGH)

**Location:** Preflight returns `Access-Control-Allow-Origin: *` while allowlist checks Origin

**Evidence:** These are contradictory approaches.

**Recommendation:** Always echo back the validated Origin, never `*`. If Origin is absent or not allowed, return 403.

---

### SEC-004: URL-in-query leaks sensitive browsing targets via logs/analytics/referrers (HIGH)

**Location:** Worker endpoint design (`GET /?url=...`)

**Evidence:** Query strings are captured in Cloudflare request logs/analytics, browser history, and potentially Referer headers. Users may fetch internal docs, incident reports, etc.

**Impact:** Privacy concern is especially sharp for a local-first tool.

**Recommendation:**
- Prefer **POST** with JSON `{ "url": "..." }` to keep URLs out of query strings
- Add `Cache-Control: no-store` on responses
- Use `referrerPolicy: "no-referrer"` in the SPA when calling the proxy

---

### SEC-005: Response/header forwarding is underspecified (MEDIUM)

**Location:** "forwards the response body and key headers"

**Evidence:** Forwarding without an explicit allowlist could forward `Set-Cookie`, `Location`, or security headers.

**Recommendation:** Hard-allowlist response headers: `Content-Type`, `x-markdown-tokens`. Do **not** forward `Set-Cookie`, `Cookie`, `Authorization`, `CF-*`.

---

## PASS 3: Technical Feasibility & Complexity

*Is the implementation realistic? Edge cases?*

### TECH-001: "Catch CORS TypeError then retry proxy" is too broad (HIGH)

**Location:** Client-side changes; current `fetch.js` L44–L51

**Evidence:** In browsers, `fetch()` rejects with `TypeError` for CORS *and* for DNS failure, offline, TLS failures, etc.

**Impact:** The app may route ordinary outages through the proxy unnecessarily and produce confusing error messages.

**Recommendation:** Change messaging: "Direct fetch failed (CORS or network). Trying proxy…". If proxy also fails, surface both failures distinctly.

---

### TECH-002: 2MB cap needs streaming implementation (HIGH)

**Location:** Security controls, Worker fetch handler

**Evidence:** Size enforcement must be done while reading the stream; `await res.text()` reads the whole body first.

**Recommendation:**
- Check `Content-Length` header first (fast reject if >2MB)
- Stream read with a running byte count; abort at >2MB and return 413

---

### TECH-003: Redirect handling is missing but essential (MEDIUM)

**Location:** Worker behavior description

**Evidence:** Many URLs redirect (http→https, canonicalization). Redirects can bypass scheme/private-host checks.

**Recommendation:** Set `redirect: "manual"` and implement a small redirect loop (max 5 hops), validating each URL.

---

### TECH-004: Content-type expectations are too loose (MEDIUM)

**Location:** Fetch target with `Accept: text/markdown, text/html`

**Evidence:** Many endpoints return binary (PDF), huge files, or `application/octet-stream`.

**Recommendation:** Enforce allowlist of response Content-Type prefixes: allow `text/*`, maybe `application/json` and `application/xml`. Reject others with 415.

---

### TECH-005: Standalone Worker adds a TypeScript toolchain for a 1-file script (LOW)

**Location:** Implementation Phase 1

**Recommendation:** Consider plain JS Worker to keep it lean. Not a blocker.

---

## PASS 4: Operations & Maintenance

*Deployment, monitoring, cost, who maintains it?*

### OPS-001: Plan creates an operational commitment without an owner or runbook (CRITICAL)

**Location:** Summary + Implementation steps

**Evidence:** No mention of: who owns the Cloudflare account, how keys/config are managed, how abuse incidents are handled, what happens when free-tier limits are exceeded.

**Impact:**
- Quota exhaustion → feature stops working for everyone
- Abuse → potential ToS issues / account enforcement
- Silent privacy expectations → reputational risk

**Recommendation:** Add a minimal ops section:
- Explicit owner
- "Panic button": env var to disable proxy globally (return 503)
- Documented rate-limit strategy
- Clear statement: no SLA; feature may be disabled if abused

---

### OPS-002: Two independent deploy surfaces increases drift and maintenance (HIGH)

**Location:** Option B rationale + new `worker/` tree

**Evidence:** GitHub Pages deploy is separate from Worker deploy.

**Recommendation:** Add CI check to ensure `PROXY_URL` + allowed origins match; add release checklist.

---

### OPS-003: Logging/privacy posture is not addressed (MEDIUM)

**Location:** Entire plan; implicit in `GET ?url=...`

**Impact:** Even "stateless" can still be "observability state." Conflicts with local-first expectations.

**Recommendation:** Document privacy posture; prefer POST body; disable unnecessary analytics; avoid logging full URLs.

---

## PASS 5: Scope & Prioritization

*Is this the right thing to build now? MVP cutline?*

### SCOPE-001: This is a "policy change" disguised as a small feature (reinforces ARCH-001)

**Location:** Plan framing as "smallest possible change"

**Evidence:** AGENTS.md philosophy; research explicitly warns that any proxy conflicts.

**Recommendation:** Gate behind an explicit project decision: either update philosophy ("web version may use a minimal stateless proxy") or keep philosophy and ship BYO proxy only.

---

### SCOPE-002: There are local-first improvements that deliver value without servers (HIGH from research)

**Location:** Not in plan; in research "Potential Improvements" (research L115–L130)

**Evidence:** You can improve URL fetching UX by:
- Detecting markdown responses (check `Content-Type`, read `x-markdown-tokens`)
- Client-side HTML-to-text fallback (DOMParser → textContent)

**Impact:** These deliver 60–80% of the UX win without any proxy.

**Recommendation (MVP cutline):**
- Do client-only improvements first
- Reassess proxy only if users still demand arbitrary URL fetch and accept trade-offs

---

### SCOPE-003: UI phase is marked optional, but consent/visibility is not optional (MEDIUM)

**Location:** UI feedback out-of-scope

**Recommendation:** If any proxy exists, the UI must minimally show "via proxy" — hiding it is a privacy/trust risk.

---

## Convergence Assessment

| Theme | Passes | Issues |
|-------|--------|--------|
| Local-first philosophy / product contract change | 1, 4, 5 | ARCH-001, OPS-001, SCOPE-001 |
| Proxy abuse risk / "CORS is not security" | 2, 4 | SEC-001, SEC-003, OPS-001 |
| Privacy / visibility / consent | 1, 2, 4, 5 | ARCH-002, SEC-004, OPS-003, SCOPE-003 |
| Redirect + validation correctness | 2, 3 | SEC-002, TECH-003 |
| Complexity hidden behind "simple Worker" | 3, 4 | TECH-002, OPS-002 |

| Pass | New CRITICAL | New HIGH | New Issues Total | False Positive Rate |
|------|-------------|----------|------------------|---------------------|
| 1 (Architecture) | 1 | 1 | 3 | 0% |
| 2 (Security) | 2 | 2 | 5 | 0% |
| 3 (Technical) | 0 | 2 | 5 | 10% |
| 4 (Operations) | 1 | 1 | 3 | 5% |
| 5 (Scope) | 0 | 0 | 3 (reinforcements) | 15% |

**Status: CONVERGED** after pass 4. Pass 5 yielded only reinforcements with higher false positive rate.

---

## Recommended Next Actions

### Immediate (Before implementing the plan)

1. **Make the policy decision:** Does fabbro allow a hosted server component for web? Update AGENTS.md accordingly
2. **Ship client-only improvements first:** Content-type detection + HTML-to-text fallback in `fetchMarkdown()` (S–M effort, <3h)
3. **If proxy proceeds:** Reframe as BYO/self-hosted, off-by-default, with `PROXY_URL` config

### If implementing the Worker (hardening required)

4. **SEC-001:** Use POST body for target URL, not query string
5. **SEC-002:** Manual redirect loop with per-hop validation
6. **SEC-003:** Echo validated Origin, never `*`
7. **SEC-005:** Hard-allowlist forwarded response headers
8. **TECH-002:** Streaming size cap enforcement
9. **TECH-004:** Response content-type allowlist
10. **OPS-001:** Add ops section with owner, kill switch, rate-limit strategy

### Effort Summary

| Path | Effort |
|------|--------|
| Client-only improvements (no proxy) | S–M (1–3h) |
| BYO Worker + docs + SPA config | M (1–3h) |
| Shared hosted proxy hardened + ops/runbook | L (1–2d) |

---

## References

- Plan under review: `plans/2026-03-07-cloudflare-worker-cors-proxy.md`
- Research: `research/2026-03-07-cloudflare-markdown-for-agents.md`
- Rule of 5 methodology: Steve Yegge's "Six New Tips for Better Coding with Agents"
- Prior Rule of 5 code review: `research/2026-01-14-rule-of-5-code-review.md`
- Prior Rule of 5 UX review: `research/2026-02-11-web-ux-rule-of-5-review.md`
