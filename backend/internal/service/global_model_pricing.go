package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var ErrGlobalModelPricingNotFound = errors.New("global model pricing override not found")

type GlobalModelPricingOverrideRepository interface {
	GetByModel(ctx context.Context, model string) (*GlobalModelPricingOverride, error)
	List(ctx context.Context) ([]GlobalModelPricingOverride, error)
	Upsert(ctx context.Context, override *GlobalModelPricingOverride) error
	DeleteByModel(ctx context.Context, model string) error
}

type GlobalModelPricingOverride struct {
	ID               int64       `json:"id"`
	Model            string      `json:"model"`
	Provider         string      `json:"provider"`
	BillingMode      BillingMode `json:"billing_mode"`
	InputPrice       *float64    `json:"input_price"`
	OutputPrice      *float64    `json:"output_price"`
	CacheWritePrice  *float64    `json:"cache_write_price"`
	CacheReadPrice   *float64    `json:"cache_read_price"`
	ImageOutputPrice *float64    `json:"image_output_price"`
	PerRequestPrice  *float64    `json:"per_request_price"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

func (o *GlobalModelPricingOverride) ToChannelPricing() *ChannelModelPricing {
	if o == nil {
		return nil
	}
	mode := o.BillingMode
	if mode == "" {
		mode = BillingModeToken
	}
	return &ChannelModelPricing{
		Platform:         o.Provider,
		Models:           []string{o.Model},
		BillingMode:      mode,
		InputPrice:       o.InputPrice,
		OutputPrice:      o.OutputPrice,
		CacheWritePrice:  o.CacheWritePrice,
		CacheReadPrice:   o.CacheReadPrice,
		ImageOutputPrice: o.ImageOutputPrice,
		PerRequestPrice:  o.PerRequestPrice,
	}
}

type GlobalModelPricingFilters struct {
	Search      string
	Provider    string
	Source      string
	BillingMode string
}

type GlobalModelPricingRow struct {
	Model            string                      `json:"model"`
	Provider         string                      `json:"provider"`
	BillingMode      string                      `json:"billing_mode"`
	Source           string                      `json:"source"`
	InputPrice       *float64                    `json:"input_price"`
	OutputPrice      *float64                    `json:"output_price"`
	CacheWritePrice  *float64                    `json:"cache_write_price"`
	CacheReadPrice   *float64                    `json:"cache_read_price"`
	ImageOutputPrice *float64                    `json:"image_output_price"`
	PerRequestPrice  *float64                    `json:"per_request_price"`
	Fallback         *GlobalModelPricingSnapshot `json:"fallback"`
	Override         *GlobalModelPricingOverride `json:"override"`
	UpdatedAt        *time.Time                  `json:"updated_at"`
}

type GlobalModelPricingSnapshot struct {
	InputPrice       *float64 `json:"input_price"`
	OutputPrice      *float64 `json:"output_price"`
	CacheWritePrice  *float64 `json:"cache_write_price"`
	CacheReadPrice   *float64 `json:"cache_read_price"`
	ImageOutputPrice *float64 `json:"image_output_price"`
	PerRequestPrice  *float64 `json:"per_request_price"`
}

type GlobalModelPricingListResult struct {
	Items      []GlobalModelPricingRow
	Pagination *pagination.PaginationResult
}

type GlobalModelPricingService struct {
	repo           GlobalModelPricingOverrideRepository
	billingService *BillingService
}

func NewGlobalModelPricingService(repo GlobalModelPricingOverrideRepository, billingService *BillingService) *GlobalModelPricingService {
	return &GlobalModelPricingService{repo: repo, billingService: billingService}
}

func (s *GlobalModelPricingService) List(ctx context.Context, params pagination.PaginationParams, filters GlobalModelPricingFilters) (*GlobalModelPricingListResult, error) {
	known := map[string]GlobalModelPricingRow{}
	if s.billingService != nil {
		for _, item := range s.billingService.ListKnownModelPricing() {
			row := rowFromKnownModelPricing(item)
			known[strings.ToLower(row.Model)] = row
		}
	}

	overrides, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range overrides {
		override := overrides[i]
		key := strings.ToLower(override.Model)
		row, ok := known[key]
		if !ok {
			row = GlobalModelPricingRow{
				Model:       override.Model,
				Provider:    override.Provider,
				BillingMode: string(defaultBillingMode(override.BillingMode)),
				Fallback:    &GlobalModelPricingSnapshot{},
			}
		}
		applyGlobalOverrideToRow(&row, &override)
		known[key] = row
	}

	rows := make([]GlobalModelPricingRow, 0, len(known))
	for _, row := range known {
		if globalModelPricingRowMatches(row, filters) {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Provider == rows[j].Provider {
			return rows[i].Model < rows[j].Model
		}
		return rows[i].Provider < rows[j].Provider
	})

	pageSize := params.Limit()
	page := params.Page
	if page < 1 {
		page = 1
	}
	total := len(rows)
	offset := (page - 1) * pageSize
	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}
	pages := 0
	if pageSize > 0 {
		pages = (total + pageSize - 1) / pageSize
	}

	return &GlobalModelPricingListResult{
		Items: rows[offset:end],
		Pagination: &pagination.PaginationResult{
			Total:    int64(total),
			Page:     page,
			PageSize: pageSize,
			Pages:    pages,
		},
	}, nil
}

func (s *GlobalModelPricingService) Upsert(ctx context.Context, override *GlobalModelPricingOverride) (*GlobalModelPricingOverride, error) {
	if err := validateGlobalModelPricingOverride(override); err != nil {
		return nil, err
	}
	normalizeGlobalModelPricingOverride(override)
	if err := s.repo.Upsert(ctx, override); err != nil {
		return nil, err
	}
	return s.repo.GetByModel(ctx, override.Model)
}

func (s *GlobalModelPricingService) DeleteByModel(ctx context.Context, model string) error {
	model = strings.TrimSpace(model)
	if model == "" {
		return errors.New("model is required")
	}
	return s.repo.DeleteByModel(ctx, model)
}

func rowFromKnownModelPricing(item KnownModelPricing) GlobalModelPricingRow {
	snapshot := &GlobalModelPricingSnapshot{
		InputPrice:       nonZeroFloatPtr(item.Pricing.InputPricePerToken),
		OutputPrice:      nonZeroFloatPtr(item.Pricing.OutputPricePerToken),
		CacheWritePrice:  nonZeroFloatPtr(item.Pricing.CacheCreationPricePerToken),
		CacheReadPrice:   nonZeroFloatPtr(item.Pricing.CacheReadPricePerToken),
		ImageOutputPrice: nonZeroFloatPtr(item.Pricing.ImageOutputPricePerToken),
		PerRequestPrice:  nonZeroFloatPtr(item.OutputCostPerImage),
	}
	return GlobalModelPricingRow{
		Model:            item.Model,
		Provider:         item.Provider,
		BillingMode:      string(BillingModeToken),
		Source:           item.Source,
		InputPrice:       snapshot.InputPrice,
		OutputPrice:      snapshot.OutputPrice,
		CacheWritePrice:  snapshot.CacheWritePrice,
		CacheReadPrice:   snapshot.CacheReadPrice,
		ImageOutputPrice: snapshot.ImageOutputPrice,
		PerRequestPrice:  snapshot.PerRequestPrice,
		Fallback:         snapshot,
	}
}

func applyGlobalOverrideToRow(row *GlobalModelPricingRow, override *GlobalModelPricingOverride) {
	overrideCopy := *override
	row.Override = &overrideCopy
	row.Source = PricingSourceGlobalOverride
	if override.Provider != "" {
		row.Provider = override.Provider
	}
	row.BillingMode = string(defaultBillingMode(override.BillingMode))
	if override.InputPrice != nil {
		row.InputPrice = override.InputPrice
	}
	if override.OutputPrice != nil {
		row.OutputPrice = override.OutputPrice
	}
	if override.CacheWritePrice != nil {
		row.CacheWritePrice = override.CacheWritePrice
	}
	if override.CacheReadPrice != nil {
		row.CacheReadPrice = override.CacheReadPrice
	}
	if override.ImageOutputPrice != nil {
		row.ImageOutputPrice = override.ImageOutputPrice
	}
	if override.PerRequestPrice != nil {
		row.PerRequestPrice = override.PerRequestPrice
	}
	row.UpdatedAt = &override.UpdatedAt
}

func globalModelPricingRowMatches(row GlobalModelPricingRow, filters GlobalModelPricingFilters) bool {
	search := strings.ToLower(strings.TrimSpace(filters.Search))
	if search != "" && !strings.Contains(strings.ToLower(row.Model), search) && !strings.Contains(strings.ToLower(row.Provider), search) {
		return false
	}
	if provider := strings.ToLower(strings.TrimSpace(filters.Provider)); provider != "" && strings.ToLower(row.Provider) != provider {
		return false
	}
	if source := strings.ToLower(strings.TrimSpace(filters.Source)); source != "" && strings.ToLower(row.Source) != source {
		return false
	}
	if mode := strings.ToLower(strings.TrimSpace(filters.BillingMode)); mode != "" && strings.ToLower(row.BillingMode) != mode {
		return false
	}
	return true
}

func validateGlobalModelPricingOverride(override *GlobalModelPricingOverride) error {
	if override == nil {
		return errors.New("override is required")
	}
	if strings.TrimSpace(override.Model) == "" {
		return errors.New("model is required")
	}
	if !override.BillingMode.IsValid() {
		return errors.New("billing_mode is invalid")
	}
	prices := []*float64{
		override.InputPrice,
		override.OutputPrice,
		override.CacheWritePrice,
		override.CacheReadPrice,
		override.ImageOutputPrice,
		override.PerRequestPrice,
	}
	for _, price := range prices {
		if price != nil && *price < 0 {
			return errors.New("prices must be >= 0")
		}
	}
	return nil
}

func normalizeGlobalModelPricingOverride(override *GlobalModelPricingOverride) {
	override.Model = strings.TrimSpace(override.Model)
	override.Provider = strings.TrimSpace(override.Provider)
	override.BillingMode = defaultBillingMode(override.BillingMode)
}

func defaultBillingMode(mode BillingMode) BillingMode {
	if mode == "" {
		return BillingModeToken
	}
	return mode
}

func nonZeroFloatPtr(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}
