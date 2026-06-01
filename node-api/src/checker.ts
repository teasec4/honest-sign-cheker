import { CodeIndex, stringifyLocations } from "./codes.js";
import { readCodes } from "./input/index.js";
import { defaultConfig, Matcher } from "./match.js";
import {
  StatusExact,
  StatusFuzzy,
  StatusUnknown,
  type DuplicateReport,
  type FileSummary,
  type MatchResult,
  type PrimaryReport,
  type Problem
} from "./types.js";

export async function runPrimary(
  issuedPath: string,
  returnedPath: string,
  minPercent: number
): Promise<PrimaryReport> {
  let issued: CodeIndex;
  let returned: CodeIndex;
  try {
    issued = await readCodes(issuedPath);
  } catch (error) {
    throw new Error(`выданные коды: ${messageFromError(error)}`);
  }
  try {
    returned = await readCodes(returnedPath);
  } catch (error) {
    throw new Error(`возврат поставщика: ${messageFromError(error)}`);
  }

  const config = defaultConfig();
  if (minPercent > 0) {
    config.minPercent = minPercent;
  }

  const results = new Matcher(issued, config).matchReturned(returned);
  const report: PrimaryReport = {
    minPercent: config.minPercent,
    summary: {
      issued: fileSummary(issued),
      returned: fileSummary(returned),
      exactTotal: 0,
      exactUnique: 0,
      fuzzyTotal: 0,
      fuzzyUnique: 0,
      unknownTotal: 0,
      unknownUnique: 0,
      duplicateUnique: 0
    },
    duplicates: [],
    fuzzy: [],
    unknown: []
  };

  for (const code of returned.duplicateCodes()) {
    const count = returned.count(code);
    report.duplicates.push({
      type: "ДУБЛИКАТ В ВОЗВРАТЕ",
      code,
      description: `код встречается в файле поставщика ${count} раза`,
      count,
      returnedLocations: stringifyLocations(returned.locationsFor(code))
    });
  }
  report.summary.duplicateUnique = report.duplicates.length;

  for (const result of results) {
    const count = result.returnedPlaces.length;
    switch (result.status) {
      case StatusExact:
        report.summary.exactTotal += count;
        report.summary.exactUnique += 1;
        break;
      case StatusFuzzy:
        report.summary.fuzzyTotal += count;
        report.summary.fuzzyUnique += 1;
        report.fuzzy.push(problemFromMatchResult(result));
        break;
      case StatusUnknown:
        report.summary.unknownTotal += count;
        report.summary.unknownUnique += 1;
        report.unknown.push(problemFromMatchResult(result));
        break;
    }
  }

  return report;
}

export async function runDuplicates(filePath: string): Promise<DuplicateReport> {
  const index = await readCodes(filePath);
  const report: DuplicateReport = {
    summary: fileSummary(index),
    duplicates: []
  };

  for (const code of index.duplicateCodes()) {
    const count = index.count(code);
    report.duplicates.push({
      type: "ДУБЛИКАТ",
      code,
      description: `код встречается ${count} раза`,
      count,
      returnedLocations: stringifyLocations(index.locationsFor(code))
    });
  }

  return report;
}

function fileSummary(index: CodeIndex): FileSummary {
  return {
    total: index.totalCount(),
    unique: index.uniqueCount(),
    diagnostics: index.diagnostics()
  };
}

function problemFromMatchResult(result: MatchResult): Problem {
  const problem: Problem = {
    type: result.status,
    code: result.returnedCode,
    description:
      result.status === StatusFuzzy
        ? "код похож на выданный, нужен ручной контроль"
        : "не найден достаточно похожий код среди выданных",
    returnedLocations: stringifyLocations(result.returnedPlaces)
  };

  if (result.matchPercent !== 0) {
    problem.matchPercent = result.matchPercent;
  }
  if (result.matchedCode !== "") {
    problem.matchedCode = result.matchedCode;
  }
  const issuedLocations = stringifyLocations(result.matchedPlaces);
  if (issuedLocations.length > 0) {
    problem.issuedLocations = issuedLocations;
  }

  return problem;
}

function messageFromError(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
