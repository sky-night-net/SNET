/**
 * Generates a UUID v4 compatible string.
 * Uses crypto.randomUUID() if available (secure context / localhost),
 * otherwise falls back to a manual implementation for non-secure contexts (HTTP over IP).
 */
export function generateUUID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  
  // RFC4122-compliant fallback for non-secure contexts
  return (([1e7] as any) + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c: number) =>
    (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16)
  );
}
