//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type usageServiceLogRepoStub struct {
	UsageLogRepository

	createCalls int
	lastLog     *UsageLog
	inserted    bool
}

func (s *usageServiceLogRepoStub) Create(_ context.Context, log *UsageLog) (bool, error) {
	s.createCalls++
	s.lastLog = log
	return s.inserted, nil
}

type usageServiceUserRepoStub struct {
	UserRepository

	updateBalanceCalls int
	lastBalanceAmount  float64
}

func (s *usageServiceUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	return &User{ID: id}, nil
}

func (s *usageServiceUserRepoStub) UpdateBalance(_ context.Context, _ int64, amount float64) error {
	s.updateBalanceCalls++
	s.lastBalanceAmount = amount
	return nil
}

func newUsageServiceEntClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:usage_service_byo?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestUsageServiceCreate_ConnectedAccountCapacityWritesLogWithoutBalanceCharge(t *testing.T) {
	usageRepo := &usageServiceLogRepoStub{inserted: true}
	userRepo := &usageServiceUserRepoStub{}
	svc := NewUsageService(usageRepo, userRepo, newUsageServiceEntClient(t), nil)

	log, err := svc.Create(context.Background(), CreateUsageLogRequest{
		UserID:         42,
		APIKeyID:       7,
		AccountID:      9,
		RequestID:      "usage-service-byo",
		Model:          "gpt-5.1",
		InputTokens:    100,
		OutputTokens:   20,
		TotalCost:      0.25,
		ActualCost:     0.25,
		RateMultiplier: 1,
		CapacitySource: CapacitySourceConnectedAccount,
	})

	require.NoError(t, err)
	require.NotNil(t, log)
	require.Equal(t, 1, usageRepo.createCalls)
	require.NotNil(t, usageRepo.lastLog)
	require.Equal(t, CapacitySourceConnectedAccount, usageRepo.lastLog.CapacitySource)
	require.Equal(t, CapacitySourceConnectedAccount, log.CapacitySource)
	require.Zero(t, usageRepo.lastLog.ActualCost)
	require.Zero(t, log.ActualCost)
	require.Equal(t, 0, userRepo.updateBalanceCalls)
}

func TestUsageServiceCreate_TokenGateCapacityChargesBalance(t *testing.T) {
	usageRepo := &usageServiceLogRepoStub{inserted: true}
	userRepo := &usageServiceUserRepoStub{}
	svc := NewUsageService(usageRepo, userRepo, newUsageServiceEntClient(t), nil)

	_, err := svc.Create(context.Background(), CreateUsageLogRequest{
		UserID:         42,
		APIKeyID:       7,
		AccountID:      9,
		RequestID:      "usage-service-tokengate",
		Model:          "gpt-5.1",
		InputTokens:    100,
		OutputTokens:   20,
		TotalCost:      0.25,
		ActualCost:     0.25,
		RateMultiplier: 1,
		CapacitySource: CapacitySourceTokenGate,
	})

	require.NoError(t, err)
	require.Equal(t, 1, userRepo.updateBalanceCalls)
	require.Equal(t, -0.25, userRepo.lastBalanceAmount)
}
