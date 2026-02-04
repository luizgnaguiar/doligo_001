package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"doligo_001/internal/api/dto"
	"doligo_001/internal/usecase/invoice"
)

type InvoiceHandler struct {
	usecase invoice.Usecase
}

func NewInvoiceHandler(usecase invoice.Usecase) *InvoiceHandler {
	return &InvoiceHandler{usecase: usecase}
}

func (h *InvoiceHandler) CreateInvoice(c echo.Context) error {
	var req dto.CreateInvoiceRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	createdInvoice, err := h.usecase.Create(c.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, createdInvoice)
}

func (h *InvoiceHandler) GetInvoice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	inv, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Invoice not found")
	}

	return c.JSON(http.StatusOK, inv)
}

func (h *InvoiceHandler) GenerateInvoicePDF(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	pdfBytes, filename, err := h.usecase.GenerateInvoicePDF(c.Request().Context(), id)
	if err != nil {
		// Consider more specific error handling (e.g., not found vs. internal error)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to generate PDF: %v", err))
	}

	// Set headers to prompt download
	c.Response().Header().Set(echo.HeaderContentType, "application/pdf")
	disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename))
	c.Response().Header().Set(echo.HeaderContentDisposition, disposition)

	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}
