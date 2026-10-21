import type { FormEvent } from 'react';
import { OPERATIONS } from '../domain/operations';
import { useCalculator, type CalculateFn } from '../hooks/useCalculator';

interface Props {
  /** Injectable for tests; defaults to the real REST client. */
  calculateFn?: CalculateFn;
}

export function Calculator({ calculateFn }: Props) {
  const calc = useCalculator(calculateFn);
  const { meta } = calc;

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    void calc.submit();
  };

  return (
    <section className="calculator" aria-labelledby="calc-title">
      <h1 id="calc-title">Calculator</h1>

      <form onSubmit={onSubmit} noValidate>
        <div className="field">
          <label htmlFor="operand-a">{meta.arity === 1 ? 'Number' : 'First number (a)'}</label>
          <input
            id="operand-a"
            inputMode="decimal"
            autoComplete="off"
            value={calc.a}
            onChange={(e) => calc.setA(e.target.value)}
            aria-invalid={Boolean(calc.fieldErrors.a)}
            aria-describedby={calc.fieldErrors.a ? 'operand-a-error' : undefined}
            placeholder="0"
          />
          {calc.fieldErrors.a && (
            <span id="operand-a-error" className="field-error">
              {calc.fieldErrors.a}
            </span>
          )}
        </div>

        <fieldset className="operations">
          <legend>Operation</legend>
          {OPERATIONS.map((op) => (
            <button
              key={op.id}
              type="button"
              className="op-button"
              aria-pressed={calc.operation === op.id}
              aria-label={op.label}
              title={op.label}
              onClick={() => calc.selectOperation(op.id)}
            >
              {op.symbol}
            </button>
          ))}
        </fieldset>

        {meta.arity === 2 && (
          <div className="field">
            <label htmlFor="operand-b">Second number (b)</label>
            <input
              id="operand-b"
              inputMode="decimal"
              autoComplete="off"
              value={calc.b}
              onChange={(e) => calc.setB(e.target.value)}
              aria-invalid={Boolean(calc.fieldErrors.b)}
              aria-describedby={calc.fieldErrors.b ? 'operand-b-error' : undefined}
              placeholder="0"
            />
            {calc.fieldErrors.b && (
              <span id="operand-b-error" className="field-error">
                {calc.fieldErrors.b}
              </span>
            )}
          </div>
        )}

        <div className="actions">
          <button type="button" className="secondary" onClick={calc.reset}>
            Clear
          </button>
          <button type="submit" className="primary" disabled={calc.loading}>
            {calc.loading ? 'Calculating…' : '='}
          </button>
        </div>
      </form>

      <output className="display" aria-live="polite" htmlFor="operand-a operand-b">
        {calc.error ? (
          <span role="alert" className="display-error">
            {calc.error}
          </span>
        ) : calc.result !== null ? (
          <>
            <span className="display-label">Result</span>
            <span className="display-value" data-testid="result">
              {calc.result}
            </span>
            <button type="button" className="link" onClick={calc.chainResult}>
              Use as next input
            </button>
          </>
        ) : (
          <span className="display-placeholder">Enter numbers and press =</span>
        )}
      </output>

      {calc.history.length > 0 && (
        <div className="history">
          <h2>Recent</h2>
          <ol>
            {calc.history.map((h) => (
              <li key={h.id}>
                <span>{h.expression}</span>
                <strong>= {h.result}</strong>
              </li>
            ))}
          </ol>
        </div>
      )}
    </section>
  );
}
