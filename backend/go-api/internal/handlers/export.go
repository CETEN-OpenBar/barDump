package handlers

import (
	"log"
	"net/http"

	"bardump-api/internal/services"

	"github.com/labstack/echo/v4"
)

type ExportRequest struct {
	Output string `json:"output"`
}

func ExportHandler(c echo.Context) error {
	var req ExportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := services.ExportData(req.Output); err != nil {
		log.Printf("Export failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}
