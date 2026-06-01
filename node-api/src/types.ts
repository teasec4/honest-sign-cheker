export interface Location {
  file: string;
  sheet?: string;
  cell?: string;
  line?: number;
}

export interface IgnoredSample {
  location: string;
  value: string;
  reason: string;
}

export interface Diagnostics {
  file: string;
  sheet?: string;
  column?: string;
  rows: number;
  nonEmptyCells: number;
  countedCells: number;
  ignoredNonEmptyCells: number;
  ignoredSamples: IgnoredSample[];
}

export interface FileSummary {
  total: number;
  unique: number;
  diagnostics?: Diagnostics;
}

export interface PrimarySummary {
  issued: FileSummary;
  returned: FileSummary;
  exactTotal: number;
  exactUnique: number;
  fuzzyTotal: number;
  fuzzyUnique: number;
  unknownTotal: number;
  unknownUnique: number;
  duplicateUnique: number;
}

export interface Problem {
  type: string;
  code: string;
  description: string;
  count?: number;
  matchPercent?: number;
  matchedCode?: string;
  returnedLocations?: string[];
  issuedLocations?: string[];
}

export interface PrimaryReport {
  minPercent: number;
  summary: PrimarySummary;
  duplicates: Problem[];
  fuzzy: Problem[];
  unknown: Problem[];
}

export interface DuplicateReport {
  summary: FileSummary;
  duplicates: Problem[];
}

export const StatusExact = "ТОЧНОЕ СОВПАДЕНИЕ";
export const StatusFuzzy = "ПОХОЖИЙ КОД";
export const StatusUnknown = "НЕ НАЙДЕН";

export type MatchStatus =
  | typeof StatusExact
  | typeof StatusFuzzy
  | typeof StatusUnknown;

export interface MatchConfig {
  minPercent: number;
  gramSize: number;
  maxCandidates: number;
}

export interface MatchResult {
  status: MatchStatus;
  returnedCode: string;
  returnedPlaces: Location[];
  matchedCode: string;
  matchedPlaces: Location[];
  matchPercent: number;
  sharedGramCount: number;
}
