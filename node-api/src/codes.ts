import path from "node:path";
import type { Diagnostics, Location } from "./types.js";

export class CodeIndex {
  private readonly locations = new Map<string, Location[]>();
  private total = 0;
  private diagnosticsValue?: Diagnostics;

  add(code: string, location: Location): void {
    const current = this.locations.get(code) ?? [];
    current.push(location);
    this.locations.set(code, current);
    this.total += 1;
  }

  setDiagnostics(diagnostics: Diagnostics): void {
    this.diagnosticsValue = diagnostics;
  }

  diagnostics(): Diagnostics | undefined {
    return this.diagnosticsValue;
  }

  totalCount(): number {
    return this.total;
  }

  uniqueCount(): number {
    return this.locations.size;
  }

  count(code: string): number {
    return this.locations.get(code)?.length ?? 0;
  }

  has(code: string): boolean {
    return this.locations.has(code);
  }

  locationsFor(code: string): Location[] {
    return this.locations.get(code) ?? [];
  }

  codes(): string[] {
    return [...this.locations.keys()].sort(compareStrings);
  }

  duplicateCodes(): string[] {
    return [...this.locations.entries()]
      .filter(([, locations]) => locations.length > 1)
      .map(([code]) => code)
      .sort(compareStrings);
  }
}

export function stringifyLocation(location: Location): string {
  const fileName = path.basename(location.file);
  if (location.sheet && location.cell) {
    return `${fileName}:${location.sheet}!${location.cell}`;
  }
  if (location.line && location.line > 0) {
    return `${fileName}:строка ${location.line}`;
  }
  return fileName;
}

export function stringifyLocations(locations: Location[]): string[] {
  return locations.map(stringifyLocation);
}

export function compareStrings(a: string, b: string): number {
  if (a < b) {
    return -1;
  }
  if (a > b) {
    return 1;
  }
  return 0;
}
