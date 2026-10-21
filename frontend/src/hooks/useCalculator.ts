import { useCallback, useState } from 'react';
import { calculate as apiCalculate, friendlyMessage } from '../api/calculatorApi';
import { getOperation, type Operation } from '../domain/operations';
import { formatResult, parseOperand } from '../domain/validation';

export type CalculateFn = typeof apiCalculate;

export interface HistoryEntry {
  id: number;
  expression: string;
  result: string;
}

export interface FieldErrors {
  a?: string;
  b?: string;
}

const HISTORY_LIMIT = 5;

/**
 * Holds all calculator state and the submit flow. Kept separate from the
 * component so the logic is easy to test and the JSX stays presentational.
 */
export function useCalculator(calculateFn: CalculateFn = apiCalculate) {
  const [a, setA] = useState('');
  const [b, setB] = useState('');
  const [operation, setOperation] = useState<Operation>('add');
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [loading, setLoading] = useState(false);
  const [history, setHistory] = useState<HistoryEntry[]>([]);

  const meta = getOperation(operation);

  const selectOperation = useCallback((op: Operation) => {
    setOperation(op);
    setError(null);
    setFieldErrors({});
  }, []);

  const submit = useCallback(async () => {
    const parsedA = parseOperand(a);
    const parsedB = meta.arity === 2 ? parseOperand(b) : null;

    const errors: FieldErrors = {};
    if (!parsedA.ok) errors.a = parsedA.error;
    if (parsedB && !parsedB.ok) errors.b = parsedB.error;
    setFieldErrors(errors);
    setError(null);
    if (!parsedA.ok || (parsedB && !parsedB.ok)) return;

    setLoading(true);
    try {
      const res = await calculateFn(operation, parsedA.value, parsedB?.value);
      const formatted = formatResult(res.result);
      setResult(formatted);
      setHistory((h) =>
        [{ id: Date.now() + Math.random(), expression: meta.format(a.trim(), b.trim()), result: formatted }, ...h].slice(
          0,
          HISTORY_LIMIT,
        ),
      );
    } catch (err) {
      setResult(null);
      setError(friendlyMessage(err));
    } finally {
      setLoading(false);
    }
  }, [a, b, meta, operation, calculateFn]);

  const reset = useCallback(() => {
    setA('');
    setB('');
    setResult(null);
    setError(null);
    setFieldErrors({});
  }, []);

  /** Reuse the last result as the first operand, like a chained calculation. */
  const chainResult = useCallback(() => {
    if (result !== null) {
      setA(result);
      setB('');
    }
  }, [result]);

  return {
    a, b, operation, meta, result, error, fieldErrors, loading, history,
    setA, setB, selectOperation, submit, reset, chainResult,
  };
}
