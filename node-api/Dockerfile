# Imagen ligera de Node con solo las dependencias de producción.
FROM node:22-alpine

WORKDIR /app
ENV NODE_ENV=production

COPY package*.json ./
RUN npm ci --omit=dev

COPY src ./src

# Ejecutar sin privilegios de root.
USER node
EXPOSE 3001
CMD ["node", "src/server.js"]
