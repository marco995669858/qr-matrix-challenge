/**
 * Cálculo de estadísticas sobre una o más matrices.
 *
 * Funciones puras (sin Express) para poder probarlas de forma aislada.
 */

/**
 * Tolerancia para comparar con cero. La factorización QR trabaja con
 * punto flotante, así que un "cero" puede llegar como 1e-17.
 */
const EPSILON = 1e-10;

/**
 * Indica si una matriz es diagonal: todos los elementos fuera de la
 * diagonal principal son 0 (con tolerancia EPSILON).
 * Se acepta también una matriz rectangular (diagonal en sentido amplio).
 *
 * @param {number[][]} matrix
 * @returns {boolean}
 */
function isDiagonal(matrix) {
  for (let i = 0; i < matrix.length; i++) {
    for (let j = 0; j < matrix[i].length; j++) {
      if (i !== j && Math.abs(matrix[i][j]) > EPSILON) {
        return false;
      }
    }
  }
  return true;
}

/**
 * Estadísticas de una sola matriz, en un único recorrido O(filas × columnas).
 *
 * @param {number[][]} matrix
 */
function matrixStats(matrix) {
  let max = -Infinity;
  let min = Infinity;
  let sum = 0;
  let count = 0;

  for (const row of matrix) {
    for (const value of row) {
      if (value > max) max = value;
      if (value < min) min = value;
      sum += value;
      count++;
    }
  }

  return {
    rows: matrix.length,
    cols: matrix[0].length,
    max,
    min,
    sum,
    average: sum / count,
    count,
    isDiagonal: isDiagonal(matrix),
  };
}

/**
 * Estadísticas globales (todas las matrices juntas) y el detalle por matriz.
 *
 * @param {{name: string, data: number[][]}[]} matrices
 */
function computeStats(matrices) {
  const perMatrix = matrices.map(({ name, data }) => ({
    name,
    ...matrixStats(data),
  }));

  let max = -Infinity;
  let min = Infinity;
  let sum = 0;
  let count = 0;
  for (const s of perMatrix) {
    max = Math.max(max, s.max);
    min = Math.min(min, s.min);
    sum += s.sum;
    count += s.count;
  }

  return {
    max: clean(max),
    min: clean(min),
    average: clean(sum / count),
    sum: clean(sum),
    anyDiagonal: perMatrix.some((s) => s.isDiagonal),
    diagonalMatrices: perMatrix.filter((s) => s.isDiagonal).map((s) => s.name),
    matrices: perMatrix.map((s) => ({
      ...s,
      max: clean(s.max),
      min: clean(s.min),
      sum: clean(s.sum),
      average: clean(s.average),
    })),
  };
}

/**
 * Convierte el "ruido" numérico en un 0 exacto (y -0 en 0) para que la
 * respuesta sea legible.
 */
function clean(value) {
  return Math.abs(value) < EPSILON ? 0 : value;
}

module.exports = { computeStats, matrixStats, isDiagonal, EPSILON };
