import { spawn } from "node:child_process";
import { dirname, join } from "node:path";
import { existsSync } from "node:fs";

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { Type } from "@sinclair/typebox";

const STATUS_KEY = "fabbro";
const SESSION_ID_PATTERN = /^\d{8}-[a-f0-9]+$/i;
const MISSING_MESSAGE =
  "fabbro CLI not found on PATH. Install it and ensure `fabbro` is runnable before using the pi integration.";

type Level = "info" | "warning" | "error" | "success";

type ExecResult = {
  stdout: string;
  stderr: string;
  exitCode: number;
};

type FabbroCommand = {
  command: string;
  baseArgs: string[];
  display: string;
  cwd?: string;
};

type FabbroAvailability = {
  available: boolean;
  message: string;
  version?: string;
  resolvedCommand?: FabbroCommand;
};

class CancelledError extends Error {
  constructor() {
    super("Review creation cancelled");
    this.name = "CancelledError";
  }
}

const PATH_COMMAND: FabbroCommand = {
  command: "fabbro",
  baseArgs: [],
  display: "fabbro",
};

function repoCommand(repoRoot: string): FabbroCommand {
  return {
    command: "go",
    baseArgs: ["run", "./cmd/fabbro"],
    display: "go run ./cmd/fabbro",
    cwd: repoRoot,
  };
}

function isCancelledError(error: unknown): error is CancelledError {
  return error instanceof CancelledError;
}

function findRepoRoot(startCwd: string): string | null {
  let current = startCwd;

  while (true) {
    const hasGoMod = existsSync(join(current, "go.mod"));
    const hasMain = existsSync(join(current, "cmd", "fabbro", "main.go"));
    if (hasGoMod && hasMain) {
      return current;
    }

    const parent = dirname(current);
    if (parent === current) {
      return null;
    }
    current = parent;
  }
}

async function execWithPi(pi: ExtensionAPI, fabbro: FabbroCommand, args: string[]) {
  return pi.exec(fabbro.command, [...fabbro.baseArgs, ...args], {
    cwd: fabbro.cwd,
    timeout: 10_000,
  });
}

async function probeCommand(pi: ExtensionAPI, fabbro: FabbroCommand): Promise<FabbroAvailability | null> {
  try {
    const help = await execWithPi(pi, fabbro, ["review", "--help"]);
    const helpText = [help.stdout, help.stderr].join("\n");

    if (help.code !== 0) {
      return null;
    }

    if (!helpText.includes("--no-interactive")) {
      return {
        available: false,
        message: `${fabbro.display} is available but does not support --no-interactive yet.`,
      };
    }

    const versionResult = await execWithPi(pi, fabbro, ["--version"]);
    const version = [versionResult.stdout, versionResult.stderr].join("\n").trim();

    return {
      available: true,
      message: `${fabbro.display}${version ? ` (${version})` : ""}`,
      version: version || undefined,
      resolvedCommand: fabbro,
    };
  } catch {
    return null;
  }
}

async function detectFabbro(pi: ExtensionAPI, cwd: string): Promise<FabbroAvailability> {
  const pathStatus = await probeCommand(pi, PATH_COMMAND);
  if (pathStatus?.available) {
    return pathStatus;
  }

  const repoRoot = findRepoRoot(cwd);
  if (repoRoot) {
    const repoStatus = await probeCommand(pi, repoCommand(repoRoot));
    if (repoStatus?.available) {
      if (pathStatus && !pathStatus.available) {
        return {
          ...repoStatus,
          message: `Using ${repoStatus.resolvedCommand?.display} from ${repoRoot} because PATH fabbro is missing --no-interactive support.`,
        };
      }
      return repoStatus;
    }
  }

  if (pathStatus && !pathStatus.available) {
    return pathStatus;
  }

  return {
    available: false,
    message: MISSING_MESSAGE,
  };
}

async function ensureFabbro(pi: ExtensionAPI, cwd: string): Promise<FabbroAvailability> {
  const status = await detectFabbro(pi, cwd);
  if (!status.available || !status.resolvedCommand) {
    throw new Error(status.message);
  }
  return status;
}

async function execFabbroWithInput(
  cwd: string,
  fabbro: FabbroCommand,
  args: string[],
  input: string,
  signal?: AbortSignal,
): Promise<ExecResult> {
  return new Promise((resolve, reject) => {
    const child = spawn(fabbro.command, [...fabbro.baseArgs, ...args], {
      cwd: fabbro.cwd ?? cwd,
      stdio: ["pipe", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";
    let settled = false;

    const finish = (handler: () => void) => {
      if (settled) return;
      settled = true;
      handler();
    };

    child.stdout.on("data", (chunk: Buffer | string) => {
      stdout += chunk.toString();
    });

    child.stderr.on("data", (chunk: Buffer | string) => {
      stderr += chunk.toString();
    });

    child.on("error", (error) => finish(() => reject(error)));
    child.on("close", (code) => finish(() => resolve({ stdout, stderr, exitCode: code ?? 1 })));

    if (signal) {
      const onAbort = () => {
        child.kill("SIGTERM");
        finish(() => reject(new CancelledError()));
      };

      if (signal.aborted) {
        onAbort();
        return;
      }

      signal.addEventListener("abort", onAbort, { once: true });
      child.on("close", () => signal.removeEventListener("abort", onAbort));
    }

    child.stdin.write(input);
    child.stdin.end();
  });
}

function formatCommandFailure(fabbro: FabbroCommand, args: string[], result: ExecResult): string {
  const output = [result.stderr.trim(), result.stdout.trim()].filter(Boolean).join("\n");
  return output || `${fabbro.display} ${args.join(" ")} failed with exit code ${result.exitCode}`;
}

function extractSessionId(output: string): string {
  const sessionId = output.trim().split(/\s+/).pop() ?? "";
  if (!sessionId) {
    throw new Error("fabbro review did not return a session ID");
  }
  if (!SESSION_ID_PATTERN.test(sessionId)) {
    throw new Error(`fabbro review returned an unexpected session ID token: ${sessionId}`);
  }
  return sessionId;
}

async function getPrimeInfo(pi: ExtensionAPI, cwd: string) {
  const status = await ensureFabbro(pi, cwd);
  const result = await execWithPi(pi, status.resolvedCommand!, ["prime", "--json"]);
  if (result.code !== 0) {
    throw new Error((result.stderr || result.stdout || "fabbro prime failed").trim());
  }

  try {
    return JSON.parse(result.stdout);
  } catch (error) {
    const details = error instanceof Error ? error.message : String(error);
    throw new Error(`fabbro prime returned invalid JSON: ${details}`);
  }
}

async function createReviewSession(pi: ExtensionAPI, content: string, cwd: string, signal?: AbortSignal) {
  const status = await ensureFabbro(pi, cwd);
  const fabbro = status.resolvedCommand!;
  const trimmedContent = content.trim();

  if (!trimmedContent) {
    throw new Error("Review content is empty. Provide text after /fabbro-review or via the fabbro_create_review tool.");
  }

  const result = await execFabbroWithInput(cwd, fabbro, ["review", "--stdin", "--no-interactive"], content, signal);
  if (result.exitCode !== 0) {
    throw new Error(formatCommandFailure(fabbro, ["review", "--stdin", "--no-interactive"], result));
  }

  const sessionId = extractSessionId(result.stdout);
  return {
    sessionId,
    resumeCommand: `fabbro session resume ${sessionId}`,
    command: fabbro.display,
    contentPreview: trimmedContent.slice(0, 200),
    bytes: Buffer.byteLength(content, "utf8"),
  };
}

function printOrNotify(
  ctx: {
    hasUI: boolean;
    ui: {
      setStatus: (key: string, text: string) => void;
      notify: (text: string, level: Level) => void;
    };
  },
  message: string,
  level: Level,
) {
  if (ctx.hasUI) {
    ctx.ui.notify(message, level);
    return;
  }
  console.log(message);
}

function handleCommandError(error: unknown): never {
  if (isCancelledError(error)) {
    throw error;
  }
  throw error instanceof Error ? error : new Error(String(error));
}

export default function fabbroExtension(pi: ExtensionAPI) {
  pi.on("session_start", async (_event, ctx) => {
    const status = await detectFabbro(pi, ctx.cwd);

    if (status.available) {
      ctx.ui.setStatus(STATUS_KEY, status.message);
      return;
    }

    ctx.ui.setStatus(STATUS_KEY, "fabbro missing");
    ctx.ui.notify(status.message, "warning");
  });

  pi.registerTool({
    name: "fabbro_prime",
    label: "Fabbro Prime",
    description: "Fetch the machine-readable fabbro primer for agent workflows",
    promptSnippet: "Get the fabbro CLI primer and command overview as structured JSON.",
    promptGuidelines: [
      "Use fabbro_prime before a fabbro workflow when you need the current CLI-oriented primer.",
    ],
    parameters: Type.Object({}),
    async execute(_toolCallId, _params, _signal, _onUpdate, ctx) {
      const primer = await getPrimeInfo(pi, ctx.cwd);
      return {
        content: [{ type: "text", text: "Loaded fabbro primer data." }],
        details: { primer },
      };
    },
  });

  pi.registerTool({
    name: "fabbro_create_review",
    label: "Create Fabbro Review",
    description: "Create a non-interactive fabbro review session from generated text",
    promptSnippet: "Create a fabbro review session from generated text and return the session ID.",
    promptGuidelines: [
      "Use fabbro_create_review when you want a human to review generated text in fabbro.",
      "Tell the human to run the returned `fabbro session resume <id>` command outside pi.",
    ],
    parameters: Type.Object({
      content: Type.String({ description: "The text to send into `fabbro review --stdin --no-interactive`" }),
    }),
    async execute(_toolCallId, params, signal, _onUpdate, ctx) {
      try {
        const result = await createReviewSession(pi, params.content, ctx.cwd, signal);
        return {
          content: [
            {
              type: "text",
              text: `Created fabbro review session ${result.sessionId}.\nNext step for the human: ${result.resumeCommand}`,
            },
          ],
          details: result,
        };
      } catch (error) {
        if (isCancelledError(error)) {
          return {
            content: [{ type: "text", text: "Review creation cancelled." }],
            details: { cancelled: true },
          };
        }
        throw handleCommandError(error);
      }
    },
  });

  pi.registerCommand("fabbro-status", {
    description: "Check whether the fabbro CLI is available for the pi integration",
    handler: async (_args, ctx) => {
      const status = await detectFabbro(pi, ctx.cwd);
      const level: Level = status.available ? "success" : "error";

      if (ctx.hasUI) {
        ctx.ui.setStatus(STATUS_KEY, status.message);
      }
      printOrNotify(ctx, status.message, level);
    },
  });

  pi.registerCommand("fabbro-help", {
    description: "Show the current scope of the fabbro pi extension scaffold",
    handler: async (_args, ctx) => {
      const status = await detectFabbro(pi, ctx.cwd);
      const message = [
        "pi-fabbro scaffold is loaded.",
        "Current scope:",
        "- verify a usable fabbro command is available",
        "- expose `fabbro prime --json` to pi as a tool",
        "- create non-interactive review sessions from generated text",
        "- LLM tools: `fabbro_prime`, `fabbro_create_review`",
        `- fabbro status: ${status.message}`,
        "Next phases will add feedback retrieval and session listing commands.",
      ].join("\n");

      if (ctx.hasUI) {
        ctx.ui.setStatus(STATUS_KEY, status.available ? status.message : "fabbro missing");
      }
      printOrNotify(ctx, message, status.available ? "info" : "warning");
    },
  });

  pi.registerCommand("fabbro-prime", {
    description: "Print the current fabbro primer JSON",
    handler: async (_args, ctx) => {
      const primer = await getPrimeInfo(pi, ctx.cwd);
      if (ctx.hasUI) {
        ctx.ui.setStatus(STATUS_KEY, "fabbro primer loaded");
      }
      printOrNotify(ctx, JSON.stringify(primer, null, 2), "info");
    },
  });

  pi.registerCommand("fabbro-review", {
    description: "Create a non-interactive fabbro review session from command text",
    handler: async (args, ctx) => {
      try {
        const result = await createReviewSession(pi, args, ctx.cwd);
        const message = [
          `Created fabbro review session ${result.sessionId}.`,
          `Run this outside pi to review it: ${result.resumeCommand}`,
          `Resolver used: ${result.command}`,
        ].join("\n");

        if (ctx.hasUI) {
          ctx.ui.setStatus(STATUS_KEY, `review ${result.sessionId}`);
        }
        printOrNotify(ctx, message, "success");
      } catch (error) {
        if (isCancelledError(error)) {
          printOrNotify(ctx, "Review creation cancelled.", "warning");
          return;
        }
        throw handleCommandError(error);
      }
    },
  });
}
