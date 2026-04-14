# pi-fabbro

Project-local pi package scaffold for integrating the `fabbro` CLI into pi.

## Current scope

Current implementation:

- package structure for a pi package
- TypeScript extension entrypoint
- runtime `fabbro` availability check
- `fabbro_prime` tool and `/fabbro-prime` command
- `fabbro_create_review` tool and `/fabbro-review` command
- local loading and testing notes

Next phases will add feedback retrieval and session discovery.

## Files

- `package.json` — pi package manifest
- `extensions/fabbro.ts` — extension entrypoint
- `scripts/smoke-test.sh` — repeatable print-mode verification

## LLM / tool surface

The custom tools exposed to pi are:

- `fabbro_prime` — load the current machine-readable fabbro primer
- `fabbro_create_review` — create a non-interactive fabbro review session from text and return the session ID

## Load locally

From the `fabbro/` repo root:

```bash
pi -e ./pi-fabbro
```

Or install it into project-local pi settings:

```bash
pi install -l ./pi-fabbro
```

Because this package uses the pi package manifest, pi will load the extension from `./extensions`.

## Verify the extension

Quote the whole slash command when using `pi -p` so pi receives it as a single prompt.

### 1. Confirm the extension loads

```bash
pi -e ./pi-fabbro -p '/fabbro-help'
```

### 2. Confirm a usable `fabbro` command is available

```bash
pi -e ./pi-fabbro -p '/fabbro-status'
```

If the `fabbro` binary on `PATH` is older and does not support `--no-interactive`, the extension falls back to `go run ./cmd/fabbro` when your current working directory is the fabbro repo root or a subdirectory inside this repo.

### 3. Print the machine-readable primer

```bash
pi -e ./pi-fabbro -p '/fabbro-prime'
```

### 4. Create a review session from text

```bash
pi -e ./pi-fabbro -p '/fabbro-review Example review content from pi extension test.'
```

`/fabbro-review` requires inline text after the command. Expected result: pi prints the new session ID and tells you to run `fabbro session resume <id>` outside pi.

### 5. Confirm the missing-binary message is clear

Run from a directory that is not the fabbro repo, with a PATH that does not contain `fabbro`:

```bash
env PATH=/usr/bin:/bin /home/linuxbrew/.linuxbrew/bin/pi -e /absolute/path/to/pi-fabbro -p '/fabbro-status'
```

Expected result: pi stays up, the extension loads, and `/fabbro-status` reports that `fabbro` is not available.

### 6. Run the smoke test

From the repo root:

```bash
cd pi-fabbro && npm run smoke
```

## Notes

- pi loads extensions through `jiti`, so no TypeScript build step is required.
- This extension intentionally does not embed fabbro's TUI.
- The v1 integration surface stays CLI-first.
