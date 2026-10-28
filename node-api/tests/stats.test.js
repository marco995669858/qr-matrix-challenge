const { computeStats, isDiagonal } = require('../src/stats');
const { validateStatsRequest } = require('../src/validation');

describe('isDiagonal', () => {
  test('identidad es diagonal', () => {
    expect(isDiagonal([[1, 0], [0, 1]])).toBe(true);
  });

  test('matriz triangular no es diagonal', () => {
    expect(isDiagonal([[1, 2], [0, 3]])).toBe(false);
  });

  test('ignora el ruido de punto flotante', () => {
    expect(isDiagonal([[5, 1e-16], [-2e-17, 3]])).toBe(true);
  });

  test('acepta matrices rectangulares', () => {
    expect(isDiagonal([[2, 0], [0, 3], [0, 0]])).toBe(true);
  });
});

describe('computeStats', () => {
  const matrices = [
    { name: 'Q', data: [[1, 0], [0, 1]] },
    { name: 'R', data: [[2, 3], [0, -4]] },
  ];

  test('calcula estadísticas globales', () => {
    const stats = computeStats(matrices);
    expect(stats.max).toBe(3);
    expect(stats.min).toBe(-4);
    expect(stats.sum).toBe(3);
    expect(stats.average).toBeCloseTo(3 / 8);
    expect(stats.anyDiagonal).toBe(true);
    expect(stats.diagonalMatrices).toEqual(['Q']);
  });

  test('devuelve el detalle por matriz', () => {
    const [q, r] = computeStats(matrices).matrices;
    expect(q).toMatchObject({ name: 'Q', rows: 2, cols: 2, sum: 2, isDiagonal: true });
    expect(r).toMatchObject({ name: 'R', max: 3, min: -4, isDiagonal: false });
  });
});

describe('validateStatsRequest', () => {
  test('acepta un cuerpo válido', () => {
    expect(validateStatsRequest({ matrices: [{ name: 'A', data: [[1]] }] })).toBeNull();
  });

  test.each([
    [{}, /no vacío/],
    [{ matrices: [] }, /no vacío/],
    [{ matrices: [{ data: [[1]] }] }, /name/],
    [{ matrices: [{ name: 'A', data: [] }] }, /no vacío/],
    [{ matrices: [{ name: 'A', data: [[1, 2], [3]] }] }, /rectangular/],
    [{ matrices: [{ name: 'A', data: [[1, 'x']] }] }, /números/],
  ])('rechaza %j', (body, message) => {
    expect(validateStatsRequest(body)).toMatch(message);
  });
});
