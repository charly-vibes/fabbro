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

type FabbroApplyResult = {
  sessionId: string;
  sourceFile: string;
  createdAt: string;
  annotations: Array<{
    type: string;
    text: string;
    startLine: number;
    endLine: number;
  }>;
  warnings: string[];
  stderr?: string;
  command: string;
  annotationCount: number;
};

type FabbroSessionListEntry = {
  id: string;
  createdAt: string;
  sourceFile?: string;
  annotations: number;
  resumeCommand: string;
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

function parseJSON<T>(commandLabel: string, raw: string): T {
  try {
    return JSON.parse(raw) as T;
  } catch (error) {
    const details = error instanceof Error ? error.message : String(error);
    throw new Error(`${commandLabel} returned invalid JSON: ${details}`);
  }
}

function extractWarnings(stderr: string): string[] {
  return stderr
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => line.replace(/^warning:\s*/i, ""));
}

async function getPrimeInfo(pi: ExtensionAPI, cwd: string) {
  const status = await ensureFabbro(pi, cwd);
  const result = await execWithPi(pi, status.resolvedCommand!, ["prime", "--json"]);
  if (result.code !== 0) {
    throw new Error((result.stderr || result.stdout || "fabbro prime failed").trim());
  }

  return parseJSON("fabbro prime", result.stdout);
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

async function listSessions(pi: ExtensionAPI, cwd: string): Promise<FabbroSessionListEntry[]> {
  const status = await ensureFabbro(pi, cwd);
  const fabbro = status.resolvedCommand!;
  const result = await execWithPi(pi, fabbro, ["session", "list", "--json"]);

  if (result.code !== 0) {
    throw new Error((result.stderr || result.stdout || "fabbro session list failed").trim());
  }

  const sessions = parseJSON<Array<Omit<FabbroSessionListEntry, "resumeCommand">>>(
    "fabbro session list --json",
    result.stdout,
  );

  return (Array.isArray(sessions) ? sessions : []).map((session) => ({
    ...session,
    resumeCommand: `fabbro session resume ${session.id}`,
  }));
}

async function applyFeedback(pi: ExtensionAPI, sessionId: string, cwd: string): Promise<FabbroApplyResult> {
  const trimmedSessionId = sessionId.trim();
  if (!trimmedSessionId) {
    throw new Error("Session ID is required. Provide it after /fabbro-apply or via the fabbro_apply_feedback tool.");
  }

  const status = await ensureFabbro(pi, cwd);
  const fabbro = status.resolvedCommand!;
  const result = await execWithPi(pi, fabbro, ["apply", trimmedSessionId, "--json"]);

  if (result.code !== 0) {
    const output = [result.stderr.trim(), result.stdout.trim()].filter(Boolean).join("\n");
    throw new Error(output || `${fabbro.display} apply ${trimmedSessionId} --json failed`);
  }

  const feedback = parseJSON<Omit<FabbroApplyResult, "warnings" | "stderr" | "command" | "annotationCount">>(
    `fabbro apply ${trimmedSessionId} --json`,
    result.stdout,
  );
  const warnings = extractWarnings(result.stderr || "");
  const annotations = Array.isArray(feedback.annotations) ? feedback.annotations : [];

  return {
    ...feedback,
    annotations,
    warnings,
    stderr: result.stderr.trim() || undefined,
    command: fabbro.display,
    annotationCount: annotations.length,
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

  pi.registerTool({
    name: "fabbro_apply_feedback",
    label: "Apply Fabbro Feedback",
    description: "Load structured feedback from `fabbro apply <session-id> --json`",
    promptSnippet: "Load structured fabbro annotations for an existing review session.",
    promptGuidelines: [
      "Use fabbro_apply_feedback after a human has reviewed a session in fabbro.",
      "If warnings are returned, surface them because they may indicate source drift or other review caveats.",
    ],
    parameters: Type.Object({
      sessionId: Type.String({ description: "The fabbro session ID to apply with `fabbro apply <id> --json`" }),
    }),
    async execute(_toolCallId, params, _signal, _onUpdate, ctx) {
      const feedback = await applyFeedback(pi, params.sessionId, ctx.cwd);
      const warningSummary = feedback.warnings.length > 0 ? `\nWarnings:\n- ${feedback.warnings.join("\n- ")}` : "";
      return {
        content: [
          {
            type: "text",
            text: `Loaded fabbro feedback for ${feedback.sessionId} (${feedback.annotationCount} annotations).${warningSummary}`,
          },
        ],
        details: feedback,
      };
    },
  });

  pi.registerTool({
    name: "fabbro_list_sessions",
    label: "List Fabbro Sessions",
    description: "List fabbro sessions via `fabbro session list --json`",
    promptSnippet: "List the available fabbro sessions and show how to resume them outside pi.",
    promptGuidelines: [
      "Use fabbro_list_sessions when you need to discover an existing session to inspect or resume.",
      "Recommend the returned `fabbro session resume <id>` command when the human needs to continue review in the external TUI.",
    ],
    parameters: Type.Object({}),
    async execute(_toolCallId, _params, _signal, _onUpdate, ctx) {
      const sessions = await listSessions(pi, ctx.cwd);
      const summary = sessions.length
        ? sessions.slice(0, 10).map((session) => `- ${session.id} (${session.annotations} annotations) → ${session.resumeCommand}`).join("\n")
        : "No sessions found.";
      return {
        content: [{ type: "text", text: `Loaded ${sessions.length} fabbro session(s).\n${summary}` }],
        details: { sessions },
      };
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
        "- load structured feedback from an existing session",
        "- list available sessions and show resume commands",
        "- LLM tools: `fabbro_prime`, `fabbro_create_review`, `fabbro_apply_feedback`, `fabbro_list_sessions`",
        `- fabbro status: ${status.message}`,
        "Next phases will assess reuse boundaries and workflow validation.",
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

  pi.registerCommand("fabbro-apply", {
    description: "Load structured feedback from an existing fabbro session",
    handler: async (args, ctx) => {
      const feedback = await applyFeedback(pi, args, ctx.cwd);
      const message = [
        `Loaded fabbro feedback for ${feedback.sessionId}.`,
        `Annotations: ${feedback.annotationCount}`,
        `Resolver used: ${feedback.command}`,
        feedback.warnings.length > 0 ? `Warnings:\n- ${feedback.warnings.join("\n- ")}` : null,
        JSON.stringify(
          {
            sessionId: feedback.sessionId,
            sourceFile: feedback.sourceFile,
            createdAt: feedback.createdAt,
            annotations: feedback.annotations,
          },
          null,
          2,
        ),
      ]
        .filter(Boolean)
        .join("\n");

      if (ctx.hasUI) {
        ctx.ui.setStatus(STATUS_KEY, `feedback ${feedback.sessionId}`);
      }
      printOrNotify(ctx, message, feedback.warnings.length > 0 ? "warning" : "success");
    },
  });

  pi.registerCommand("fabbro-sessions", {
    description: "List fabbro sessions and show how to resume them outside pi",
    handler: async (_args, ctx) => {
      const sessions = await listSessions(pi, ctx.cwd);
      const message = sessions.length
        ? [
            `Found ${sessions.length} fabbro session(s).`,
            ...sessions.map(
              (session) =>
                `${session.id} | ${session.createdAt} | ${session.sourceFile || "(stdin)"} | ${session.annotations} annotations | ${session.resumeCommand}`,
            ),
          ].join("\n")
        : "No fabbro sessions found. Create one with /fabbro-review or fabbro review <file>.";

      if (ctx.hasUI) {
        ctx.ui.setStatus(STATUS_KEY, sessions.length > 0 ? `${sessions.length} sessions` : "no sessions");
      }
      printOrNotify(ctx, message, "info");
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
