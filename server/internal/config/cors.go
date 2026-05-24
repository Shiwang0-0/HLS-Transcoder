package config

import "github.com/gofiber/fiber/v2/middleware/cors"

var CorsConfig = cors.New(cors.Config{
	AllowOrigins: "http://localhost:5173",
	AllowMethods: "GET,POST,PUT,DELETE",
	AllowHeaders: "Origin, Content-Type, Accept",
})
