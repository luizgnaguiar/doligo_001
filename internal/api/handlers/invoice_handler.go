package handlers

import (
	"net/http"

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
