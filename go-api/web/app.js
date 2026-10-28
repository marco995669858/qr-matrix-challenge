// Frontend: login → POST /api/qr → muestra Q, R y estadísticas.
// Se sirve desde la API de Go (mismo origen), así que usa rutas relativas.

const $ = (id) => document.getElementById(id);
const TOKEN_KEY = 'qr-token';

function getToken() {
  try { return sessionStorage.getItem(TOKEN_KEY); } catch { return null; }
}

function setToken(token) {
  try {
    if (token) sessionStorage.setItem(TOKEN_KEY, token);
    else sessionStorage.removeItem(TOKEN_KEY);
  } catch { /* almacenamiento no disponible */ }
  render();
}

function render() {
  const logged = Boolean(getToken());
  $('login-section').hidden = logged;
  $('matrix-section').hidden = !logged;
  if (!logged) $('results').hidden = true;
}

function showError(message) {
  $('error').textContent = message;
  $('error').hidden = !message;
}

async function api(path, body) {
  const headers = { 'Content-Type': 'application/json' };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(path, { method: 'POST', headers, body: JSON.stringify(body) });
  const data = await res.json().catch(() => ({}));
  if (res.status === 401 && path !== '/auth/login') setToken(null); // token vencido
  if (!res.ok) throw new Error(data.error || `Error ${res.status}`);
  return data;
}

// Acepta JSON ([[1,2],[3,4]]) o texto: una fila por línea.
function parseMatrix(text) {
  const trimmed = text.trim();
  const matrix = trimmed.startsWith('[')
    ? JSON.parse(trimmed)
    : trimmed
      .split('\n')
      .filter((line) => line.trim())
      .map((line) => line.trim().split(/[\s,;]+/).map(Number));

  // Number('a') es NaN, y JSON lo enviaría como null: mejor avisar aquí.
  if (matrix.flat().some((v) => typeof v !== 'number' || Number.isNaN(v))) {
    throw new Error('La matriz solo puede contener números');
  }
  return matrix;
}

// Redondea a 4 decimales y quita ceros sobrantes: 21.000000000000004 → "21".
const fmt = (n) => String(Number(n.toFixed(4)));

function matrixTable(matrix) {
  const table = document.createElement('table');
  table.className = 'matrix';
  for (const row of matrix) {
    const tr = table.insertRow();
    for (const v of row) tr.insertCell().textContent = fmt(v);
  }
  return table;
}

function renderResults({ q, r, stats }) {
  $('matrix-q').replaceChildren(matrixTable(q));
  $('matrix-r').replaceChildren(matrixTable(r));

  const items = [
    ['Máximo', fmt(stats.max)],
    ['Mínimo', fmt(stats.min)],
    ['Promedio', fmt(stats.average)],
    ['Suma total', fmt(stats.sum)],
    ['¿Alguna diagonal?', stats.anyDiagonal ? `Sí (${stats.diagonalMatrices.join(', ')})` : 'No'],
  ];
  $('stats').replaceChildren(...items.map(([label, value]) => {
    const div = document.createElement('div');
    div.className = 'stat';
    div.innerHTML = '<div class="label"></div><div class="value"></div>';
    div.querySelector('.label').textContent = label;
    div.querySelector('.value').textContent = value;
    return div;
  }));
  $('results').hidden = false;
}

async function withLoading(form, fn) {
  const button = form.querySelector('button[type="submit"]');
  button.disabled = true;
  showError('');
  try { await fn(); } catch (err) { showError(err.message); } finally { button.disabled = false; }
}

$('login-form').addEventListener('submit', (e) => {
  e.preventDefault();
  const form = e.target;
  withLoading(form, async () => {
    const { token } = await api('/auth/login', {
      username: form.username.value,
      password: form.password.value,
    });
    setToken(token);
  });
});

$('matrix-form').addEventListener('submit', (e) => {
  e.preventDefault();
  withLoading(e.target, async () => {
    let matrix;
    try { matrix = parseMatrix($('matrix-input').value); } catch (err) {
      throw new Error(err instanceof SyntaxError ? 'JSON inválido' : err.message);
    }
    renderResults(await api('/api/qr', { matrix }));
  });
});

document.querySelectorAll('[data-example]').forEach((btn) => {
  btn.addEventListener('click', () => { $('matrix-input').value = btn.dataset.example; });
});

$('logout').addEventListener('click', () => setToken(null));

render();
