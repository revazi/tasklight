#!/usr/bin/env node

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const platform = process.platform;
const arch = process.arch;
const key = `${platform}-${arch}`;

const targets = {
  "darwin-arm64": "darwin-arm64",
  "darwin-x64": "darwin-amd64",
  "linux-arm64": "linux-arm64",
  "linux-x64": "linux-amd64",
};

const target = targets[key];

if (!target) {
  console.error(`Tasklight: unsupported platform ${platform}/${arch}.`);
  console.error("Supported platforms: macOS/Linux on arm64/x64.");
  process.exit(1);
}

const binary = path.join(__dirname, "..", "vendor", target, "tasklight");

if (!fs.existsSync(binary)) {
  console.error(`Tasklight binary not found for ${platform}/${arch}: ${binary}`);
  console.error("If you are developing locally, run: npm run build:vendor");
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
