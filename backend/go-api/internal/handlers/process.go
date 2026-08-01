package handlers

import (
	"net/http"

	"bardump-api/internal/services"

	"github.com/labstack/echo/v4"
)

type ProcessRequest struct {
	Input     string `json:"input"`
	Output    string `json:"output"`
	Accounts  string `json:"accounts"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func ProcessHandler(c echo.Context) error {
	var req ProcessRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	err := services.ProcessTransactions(req.Input, req.Output, req.Accounts, req.StartDate, req.EndDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}
