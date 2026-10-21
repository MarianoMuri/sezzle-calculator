export type ParseResult = { ok: true; value: number } | { ok: false; error: string };

// Optional sign, digits with optional decimals (or leading-dot decimals),
// optional exponent. Rejects things Number() would silently accept,
// such as "0x10", "Infinity" or whitespace-only strings.
const NUMBER_PATTERN = /^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$/;

/**
 * Client-side validation of a numeric operand. The backend validates again;
 * this layer exists to give instant feedback without a network round-trip.
 */
export function parseOperand(raw: string): ParseResult {
  const value = raw.trim().replace(',', '.'); // accept "3,5" as well as "3.5"
  if (value === '') return { ok: false, error: 'Enter a number' };
  if (!NUMBER_PATTERN.test(value)) return { ok: false, error: 'Not a valid number' };
  const n = Number(value);
  if (!Number.isFinite(n)) return { ok: false, error: 'Number is too large' };
  return { ok: true, value: n };
}

/** Formats a result for display, trimming floating-point noise (0.1 + 0.2). */
export function formatResult(n: number): string {
  if (Number.isInteger(n) && Math.abs(n) < 1e15) return n.toString();
  return Number.parseFloat(n.toPrecision(12)).toString();
}
