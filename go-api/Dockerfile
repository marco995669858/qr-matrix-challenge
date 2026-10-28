# Etapa 1: compilar un binario estático.
FROM golang:1.25-alpine AS build
WORKDIR /src

# Copiar primero go.mod/go.sum aprovecha la caché de Docker para las dependencias.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server .

# Etapa 2: imagen mínima y usuario no root.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server ./server
COPY web ./web

EXPOSE 3000
ENTRYPOINT ["/app/server"]
