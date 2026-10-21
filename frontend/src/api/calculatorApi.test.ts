import { describe, expect, it, vi } from 'vitest';
import { ApiError, calculate, friendlyMessage } from './calculatorApi';

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('calculate', () => {
  it('POSTs both operands for binary operations', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { operation: 'add', a: 2, b: 3, result: 5 }));

    const res = await calculate('add', 2, 3, fetchMock);

    expect(res.result).toBe(5);
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/add', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ a: 2, b: 3 }),
    });
  });

  it('omits b for unary operations', async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { operation: 'sqrt', a: 9, result: 3 }));
    await calculate('sqrt', 9, undefined, fetchMock);
    expect(fetchMock.mock.calls[0][1].body).toBe(JSON.stringify({ a: 9 }));
  });

  it('throws ApiError with the backend error code', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse(422, { error: { code: 'division_by_zero', message: 'division by zero' } }));

    const err = await calculate('divide', 1, 0, fetchMock).catch((e: unknown) => e);

    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ code: 'division_by_zero', status: 422 });
  });

  it('handles non-JSON error bodies', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response('Bad Gateway', { status: 502 }));
    await expect(calculate('add', 1, 2, fetchMock)).rejects.toMatchObject({
      code: 'http_error',
      message: 'Request failed (502)',
    });
  });

  it('maps network failures to network_error', async () => {
    const fetchMock = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
    await expect(calculate('add', 1, 2, fetchMock)).rejects.toMatchObject({ code: 'network_error', status: 0 });
  });
});

describe('friendlyMessage', () => {
  it('uses a friendly message for known codes', () => {
    expect(friendlyMessage(new ApiError('division_by_zero', 'x', 422))).toBe('Cannot divide by zero');
  });
  it('falls back to the server message for unknown codes', () => {
    expect(friendlyMessage(new ApiError('weird', 'Server says no', 400))).toBe('Server says no');
  });
  it('handles unexpected errors', () => {
    expect(friendlyMessage(new Error('boom'))).toBe('Something went wrong');
  });
});
