import XLSX from "xlsx";
import { CodeIndex, stringifyLocation } from "../codes.js";
import { normalizeCode } from "../normalize.js";
import type { Diagnostics, Location } from "../types.js";

export function readExcel(filePath: string): CodeIndex {
  const workbook = XLSX.readFile(filePath, {
    cellDates: false
  });
  const sheetName = workbook.SheetNames[0];
  if (!sheetName) {
    throw new Error("в файле нет листов");
  }
  const sheet = workbook.Sheets[sheetName];
  if (!sheet) {
    throw new Error(`лист "${sheetName}" не найден`);
  }

  const rows = sheetToRows(sheet);
  const index = new CodeIndex();
  const diagnostics: Diagnostics = {
    file: filePath,
    sheet: sheetName,
    rows: rows.length,
    nonEmptyCells: 0,
    countedCells: 0,
    ignoredNonEmptyCells: 0,
    ignoredSamples: []
  };

  const headerColumn = findHeaderColumn(rows, "DM CODE");
  if (headerColumn !== -1) {
    readColumn(filePath, sheetName, rows, headerColumn, index, diagnostics);
    index.setDiagnostics(diagnostics);
    return index;
  }

  rows.forEach((row, rowIndex) => {
    row.forEach((cell, columnIndex) => {
      const code = normalizeCode(cell);
      if (code === "") {
        return;
      }
      diagnostics.nonEmptyCells += 1;

      const location: Location = {
        file: filePath,
        sheet: sheetName,
        cell: XLSX.utils.encode_cell({ r: rowIndex, c: columnIndex })
      };

      if (isCodeLike(code)) {
        diagnostics.countedCells += 1;
        index.add(code, location);
      } else {
        diagnostics.ignoredNonEmptyCells += 1;
        addIgnoredSample(diagnostics, location, code, "ячейка не похожа на код маркировки");
      }
    });
  });

  index.setDiagnostics(diagnostics);
  return index;
}

function sheetToRows(sheet: XLSX.WorkSheet): string[][] {
  const rawRows = XLSX.utils.sheet_to_json(sheet, {
    header: 1,
    defval: "",
    raw: false,
    blankrows: false
  }) as unknown[][];

  return rawRows.map((row) => {
    if (!Array.isArray(row)) {
      return [];
    }
    return row.map((cell) => (cell == null ? "" : String(cell)));
  });
}

function readColumn(
  filePath: string,
  sheetName: string,
  rows: string[][],
  columnIndex: number,
  index: CodeIndex,
  diagnostics: Diagnostics
): void {
  diagnostics.column = XLSX.utils.encode_col(columnIndex);

  for (let rowIndex = 1; rowIndex < rows.length; rowIndex += 1) {
    const row = rows[rowIndex];
    if (!row || columnIndex >= row.length) {
      continue;
    }

    const code = normalizeCode(row[columnIndex] ?? "");
    if (code === "") {
      continue;
    }

    diagnostics.nonEmptyCells += 1;
    diagnostics.countedCells += 1;
    index.add(code, {
      file: filePath,
      sheet: sheetName,
      cell: XLSX.utils.encode_cell({ r: rowIndex, c: columnIndex })
    });
  }
}

function findHeaderColumn(rows: string[][], header: string): number {
  const firstRow = rows[0];
  if (!firstRow) {
    return -1;
  }

  const normalizedHeader = normalizeHeader(header);
  return firstRow.findIndex((cell) => normalizeHeader(cell) === normalizedHeader);
}

function normalizeHeader(value: string): string {
  return value.trim().toUpperCase().split(/\s+/).filter(Boolean).join(" ");
}

function isCodeLike(code: string): boolean {
  if (code.startsWith("01")) {
    return true;
  }
  return [...code].length >= 20 && code.includes("93");
}

function addIgnoredSample(
  diagnostics: Diagnostics,
  location: Location,
  value: string,
  reason: string
): void {
  if (diagnostics.ignoredSamples.length >= 30) {
    return;
  }
  diagnostics.ignoredSamples.push({
    location: stringifyLocation(location),
    value: truncate(value, 80),
    reason
  });
}

function truncate(value: string, limit: number): string {
  const runes = [...value];
  if (runes.length <= limit) {
    return value;
  }
  return `${runes.slice(0, limit).join("")}...`;
}
