package handlers

import (
	"fmt"
	"net/http"
	"time"

	"bardump-api/internal/services"

	"github.com/labstack/echo/v4"
)

type MailRequest struct {
	AccountsFile string `json:"accounts_file"`
	DumpID       string `json:"dump_id"`
}

func SendEmailsHandler(c echo.Context) error {
	var req MailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	logChan := make(chan string)
	go services.ProcessAndSendEmails(c.Request().Context(), req.AccountsFile, req.DumpID, logChan)

	// Set status code to 200 explicitly before flushing
	c.Response().WriteHeader(http.StatusOK)

	for {
		select {
		case msg, ok := <-logChan:
			if !ok {
				return nil
			}
			fmt.Fprintf(c.Response().Writer, "data: %s\n\n", msg)
			c.Response().Flush()
		case <-c.Request().Context().Done():
			// Client disconnected
			return nil
		case <-time.After(30 * time.Second):
			// Keep-alive ping
			fmt.Fprintf(c.Response().Writer, ": ping\n\n")
			c.Response().Flush()
		}
	}
}
