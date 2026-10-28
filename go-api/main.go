// matrix-qr-api: API REST (Fiber) que recibe una matriz, calcula su
// factorización QR y obtiene estadísticas desde la API de Node.js.
package main

import (
	"log"

	"matrix-qr-api/internal/config"
	"matrix-qr-api/internal/handlers"
	"matrix-qr-api/internal/statsclient"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app := handlers.NewApp(cfg, statsclient.New(cfg.StatsAPIURL))

	log.Printf("matrix-qr-api escuchando en el puerto %s (stats: %s)", cfg.Port, cfg.StatsAPIURL)
	log.Fatal(app.Listen(":" + cfg.Port))
}
