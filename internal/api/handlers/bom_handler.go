package handlers

import (
	"net/http"
	"time"

	"doligo_001/internal/api/dto"
	"doligo_001/internal/api/validator"
	"doligo_001/internal/domain" // For domain.UserIDFromContext
	"doligo_001/internal/domain/bom"
	bomUseCase "doligo_001/internal/usecase/bom" // Alias to avoid conflict with domain.bom
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BOMHandler handles HTTP requests related to Bill of Materials.
type BOMHandler struct {
	bomUsecase bomUseCase.BOMUsecase
	validator  *validator.CustomValidator
}

// NewBOMHandler creates a new BOMHandler.
func NewBOMHandler(bu bomUseCase.BOMUsecase, v *validator.CustomValidator) *BOMHandler {
	return &BOMHandler{
		bomUsecase: bu,
		validator:  v,
	}
}

// CreateBOM handles the creation of a new Bill of Materials.
func (h *BOMHandler) CreateBOM(c echo.Context) error {
	req := new(dto.CreateBOMRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.validator.Validate(req); err != nil {
		return err // validator already returns echo.HTTPError
	}

	components := make([]bom.BillOfMaterialsComponent, len(req.Components))
	for i, compReq := range req.Components {
		components[i] = bom.BillOfMaterialsComponent{
			ID:              uuid.New(), // ID will be overridden by DB on creation
			ComponentItemID: compReq.ComponentItemID,
			Quantity:        compReq.Quantity,
			UnitOfMeasure:   compReq.UnitOfMeasure,
			IsActive:        compReq.IsActive,
		}
	}

	// Assume user ID comes from JWT middleware context
	userID, ok := domain.UserIDFromContext(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user ID not found in context")
	}

	newBOM := &bom.BillOfMaterials{
		ID:        uuid.New(), // ID will be overridden by DB on creation
		ProductID: req.ProductID,
		Name:      req.Name,
		IsActive:  req.IsActive,
		Components: components,
	}
	newBOM.SetCreatedBy(userID)
	newBOM.SetUpdatedBy(userID) // Initial creation also sets updated by

	if err := h.bomUsecase.CreateBOM(c.Request().Context(), newBOM); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := toBOMResponse(newBOM)
	return c.JSON(http.StatusCreated, res)
}

// GetBOMByID retrieves a Bill of Materials by its ID.
func (h *BOMHandler) GetBOMByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid BOM ID format")
	}

	b, err := h.bomUsecase.GetBOMByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if b == nil {
		return echo.NewHTTPError(http.StatusNotFound, "BOM not found")
	}

	res := toBOMResponse(b)
	return c.JSON(http.StatusOK, res)
}

// GetBOMByProductID retrieves a Bill of Materials by the product it produces.
func (h *BOMHandler) GetBOMByProductID(c echo.Context) error {
	productIDStr := c.Param("productID")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid Product ID format")
	}

	b, err := h.bomUsecase.GetBOMByProductID(c.Request().Context(), productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if b == nil {
		return echo.NewHTTPError(http.StatusNotFound, "BOM not found for this product")
	}

	res := toBOMResponse(b)
	return c.JSON(http.StatusOK, res)
}

// ListBOMs retrieves all Bill of Materials.
func (h *BOMHandler) ListBOMs(c echo.Context) error {
	boms, err := h.bomUsecase.ListBOMs(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resList := make([]dto.BOMResponse, len(boms))
	for i, b := range boms {
		resList[i] = toBOMResponse(b)
	}
	return c.JSON(http.StatusOK, resList)
}

// UpdateBOM handles updating an existing Bill of Materials.
func (h *BOMHandler) UpdateBOM(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid BOM ID format")
	}

	req := new(dto.CreateBOMRequest) // Reusing CreateBOMRequest for update, adjust as needed
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.validator.Validate(req); err != nil {
		return err
	}

	existingBOM, err := h.bomUsecase.GetBOMByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if existingBOM == nil {
		return echo.NewHTTPError(http.StatusNotFound, "BOM not found")
	}

	// Update fields
	existingBOM.ProductID = req.ProductID
	existingBOM.Name = req.Name
	existingBOM.IsActive = req.IsActive

	// Handle components update - this can be complex with GORM.
	// For simplicity, we're assuming a full replacement or that components are managed separately
	// or that the Usecase will handle the diff. Here we'll pass the new set and let Usecase decide.
	newComponents := make([]bom.BillOfMaterialsComponent, len(req.Components))
	for i, compReq := range req.Components {
		newComponents[i] = bom.BillOfMaterialsComponent{
			ComponentItemID: compReq.ComponentItemID,
			Quantity:        compReq.Quantity,
			UnitOfMeasure:   compReq.UnitOfMeasure,
			IsActive:        compReq.IsActive,
		}
	}
	existingBOM.Components = newComponents

	// Assume user ID comes from JWT middleware context
	userID, ok := domain.UserIDFromContext(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user ID not found in context")
	}
	existingBOM.SetUpdatedBy(userID)

	if err := h.bomUsecase.UpdateBOM(c.Request().Context(), existingBOM); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := toBOMResponse(existingBOM)
	return c.JSON(http.StatusOK, res)
}

// DeleteBOM handles deleting a Bill of Materials.
func (h *BOMHandler) DeleteBOM(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid BOM ID format")
	}

	if err := h.bomUsecase.DeleteBOM(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}


// CalculatePredictiveCost handles the request for predictive cost calculation.
func (h *BOMHandler) CalculatePredictiveCost(c echo.Context) error {
	req := new(dto.CalculateCostRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.validator.Validate(req); err != nil {
		return err
	}

	totalCost, err := h.bomUsecase.CalculatePredictiveCost(c.Request().Context(), req.BOMID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := dto.CalculateCostResponse{
		BOMID:    req.BOMID,
		TotalCost: totalCost,
	}
	return c.JSON(http.StatusOK, res)
}

// ProduceItem handles the request for initiating a production order.
func (h *BOMHandler) ProduceItem(c echo.Context) error {
	req := new(dto.ProduceItemRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.validator.Validate(req); err != nil {
		return err
	}

	// Assume user ID comes from JWT middleware context
	userID, ok := domain.UserIDFromContext(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user ID not found in context")
	}

	productionRecordID, actualProductionCost, err := h.bomUsecase.ProduceItem(
		c.Request().Context(),
		req.BOMID,
		req.WarehouseID,
		userID,
		req.ProductionQuantity,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := dto.ProduceItemResponse{
		ProductionRecordID:   productionRecordID,
		ActualProductionCost: actualProductionCost,
		Message:             "Production order successfully processed.",
	}
	return c.JSON(http.StatusOK, res)
}


// toBOMResponse converts a domain.bom.BillOfMaterials entity to a dto.BOMResponse.
func toBOMResponse(b *bom.BillOfMaterials) dto.BOMResponse {
	components := make([]dto.BOMComponentResponse, len(b.Components))
	for i, comp := range b.Components {
		components[i] = dto.BOMComponentResponse{
			ID:                comp.ID,
			BillOfMaterialsID: comp.BillOfMaterialsID,
			ComponentItemID:   comp.ComponentItemID,
			Quantity:          comp.Quantity,
			UnitOfMeasure:     comp.UnitOfMeasure,
			IsActive:          comp.IsActive,
			CreatedAt:         comp.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         comp.UpdatedAt.Format(time.RFC3339),
			CreatedBy:         comp.CreatedBy,
			UpdatedBy:         comp.UpdatedBy,
		}
	}
	return dto.BOMResponse{
		ID:         b.ID,
		ProductID:  b.ProductID,
		Name:       b.Name,
		IsActive:   b.IsActive,
		Components: components,
		CreatedAt:  b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  b.UpdatedAt.Format(time.RFC3339),
		CreatedBy:  b.CreatedBy,
		UpdatedBy:  b.UpdatedBy,
	}
}

