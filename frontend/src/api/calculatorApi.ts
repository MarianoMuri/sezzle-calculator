import type { Operation } from '../domain/operations';

export interface CalculationResponse {
  operation: Operation;
  a: number;
  b?: number;
  result: number;
}

interface ErrorResponse {
  error: { code: string; message: string };
}

/** Error raised for any failed calculation, carrying the backend error code. */
export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

const FRIENDLY_MESSAGES: Record<string, string> = {
  division_by_zero: 'Cannot divide by zero',
  negative_sqrt: 'Cannot take the square root of a negative number',
  undefined_result: 'The result is undefined for these inputs',
  result_out_of_range: 'The result is too large to represent',
  network_error: 'Cannot reach the calculator service. Is the backend running?',
};

export function friendlyMessage(err: unknown): string {
  if (err instanceof ApiError) return FRIENDLY_MESSAGES[err.code] ?? err.message;
  return 'Something went wrong';
}

// Empty base URL = same origin (Vite dev proxy or nginx in Docker).
const BASE_URL: string = import.meta.env.VITE_API_URL ?? '';

export async function calculate(
  operation: Operation,
  a: number,
  b?: number,
  fetchImpl: typeof fetch = fetch,
): Promise<CalculationResponse> {
  const body = b === undefined ? { a } : { a, b };

  let res: Response;
  try {
    res = await fetchImpl(`${BASE_URL}/api/v1/${operation}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  } catch {
    throw new ApiError('network_error', 'Network request failed', 0);
  }

  const data: unknown = await res.json().catch(() => null);

  if (!res.ok) {
    const err = (data as ErrorResponse | null)?.error;
    throw new ApiError(err?.code ?? 'http_error', err?.message ?? `Request failed (${res.status})`, res.status);
  }
  return data as CalculationResponse;
}
