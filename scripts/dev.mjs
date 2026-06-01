import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";

const rootDir = fileURLToPath(new URL("..", import.meta.url));
const commands = [
  ["api", "bun", ["run", "--cwd", "node-api", "dev"]],
  ["web", "bun", ["run", "--cwd", "web", "dev"]]
];

let stopping = false;
const children = commands.map(([name, command, args]) => {
  const child = spawn(command, args, {
    cwd: rootDir,
    stdio: "inherit"
  });

  child.on("exit", (code, signal) => {
    if (stopping) {
      return;
    }
    stopping = true;
    console.log(`[${name}] stopped (${signal ?? code ?? 0})`);
    stopChildren();
    process.exitCode = code ?? 1;
  });

  return child;
});

process.on("SIGINT", () => {
  stopping = true;
  stopChildren();
});

process.on("SIGTERM", () => {
  stopping = true;
  stopChildren();
});

function stopChildren() {
  for (const child of children) {
    if (!child.killed) {
      child.kill("SIGTERM");
    }
  }
}
