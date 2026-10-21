export type Operation =
  | 'add'
  | 'subtract'
  | 'multiply'
  | 'divide'
  | 'power'
  | 'sqrt'
  | 'percentage';

export interface OperationMeta {
  id: Operation;
  symbol: string;
  label: string;
  arity: 1 | 2;
  /** Human-readable expression used in the result/history panel. */
  format: (a: string, b?: string) => string;
}

export const OPERATIONS: readonly OperationMeta[] = [
  { id: 'add', symbol: '+', label: 'Add', arity: 2, format: (a, b) => `${a} + ${b}` },
  { id: 'subtract', symbol: '−', label: 'Subtract', arity: 2, format: (a, b) => `${a} − ${b}` },
  { id: 'multiply', symbol: '×', label: 'Multiply', arity: 2, format: (a, b) => `${a} × ${b}` },
  { id: 'divide', symbol: '÷', label: 'Divide', arity: 2, format: (a, b) => `${a} ÷ ${b}` },
  { id: 'power', symbol: 'xʸ', label: 'Power', arity: 2, format: (a, b) => `${a} ^ ${b}` },
  { id: 'sqrt', symbol: '√', label: 'Square root', arity: 1, format: (a) => `√${a}` },
  { id: 'percentage', symbol: '%', label: 'Percentage (a% of b)', arity: 2, format: (a, b) => `${a}% of ${b}` },
];

export function getOperation(id: Operation): OperationMeta {
  const op = OPERATIONS.find((o) => o.id === id);
  if (!op) throw new Error(`Unknown operation: ${id}`);
  return op;
}
