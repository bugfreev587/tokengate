package payment

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestDefaultLoadBalancerSelectInstanceFiltersStripeByProviderEnvironment(t *testing.T) {
	ctx := context.Background()
	client := newLoadBalancerEnvironmentTestClient(t)
	lb := NewDefaultLoadBalancer(client, nil)

	live, err := client.PaymentProviderInstance.Create().
		SetProviderKey(TypeStripe).
		SetName("Stripe live").
		SetEnvironment(ProviderEnvironmentLive).
		SetConfig(`{"secretKey":"sk_live_123","currency":"USD"}`).
		SetSupportedTypes("card,link").
		SetEnabled(true).
		SetSortOrder(1).
		Save(ctx)
	require.NoError(t, err)

	sandbox, err := client.PaymentProviderInstance.Create().
		SetProviderKey(TypeStripe).
		SetName("Stripe sandbox").
		SetEnvironment(ProviderEnvironmentSandbox).
		SetConfig(`{"secretKey":"sk_test_123","currency":"USD"}`).
		SetSupportedTypes("card,link").
		SetEnabled(true).
		SetSortOrder(2).
		Save(ctx)
	require.NoError(t, err)

	selected, err := lb.SelectInstance(
		WithProviderEnvironment(ctx, ProviderEnvironmentSandbox),
		"",
		TypeStripe,
		StrategyRoundRobin,
		12.50,
	)
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(sandbox.ID, 10), selected.InstanceID)
	require.Equal(t, ProviderEnvironmentSandbox, selected.Environment)

	selected, err = lb.SelectInstance(
		WithProviderEnvironment(ctx, ProviderEnvironmentLive),
		"",
		TypeStripe,
		StrategyRoundRobin,
		12.50,
	)
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(live.ID, 10), selected.InstanceID)
	require.Equal(t, ProviderEnvironmentLive, selected.Environment)
}

func newLoadBalancerEnvironmentTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:payment_load_balancer_environment?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}
