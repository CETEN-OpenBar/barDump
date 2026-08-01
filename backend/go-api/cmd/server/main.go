package main

import (
	"log"

	"bardump-api/internal/handlers"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	godotenv.Load("../../wrapped-app/.env")

	e := echo.New()

	e.POST("/api/export", handlers.ExportHandler)
	e.POST("/api/process", handlers.ProcessHandler)
	e.POST("/api/send-emails", handlers.SendEmailsHandler)

	log.Println("Starting Go API (Echo) on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
