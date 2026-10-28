const express = require('express');
const cors = require('cors');
const { requireJwt } = require('./auth');
const { validateStatsRequest } = require('./validation');
const { computeStats } = require('./stats');

/**
 * Crea la aplicación Express. Se separa de server.js para poder
 * probarla con supertest sin abrir un puerto.
 *
 * @param {{ jwtSecret: string }} config
 */
function createApp({ jwtSecret }) {
  const app = express();

  app.use(cors());
  app.use(express.json({ limit: '1mb' }));

  // Pública: la usa el proveedor de nube para saber si el servicio está vivo.
  app.get('/health', (req, res) => {
    res.json({ status: 'ok', service: 'matrix-stats-api' });
  });

  /**
   * POST /api/stats
   * Recibe las matrices (Q y R desde la API de Go) y devuelve
   * máximo, mínimo, promedio, suma y si alguna es diagonal.
   */
  app.post('/api/stats', requireJwt(jwtSecret), (req, res) => {
    const error = validateStatsRequest(req.body);
    if (error) {
      return res.status(400).json({ error });
    }
    return res.json(computeStats(req.body.matrices));
  });

  app.use((req, res) => {
    res.status(404).json({ error: 'ruta no encontrada' });
  });

  // JSON mal formado u otros errores no controlados.
  app.use((err, req, res, next) => { 
    if (err.type === 'entity.parse.failed') {
      return res.status(400).json({ error: 'el cuerpo no es un JSON válido' });
    }
    console.error(err);
    return res.status(500).json({ error: 'error interno' });
  });

  return app;
}

module.exports = { createApp };
