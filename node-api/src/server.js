const { createApp } = require('./app');

const port = Number(process.env.PORT) || 3001;
const jwtSecret = process.env.JWT_SECRET;

if (!jwtSecret) {
  console.error('Falta la variable de entorno JWT_SECRET');
  process.exit(1);
}

createApp({ jwtSecret }).listen(port, () => {
  console.log(`matrix-stats-api escuchando en el puerto ${port}`);
});
