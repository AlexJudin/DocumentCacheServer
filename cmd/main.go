package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
)

// @title Document Cache Server API
// @version 1.0
// @description API для управления документами и кэширования
// @termsOfService http://swagger.io/terms/

// @contact.name Alexey Yudin
// @contact.url http://www.swagger.io/support
// @contact.email spdante@mail.ru

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:7540
// @BasePath /http
// @query.collection.format multi

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// init config
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(cfg.LogLevel)

	startApp(cfg)
}
