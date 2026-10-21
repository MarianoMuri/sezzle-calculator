import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '../api/calculatorApi';
import type { CalculateFn } from '../hooks/useCalculator';
import { Calculator } from './Calculator';
import App from '../App';

function setup(calculateFn: CalculateFn) {
  const user = userEvent.setup();
  render(<Calculator calculateFn={calculateFn} />);
  return {
    user,
    a: () => screen.getByLabelText(/first number|^number$/i),
    b: () => screen.queryByLabelText(/second number/i),
    equals: () => screen.getByRole('button', { name: '=' }),
    op: (name: string) => screen.getByRole('button', { name }),
  };
}

const ok = (result: number) => vi.fn<CalculateFn>().mockResolvedValue({ operation: 'add', a: 0, result });

describe('Calculator', () => {
  it('calculates and shows the result', async () => {
    const calc = ok(5);
    const ui = setup(calc);

    await ui.user.type(ui.a(), '2');
    await ui.user.type(ui.b()!, '3');
    await ui.user.click(ui.equals());

    expect(calc).toHaveBeenCalledWith('add', 2, 3);
    expect(await screen.findByTestId('result')).toHaveTextContent('5');
    expect(screen.getByText('2 + 3')).toBeInTheDocument(); // history entry
  });

  it('submits with the Enter key', async () => {
    const calc = ok(12);
    const ui = setup(calc);
    await ui.user.click(ui.op('Multiply'));
    await ui.user.type(ui.a(), '3');
    await ui.user.type(ui.b()!, '4{Enter}');
    expect(calc).toHaveBeenCalledWith('multiply', 3, 4);
  });

  it('hides the second operand for square root', async () => {
    const calc = ok(3);
    const ui = setup(calc);

    await ui.user.click(ui.op('Square root'));
    expect(ui.b()).not.toBeInTheDocument();
    expect(ui.op('Square root')).toHaveAttribute('aria-pressed', 'true');

    await ui.user.type(ui.a(), '9');
    await ui.user.click(ui.equals());
    expect(calc).toHaveBeenCalledWith('sqrt', 9, undefined);
  });

  it('validates input before calling the API', async () => {
    const calc = ok(0);
    const ui = setup(calc);

    await ui.user.type(ui.a(), 'abc');
    await ui.user.click(ui.equals());

    expect(calc).not.toHaveBeenCalled();
    expect(screen.getByText('Not a valid number')).toBeInTheDocument();
    expect(screen.getByText('Enter a number')).toBeInTheDocument(); // empty b
    expect(ui.a()).toHaveAttribute('aria-invalid', 'true');
  });

  it('shows a friendly message for backend errors', async () => {
    const calc = vi.fn<CalculateFn>().mockRejectedValue(new ApiError('division_by_zero', 'division by zero', 422));
    const ui = setup(calc);

    await ui.user.click(ui.op('Divide'));
    await ui.user.type(ui.a(), '1');
    await ui.user.type(ui.b()!, '0');
    await ui.user.click(ui.equals());

    expect(await screen.findByRole('alert')).toHaveTextContent('Cannot divide by zero');
    expect(screen.queryByTestId('result')).not.toBeInTheDocument();
  });

  it('disables the submit button while loading', async () => {
    let resolve!: (v: Awaited<ReturnType<CalculateFn>>) => void;
    const calc = vi.fn<CalculateFn>().mockReturnValue(new Promise((r) => (resolve = r)));
    const ui = setup(calc);

    await ui.user.type(ui.a(), '1');
    await ui.user.type(ui.b()!, '1');
    await ui.user.click(ui.equals());

    expect(screen.getByRole('button', { name: 'Calculating…' })).toBeDisabled();
    resolve({ operation: 'add', a: 1, b: 1, result: 2 });
    expect(await screen.findByTestId('result')).toHaveTextContent('2');
  });

  it('reuses the result as the next input and clears', async () => {
    const ui = setup(ok(10));

    await ui.user.type(ui.a(), '4');
    await ui.user.type(ui.b()!, '6');
    await ui.user.click(ui.equals());
    await screen.findByTestId('result');

    await ui.user.click(screen.getByRole('button', { name: /use as next input/i }));
    expect(ui.a()).toHaveValue('10');
    expect(ui.b()).toHaveValue('');

    await ui.user.click(screen.getByRole('button', { name: 'Clear' }));
    expect(ui.a()).toHaveValue('');
    expect(screen.queryByTestId('result')).not.toBeInTheDocument();
  });

  it('keeps only the five most recent calculations', async () => {
    const ui = setup(ok(1));
    for (let i = 1; i <= 6; i++) {
      await ui.user.clear(ui.a());
      await ui.user.type(ui.a(), String(i));
      await ui.user.clear(ui.b()!);
      await ui.user.type(ui.b()!, '0');
      await ui.user.click(ui.equals());
      await screen.findByText(`${i} + 0`);
    }
    expect(screen.getAllByRole('listitem')).toHaveLength(5);
    expect(screen.queryByText('1 + 0')).not.toBeInTheDocument();
  });

  it('renders inside the App shell', () => {
    render(<App />);
    expect(screen.getByRole('heading', { name: 'Calculator' })).toBeInTheDocument();
  });
});
