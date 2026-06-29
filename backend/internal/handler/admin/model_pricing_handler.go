package admin

import (
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ModelPricingHandler struct {
	service *service.GlobalModelPricingService
}

func NewModelPricingHandler(service *service.GlobalModelPricingService) *ModelPricingHandler {
	return &ModelPricingHandler{service: service}
}

type upsertGlobalModelPricingOverrideRequest struct {
	Model            string   `json:"model" binding:"required,max=255"`
	Provider         string   `json:"provider" binding:"omitempty,max=50"`
	BillingMode      string   `json:"billing_mode" binding:"omitempty,oneof=token per_request image"`
	InputPrice       *float64 `json:"input_price" binding:"omitempty,min=0"`
	OutputPrice      *float64 `json:"output_price" binding:"omitempty,min=0"`
	CacheWritePrice  *float64 `json:"cache_write_price" binding:"omitempty,min=0"`
	CacheReadPrice   *float64 `json:"cache_read_price" binding:"omitempty,min=0"`
	ImageOutputPrice *float64 `json:"image_output_price" binding:"omitempty,min=0"`
	PerRequestPrice  *float64 `json:"per_request_price" binding:"omitempty,min=0"`
}

func (h *ModelPricingHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	result, err := h.service.List(c.Request.Context(), pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}, service.GlobalModelPricingFilters{
		Search:      strings.TrimSpace(c.Query("search")),
		Provider:    strings.TrimSpace(c.Query("provider")),
		Source:      strings.TrimSpace(c.Query("source")),
		BillingMode: strings.TrimSpace(c.Query("billing_mode")),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, result.Pagination.Total, result.Pagination.Page, result.Pagination.PageSize)
}

func (h *ModelPricingHandler) UpsertOverride(c *gin.Context) {
	var req upsertGlobalModelPricingOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_MODEL_PRICING_REQUEST", err.Error()))
		return
	}

	override, err := h.service.Upsert(c.Request.Context(), &service.GlobalModelPricingOverride{
		Model:            req.Model,
		Provider:         req.Provider,
		BillingMode:      service.BillingMode(req.BillingMode),
		InputPrice:       req.InputPrice,
		OutputPrice:      req.OutputPrice,
		CacheWritePrice:  req.CacheWritePrice,
		CacheReadPrice:   req.CacheReadPrice,
		ImageOutputPrice: req.ImageOutputPrice,
		PerRequestPrice:  req.PerRequestPrice,
	})
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_MODEL_PRICING_OVERRIDE", err.Error()))
		return
	}
	response.Success(c, override)
}

func (h *ModelPricingHandler) DeleteOverride(c *gin.Context) {
	model := strings.TrimSpace(c.Query("model"))
	if model == "" {
		response.ErrorFrom(c, infraerrors.BadRequest("MISSING_PARAMETER", "model parameter is required").
			WithMetadata(map[string]string{"param": "model"}))
		return
	}
	if err := h.service.DeleteByModel(c.Request.Context(), model); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Model pricing override deleted successfully"})
}
