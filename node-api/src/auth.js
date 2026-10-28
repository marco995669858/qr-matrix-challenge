const jwt = require('jsonwebtoken');

/**
 * Middleware que exige un JWT válido en "Authorization: Bearer <token>".
 * El token lo emite la API de Go; ambas APIs comparten JWT_SECRET.
 *
 * @param {string} secret
 */
function requireJwt(secret) {
  return (req, res, next) => {
    const header = req.get('Authorization') || '';
    const [scheme, token] = header.split(' ');

    if (scheme !== 'Bearer' || !token) {
      return res.status(401).json({ error: 'falta el token (Authorization: Bearer TOKEN)' });
    }

    try {
      // Fijamos el algoritmo para evitar ataques de confusión de algoritmo.
      req.user = jwt.verify(token, secret, { algorithms: ['HS256'] });
      return next();
    } catch {
      return res.status(401).json({ error: 'token inválido o expirado' });
    }
  };
}

module.exports = { requireJwt };
