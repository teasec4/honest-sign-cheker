import { readFile } from "node:fs/promises";
import { CodeIndex } from "../codes.js";
import { normalizeCode } from "../normalize.js";
import type { Diagnostics } from "../types.js";

export async function readCsv(filePath: string): Promise<CodeIndex> {
  const data = await readFile(filePath, "utf8");
  const text = data.replaceAll("\r\n", "\n").replaceAll("\r", "\n");
  const lines = text.split("\n");
  const index = new CodeIndex();
  const diagnostics: Diagnostics = {
    file: filePath,
    rows: lines.length,
    nonEmptyCells: 0,
    countedCells: 0,
    ignoredNonEmptyCells: 0,
    ignoredSamples: []
  };

  lines.forEach((line, lineIndex) => {
    const code = normalizeCode(line);
    if (code === "") {
      return;
    }
    diagnostics.nonEmptyCells += 1;
    diagnostics.countedCells += 1;
    index.add(code, {
      file: filePath,
      line: lineIndex + 1
    });
  });

  index.setDiagnostics(diagnostics);
  return index;
}
