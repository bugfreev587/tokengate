//go:build unit

package repository

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountRepositorySetOwnerBYOAccountEntitlementMarksInactiveWithoutChangingSchedulability(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("UPDATE accounts\\s+SET schedulable = FALSE,\\s+extra = jsonb_set\\(\\s*jsonb_set").
		WithArgs(int64(42), service.BYOAccountDisabledReasonKey, service.BYOAccountDisabledReasonSubscriptionInactive, service.BYOAccountOperationalSchedulableKey).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)).AddRow(int64(2)))
	mock.ExpectExec("INSERT INTO scheduler_outbox").
		WithArgs(service.SchedulerOutboxEventAccountBulkChanged, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &accountRepository{sql: db}
	affected, err := repo.SetOwnerBYOAccountEntitlement(context.Background(), 42, false)
	require.NoError(t, err)
	require.Equal(t, int64(2), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositorySetOwnerBYOAccountEntitlementClearsMarkerWithoutChangingSchedulability(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("UPDATE accounts\\s+SET schedulable = COALESCE").
		WithArgs(int64(42), service.BYOAccountDisabledReasonKey, service.BYOAccountDisabledReasonSubscriptionInactive, service.BYOAccountOperationalSchedulableKey).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(3)))
	mock.ExpectExec("INSERT INTO scheduler_outbox").
		WithArgs(service.SchedulerOutboxEventAccountBulkChanged, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &accountRepository{sql: db}
	affected, err := repo.SetOwnerBYOAccountEntitlement(context.Background(), 42, true)
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryReconcileBYOAccountEntitlementsReturnsAffectedOwners(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("WITH byo_account_states AS").
		WithArgs(service.BYOAccountDisabledReasonKey, service.BYOAccountDisabledReasonSubscriptionInactive, service.BYOAccountOperationalSchedulableKey).
		WillReturnRows(sqlmock.NewRows([]string{"owner_user_id"}).AddRow(int64(42)).AddRow(int64(42)).AddRow(int64(77)))
	mock.ExpectExec("INSERT INTO scheduler_outbox").
		WithArgs(service.SchedulerOutboxEventFullRebuild, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &accountRepository{sql: db}
	owners, err := repo.ReconcileBYOAccountEntitlements(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int64{42, 77}, owners)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositorySetSchedulableReturnsEffectiveCompatibilityLockState(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("UPDATE accounts\\s+SET extra = CASE").
		WithArgs(int64(9), true, service.BYOAccountDisabledReasonKey, service.BYOAccountDisabledReasonSubscriptionInactive, service.BYOAccountOperationalSchedulableKey).
		WillReturnRows(sqlmock.NewRows([]string{"schedulable"}).AddRow(false))
	mock.ExpectExec("INSERT INTO scheduler_outbox").
		WithArgs(service.SchedulerOutboxEventAccountChanged, int64(9), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &accountRepository{sql: db}
	require.NoError(t, repo.SetSchedulable(context.Background(), 9, true))
	require.NoError(t, mock.ExpectationsWereMet())
}
