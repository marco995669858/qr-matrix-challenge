# Factorización QR + Estadísticas (Go + Node.js)

Solución al Coding Challenge: dos APIs REST que se comunican por HTTP.

- **`go-api`** (Go + Fiber): recibe una matriz rectangular, calcula su **factorización QR** y envía Q y R a la API de Node.
- **`node-api`** (Node.js + Express): recibe las matrices y calcula **máximo, mínimo, promedio, suma total** y si **alguna matriz es diagonal**.

Incluye también los opcionales: **frontend**, **seguridad con JWT** y **pruebas unitarias y de integración**.

## Despliegue

| Recurso | URL |
|---|---|
| Frontend + API Go | https://matrix-qr-api.onrender.com |
| API Node | https://matrix-stats-api.onrender.com |

> Las credenciales de acceso se envían por correo. El plan gratuito de Render "duerme" los servicios inactivos, así que la primera petición puede tardar ~1 minuto.

## Arquitectura

```
                 POST /auth/login ─► JWT
Cliente / Frontend ──────────────► API Go (Fiber) :3000
        │  POST /api/qr (Bearer JWT)    │ 1. valida la matriz
        │                               │ 2. QR con rotaciones de Givens
        │                               │ 3. POST /api/stats (reenvía el JWT)
        │                               ▼
        │                        API Node (Express) :3001
        │                               │ valida el JWT y calcula estadísticas
        ◄──────── { q, r, stats } ◄─────┘
```

El cliente hace **una sola petición** y recibe la factorización y las estadísticas juntas. La API de Go es la puerta de entrada: emite el JWT y lo reenvía a Node, que lo valida con el mismo secreto (defensa en profundidad: Node no confía en quien la llame solo por estar en la red).

## Decisiones técnicas

| Ambigüedad / tema | Decisión | Motivo |
|---|---|---|
| El enunciado habla de "rotación" y de "factorización QR" | QR mediante **rotaciones de Givens** | Es un método estándar de QR que literalmente aplica rotaciones planas; cumple ambas descripciones. Numéricamente estable y válido para matrices m×n de cualquier forma. |
| Implementación propia vs. librería (gonum) | Implementación propia, sin dependencias | El algoritmo es corto (~40 líneas), queda explicado y probado; evita una dependencia pesada para una sola operación. |
| QR completa vs. reducida | **Completa**: Q es m×m, R es m×n | Q es una matriz ortogonal cuadrada (Qᵀ·Q = I), que es la definición clásica. |
| Ruido de punto flotante (`1e-17`, `-0`) | Valores con \|x\| < 1e-12 se convierten en 0; los elementos bajo la diagonal de R quedan en 0 exacto | Respuestas legibles y verificación de "diagonal" fiable. |
| "Matriz diagonal" | Todos los elementos fuera de la diagonal principal son 0 (tolerancia 1e-10); se acepta en matrices rectangulares | Es la definición general; la tolerancia evita falsos negativos por redondeo. |
| ¿Estadísticas globales o por matriz? | Ambas: globales (Q y R juntas) y detalle por matriz | El enunciado dice "las matrices"; el detalle no cuesta nada extra (un solo recorrido). |
| Límite de tamaño | Máximo 100×100 y cuerpo de 1 MB | La QR completa es O(m²·n); evita abusos en un servicio público. |
| Formato entre APIs | `{ "matrices": [{ "name", "data" }] }` | Genérico: la API de Node sirve para cualquier conjunto de matrices, no solo Q y R. |
| Frontend | Archivos estáticos servidos por la API de Go | Mismo origen (sin problemas de CORS) y un servicio menos que desplegar. |
| Configuración | Variables de entorno | La misma imagen Docker funciona en local y en la nube. |

**Complejidad:** QR con Givens es O(m²·n) en tiempo; las estadísticas son O(total de elementos) en un solo recorrido.

## API

### `POST /auth/login` (Go)
```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"tu-clave"}'
# → {"token":"eyJ...","expiresIn":3600}
```
Limitado a 10 intentos por minuto por IP.

### `POST /api/qr` (Go, requiere JWT)
```bash
curl -X POST http://localhost:3000/api/qr \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"matrix":[[12,-51,4],[6,167,-68],[-4,24,-41]]}'
```
```json
{
  "q": [[0.857, -0.394, 0.331], [0.429, 0.903, -0.034], [-0.286, 0.171, 0.943]],
  "r": [[14, 21, -14], [0, 175, -70], [0, 0, -35]],
  "stats": {
    "max": 175, "min": -70, "average": 5.2178, "sum": 93.92,
    "anyDiagonal": false, "diagonalMatrices": [],
    "matrices": [
      { "name": "Q", "rows": 3, "cols": 3, "max": 0.943, "min": -0.394, "sum": 2.92, "average": 0.324, "count": 9, "isDiagonal": false },
      { "name": "R", "rows": 3, "cols": 3, "max": 175, "min": -70, "sum": 91, "average": 10.11, "count": 9, "isDiagonal": false }
    ]
  }
}
```
*(valores redondeados en este ejemplo; la API devuelve precisión completa)*

### `POST /api/stats` (Node, requiere JWT)
```json
{ "matrices": [ { "name": "Q", "data": [[1, 0], [0, 1]] }, { "name": "R", "data": [[2, 3], [0, 4]] } ] }
```

### Errores
Todas las respuestas de error tienen la forma `{"error": "mensaje"}`:
`400` entrada inválida · `401` sin token / token inválido · `502` la API de estadísticas no responde.

### `GET /health`
Disponible en ambas APIs, pública (para los health checks de la nube).

## Ejecutar en local

### Con Docker
```bash
cp .env.example .env      # y edita los valores
docker compose up --build
# Frontend: http://localhost:3000
```

### Sin Docker
Requisitos: Go 1.25+ y Node.js 20+.
```bash
# Terminal 1
cd node-api && npm install
JWT_SECRET=dev-secret npm start

# Terminal 2
cd go-api
JWT_SECRET=dev-secret AUTH_USER=admin AUTH_PASSWORD=admin123 STATS_API_URL=http://localhost:3001 go run .
```

## Pruebas

```bash
cd go-api && go test ./...      # QR, handlers, cliente HTTP
cd node-api && npm test         # estadísticas, validación, endpoints (Jest + supertest)
```

- **QR**: verifica las tres propiedades que definen la factorización (Q·R = A, Qᵀ·Q = I, R triangular superior) en matrices cuadradas, altas, anchas y casos borde, además de un ejemplo conocido.
- **Integración**: endpoints con JWT válido/inválido, entradas inválidas, reenvío del token entre APIs y caída de la API de Node (502).

## Despliegue en la nube (Render)

El archivo [`render.yaml`](render.yaml) define ambos servicios. Render construye las imágenes desde los `Dockerfile` del repositorio:

1. En Render: **New → Blueprint** y seleccionar este repositorio.
2. Completar `AUTH_USER`, `AUTH_PASSWORD` y `STATS_API_URL` (URL pública de `matrix-stats-api`).
3. `JWT_SECRET` se genera automáticamente y se comparte entre ambos servicios.

## Estructura

```
go-api/
  main.go                  arranque
  internal/qr/             factorización QR (Givens) + tests
  internal/handlers/       rutas Fiber, middlewares + tests de integración
  internal/auth/           emisión/validación de JWT
  internal/statsclient/    cliente HTTP hacia la API de Node + tests
  internal/config/         variables de entorno
  web/                     frontend estático
node-api/
  src/stats.js             cálculo de estadísticas (funciones puras)
  src/validation.js        validación de entrada
  src/auth.js              middleware JWT
  src/app.js               app Express (separada de server.js para testear)
  tests/                   Jest + supertest
docker-compose.yml         entorno local
render.yaml                despliegue en Render
```

## Seguridad

- JWT HS256 con expiración de 1 hora; algoritmo fijado al validar (evita el ataque `alg: none`).
- Comparación de credenciales en tiempo constante y límite de intentos de login.
- Contenedores sin usuario root; la imagen de Go es *distroless* (sin shell).
- Validación de tamaño y forma de toda entrada; límite de 1 MB por petición.
