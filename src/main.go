package main

import "github.com/CPU-commits/Intranet_BNews/src/server"

// @title          News API
// @version        1.0
// @description    API Server News service
// @termsOfService http://swagger.io/terms/

// @contact.name  API Support
// @contact.url   http://www.swagger.io/support
// @contact.email support@swagger.io

// lincense.name  Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @tag.name        news
// @tag.description Service of news

// @host     localhost:8080
// @BasePath /api/news

// @securityDefinitions.apikey ApiKeyAuth
// @in                         header
// @name                       Authorization
// @description                BearerJWTToken in Authorization Header

// @accept  json
// @produce json

// @schemes http https
func main() {
	server.Init()
}
