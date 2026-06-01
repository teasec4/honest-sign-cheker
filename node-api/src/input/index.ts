import { access } from "node:fs/promises";
import path from "node:path";
import { CodeIndex } from "../codes.js";
import { readCsv } from "./read-csv.js";
import { readExcel } from "./read-excel.js";

export const supportedExtensions = [".csv", ".txt", ".xlsx", ".xlsm"] as const;

export async function readCodes(filePath: string): Promise<CodeIndex> {
  const extension = path.extname(filePath).toLowerCase();
  switch (extension) {
    case ".csv":
    case ".txt":
      return readCsv(filePath);
    case ".xlsx":
    case ".xlsm":
      return readExcel(filePath);
    default:
      throw new Error(`неподдерживаемый формат файла "${filePath}"`);
  }
}

export async function findNamedFile(dir: string, baseName: string): Promise<string> {
  const matches: string[] = [];
  for (const extension of supportedExtensions) {
    const filePath = path.join(dir, `${baseName}${extension}`);
    if (await exists(filePath)) {
      matches.push(filePath);
    }
  }

  if (matches.length === 0) {
    throw new Error(`в ${dir} не найден файл ${baseName}${supportedExtensions.join("|")}`);
  }
  if (matches.length > 1) {
    throw new Error(`в ${dir} найдено несколько файлов ${baseName}.*: ${matches.join(", ")}`);
  }
  return matches[0]!;
}

async function exists(filePath: string): Promise<boolean> {
  try {
    await access(filePath);
    return true;
  } catch {
    return false;
  }
}
