export function styleVars(vars: Record<string, string | number | undefined>): string {
  return Object.entries(vars)
    .filter(([, value]) => value !== undefined)
    .map(([key, value]) => `--${key}:${value};`)
    .join(" ");
}
