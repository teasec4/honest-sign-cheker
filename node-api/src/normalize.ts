export function normalizeCode(code: string): string {
  return code
    .trim()
    .replaceAll("\x10", "")
    .replaceAll("\x1f", "")
    .replaceAll("\x1d", "")
    .replaceAll("_x001D_", "")
    .replaceAll("?", "'")
    .replaceAll("\u0092", "'");
}
