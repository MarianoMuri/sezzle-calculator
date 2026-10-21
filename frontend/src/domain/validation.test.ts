import { describe, expect, it } from 'vitest';
import { formatResult, parseOperand } from './validation';
import { getOperation, OPERATIONS } from './operations';

describe('parseOperand', () => {
  it.each([
    ['42', 42],
    ['-3.5', -3.5],
    ['+7', 7],
    ['.5', 0.5],
    ['5.', 5],
    ['1e3', 1000],
    ['  12  ', 12],
    ['3,25', 3.25],
    ['0', 0],
  ])('accepts %j', (raw, expected) => {
    expect(parseOperand(raw)).toEqual({ ok: true, value: expected });
  });

  it.each(['abc', '1.2.3', '0x10', 'Infinity', 'NaN', '--1', '1e', '5 5'])('rejects %j', (raw) => {
    expect(parseOperand(raw)).toEqual({ ok: false, error: 'Not a valid number' });
  });

  it('rejects empty input', () => {
    expect(parseOperand('   ')).toEqual({ ok: false, error: 'Enter a number' });
  });

  it('rejects numbers that overflow to Infinity', () => {
    expect(parseOperand('1e400')).toEqual({ ok: false, error: 'Number is too large' });
  });
});

describe('formatResult', () => {
  it('keeps integers intact', () => expect(formatResult(1024)).toBe('1024'));
  it('trims floating point noise', () => expect(formatResult(0.1 + 0.2)).toBe('0.3'));
  it('keeps meaningful decimals', () => expect(formatResult(2.5)).toBe('2.5'));
  it('handles very large numbers', () => expect(formatResult(1e21)).toBe('1e+21'));
});

describe('operations', () => {
  it('formats expressions', () => {
    expect(getOperation('add').format('1', '2')).toBe('1 + 2');
    expect(getOperation('sqrt').format('9')).toBe('√9');
    expect(getOperation('percentage').format('10', '50')).toBe('10% of 50');
  });

  it.each(OPERATIONS.map((o) => [o.id, o]))('%s formats an expression containing its operands', (_, op) => {
    const text = op.format('7', '8');
    expect(text).toContain('7');
    if (op.arity === 2) expect(text).toContain('8');
  });

  it('marks only sqrt as unary', () => {
    expect(OPERATIONS.filter((o) => o.arity === 1).map((o) => o.id)).toEqual(['sqrt']);
  });

  it('throws on unknown operation', () => {
    // @ts-expect-error - testing a runtime guard
    expect(() => getOperation('modulo')).toThrow('Unknown operation');
  });
});
