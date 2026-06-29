package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type globalModelPricingRepository struct {
	db *sql.DB
}

func NewGlobalModelPricingRepository(db *sql.DB) service.GlobalModelPricingOverrideRepository {
	return &globalModelPricingRepository{db: db}
}

func (r *globalModelPricingRepository) GetByModel(ctx context.Context, model string) (*service.GlobalModelPricingOverride, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, model, provider, billing_mode, input_price, output_price, cache_write_price, cache_read_price, image_output_price, per_request_price, created_at, updated_at
		 FROM global_model_pricing_overrides
		 WHERE LOWER(model) = LOWER($1)`,
		model,
	)
	override, err := scanGlobalModelPricingOverride(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get global model pricing override: %w", err)
	}
	return override, nil
}

func (r *globalModelPricingRepository) List(ctx context.Context) ([]service.GlobalModelPricingOverride, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, model, provider, billing_mode, input_price, output_price, cache_write_price, cache_read_price, image_output_price, per_request_price, created_at, updated_at
		 FROM global_model_pricing_overrides
		 ORDER BY provider, model`,
	)
	if err != nil {
		return nil, fmt.Errorf("list global model pricing overrides: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []service.GlobalModelPricingOverride
	for rows.Next() {
		override, err := scanGlobalModelPricingOverride(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *override)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate global model pricing overrides: %w", err)
	}
	return result, nil
}

func (r *globalModelPricingRepository) Upsert(ctx context.Context, override *service.GlobalModelPricingOverride) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO global_model_pricing_overrides
		 (model, provider, billing_mode, input_price, output_price, cache_write_price, cache_read_price, image_output_price, per_request_price)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT ((LOWER(model))) DO UPDATE SET
		   model = EXCLUDED.model,
		   provider = EXCLUDED.provider,
		   billing_mode = EXCLUDED.billing_mode,
		   input_price = EXCLUDED.input_price,
		   output_price = EXCLUDED.output_price,
		   cache_write_price = EXCLUDED.cache_write_price,
		   cache_read_price = EXCLUDED.cache_read_price,
		   image_output_price = EXCLUDED.image_output_price,
		   per_request_price = EXCLUDED.per_request_price,
		   updated_at = NOW()
		 RETURNING id, created_at, updated_at`,
		override.Model,
		override.Provider,
		string(override.BillingMode),
		override.InputPrice,
		override.OutputPrice,
		override.CacheWritePrice,
		override.CacheReadPrice,
		override.ImageOutputPrice,
		override.PerRequestPrice,
	).Scan(&override.ID, &override.CreatedAt, &override.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert global model pricing override: %w", err)
	}
	return nil
}

func (r *globalModelPricingRepository) DeleteByModel(ctx context.Context, model string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM global_model_pricing_overrides WHERE LOWER(model) = LOWER($1)`,
		model,
	)
	if err != nil {
		return fmt.Errorf("delete global model pricing override: %w", err)
	}
	return nil
}

type globalModelPricingScanner interface {
	Scan(dest ...any) error
}

func scanGlobalModelPricingOverride(scanner globalModelPricingScanner) (*service.GlobalModelPricingOverride, error) {
	var override service.GlobalModelPricingOverride
	var billingMode string
	if err := scanner.Scan(
		&override.ID,
		&override.Model,
		&override.Provider,
		&billingMode,
		&override.InputPrice,
		&override.OutputPrice,
		&override.CacheWritePrice,
		&override.CacheReadPrice,
		&override.ImageOutputPrice,
		&override.PerRequestPrice,
		&override.CreatedAt,
		&override.UpdatedAt,
	); err != nil {
		return nil, err
	}
	override.BillingMode = service.BillingMode(billingMode)
	return &override, nil
}
