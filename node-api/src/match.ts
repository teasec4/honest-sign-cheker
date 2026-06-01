import { CodeIndex, compareStrings } from "./codes.js";
import {
  StatusExact,
  StatusFuzzy,
  StatusUnknown,
  type MatchConfig,
  type MatchResult,
  type MatchStatus
} from "./types.js";

interface Candidate {
  code: string;
  sharedGrams: number;
}

export function defaultConfig(): MatchConfig {
  return {
    minPercent: 85,
    gramSize: 3,
    maxCandidates: 500
  };
}

export class Matcher {
  private readonly issued: CodeIndex;
  private readonly config: MatchConfig;
  private readonly grams = new Map<string, string[]>();

  constructor(issued: CodeIndex, config: Partial<MatchConfig> = {}) {
    this.issued = issued;
    this.config = {
      ...defaultConfig(),
      ...config
    };

    for (const code of issued.codes()) {
      for (const gram of uniqueGrams(code, this.config.gramSize)) {
        const current = this.grams.get(gram) ?? [];
        current.push(code);
        this.grams.set(gram, current);
      }
    }
  }

  matchReturned(returned: CodeIndex): MatchResult[] {
    const results = returned.codes().map((returnedCode) => {
      const result = this.matchCode(returnedCode);
      return {
        ...result,
        returnedCode,
        returnedPlaces: returned.locationsFor(returnedCode)
      };
    });

    return results.sort((a, b) => {
      if (a.status !== b.status) {
        return statusOrder(a.status) - statusOrder(b.status);
      }
      if (a.matchPercent !== b.matchPercent) {
        return b.matchPercent - a.matchPercent;
      }
      return compareStrings(a.returnedCode, b.returnedCode);
    });
  }

  matchCode(returnedCode: string): MatchResult {
    if (this.issued.has(returnedCode)) {
      return {
        status: StatusExact,
        returnedCode,
        returnedPlaces: [],
        matchedCode: returnedCode,
        matchedPlaces: this.issued.locationsFor(returnedCode),
        matchPercent: 100,
        sharedGramCount: 0
      };
    }

    const candidates = this.candidates(returnedCode);
    if (candidates.length === 0) {
      return emptyResult(StatusUnknown, returnedCode);
    }

    let best = emptyResult(StatusUnknown, returnedCode);
    for (const candidate of candidates) {
      const percent = similarityPercent(returnedCode, candidate.code);
      if (percent > best.matchPercent) {
        best = {
          status: StatusUnknown,
          returnedCode,
          returnedPlaces: [],
          matchedCode: candidate.code,
          matchedPlaces: this.issued.locationsFor(candidate.code),
          matchPercent: percent,
          sharedGramCount: candidate.sharedGrams
        };
      }
    }

    if (best.matchPercent >= this.config.minPercent) {
      best.status = StatusFuzzy;
    }
    return best;
  }

  private candidates(returnedCode: string): Candidate[] {
    const counts = new Map<string, number>();
    for (const gram of uniqueGrams(returnedCode, this.config.gramSize)) {
      for (const code of this.grams.get(gram) ?? []) {
        counts.set(code, (counts.get(code) ?? 0) + 1);
      }
    }

    const candidates = [...counts.entries()]
      .map(([code, sharedGrams]) => ({ code, sharedGrams }))
      .sort((a, b) => {
        if (a.sharedGrams !== b.sharedGrams) {
          return b.sharedGrams - a.sharedGrams;
        }
        return compareStrings(a.code, b.code);
      });

    return candidates.slice(0, this.config.maxCandidates);
  }
}

export function similarityPercent(a: string, b: string): number {
  const ar = [...a];
  const br = [...b];
  const maxLen = Math.max(ar.length, br.length);
  if (maxLen === 0) {
    return 100;
  }
  const distance = levenshtein(ar, br);
  return (1 - distance / maxLen) * 100;
}

function levenshtein(a: string[], b: string[]): number {
  if (a.length === 0) {
    return b.length;
  }
  if (b.length === 0) {
    return a.length;
  }

  let previous = Array.from({ length: b.length + 1 }, (_, index) => index);
  let current = new Array<number>(b.length + 1);

  for (let i = 1; i <= a.length; i += 1) {
    current[0] = i;
    for (let j = 1; j <= b.length; j += 1) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      current[j] = Math.min(
        previous[j]! + 1,
        current[j - 1]! + 1,
        previous[j - 1]! + cost
      );
    }
    [previous, current] = [current, previous];
  }

  return previous[b.length]!;
}

function uniqueGrams(value: string, size: number): string[] {
  const runes = [...value];
  if (runes.length < size || size <= 0) {
    return [];
  }

  const seen = new Set<string>();
  const grams: string[] = [];
  for (let index = 0; index <= runes.length - size; index += 1) {
    const gram = runes.slice(index, index + size).join("");
    if (seen.has(gram)) {
      continue;
    }
    seen.add(gram);
    grams.push(gram);
  }
  return grams;
}

function emptyResult(status: MatchStatus, returnedCode: string): MatchResult {
  return {
    status,
    returnedCode,
    returnedPlaces: [],
    matchedCode: "",
    matchedPlaces: [],
    matchPercent: 0,
    sharedGramCount: 0
  };
}

function statusOrder(status: MatchStatus): number {
  switch (status) {
    case StatusFuzzy:
      return 0;
    case StatusUnknown:
      return 1;
    case StatusExact:
      return 2;
  }
}
