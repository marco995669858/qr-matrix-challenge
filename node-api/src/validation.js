/**
 * Validación del cuerpo de POST /api/stats.
 *
 * Formato esperado:
 * {
 *   "matrices": [
 *     { "name": "Q", "data": [[1, 0], [0, 1]] },
 *     { "name": "R", "data": [[2, 3], [0, 4]] }
 *   ]
 * }
 */

const MAX_MATRICES = 10;
const MAX_DIMENSION = 100;

/**
 * @returns {string|null} mensaje de error, o null si el cuerpo es válido
 */
function validateStatsRequest(body) {
  if (!body || !Array.isArray(body.matrices) || body.matrices.length === 0) {
    return '"matrices" debe ser un arreglo no vacío';
  }
  if (body.matrices.length > MAX_MATRICES) {
    return `se permiten como máximo ${MAX_MATRICES} matrices`;
  }

  for (const [index, item] of body.matrices.entries()) {
    const label = item && typeof item.name === 'string' ? item.name : `#${index}`;
    if (!item || typeof item.name !== 'string' || item.name.trim() === '') {
      return `la matriz ${label} debe tener un "name" (texto)`;
    }
    const error = validateMatrix(item.data);
    if (error) return `matriz ${label}: ${error}`;
  }
  return null;
}

/**
 * Comprueba que sea una matriz rectangular, no vacía y solo con números finitos.
 */
function validateMatrix(data) {
  if (!Array.isArray(data) || data.length === 0) {
    return '"data" debe ser un arreglo de filas no vacío';
  }
  if (data.length > MAX_DIMENSION) {
    return `máximo ${MAX_DIMENSION} filas`;
  }
  const cols = Array.isArray(data[0]) ? data[0].length : 0;
  if (cols === 0 || cols > MAX_DIMENSION) {
    return `cada fila debe tener entre 1 y ${MAX_DIMENSION} columnas`;
  }
  for (const [i, row] of data.entries()) {
    if (!Array.isArray(row) || row.length !== cols) {
      return `la fila ${i} no tiene ${cols} columnas (la matriz debe ser rectangular)`;
    }
    if (!row.every((v) => typeof v === 'number' && Number.isFinite(v))) {
      return `la fila ${i} contiene valores que no son números`;
    }
  }
  return null;
}

module.exports = { validateStatsRequest, validateMatrix };
