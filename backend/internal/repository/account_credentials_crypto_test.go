//go:build unit

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type accountCredentialEncryptorStub struct{}

func (accountCredentialEncryptorStub) Encrypt(plaintext string) (string, error) {
	return "encrypted:" + strings.ReplaceAll(plaintext, "secret", "redacted"), nil
}

func (accountCredentialEncryptorStub) Decrypt(ciphertext string) (string, error) {
	return strings.ReplaceAll(strings.TrimPrefix(ciphertext, "encrypted:"), "redacted", "secret"), nil
}

func TestEncryptAccountCredentialsWrapperDoesNotContainToken(t *testing.T) {
	encryptor := accountCredentialEncryptorStub{}

	wrapped, err := encryptAccountCredentials(encryptor, map[string]any{
		"access_token":  "access-secret",
		"refresh_token": "refresh-secret",
	})
	require.NoError(t, err)

	raw, err := json.Marshal(wrapped)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "access-secret")
	require.NotContains(t, string(raw), "refresh-secret")
	require.Equal(t, true, wrapped[accountCredentialsEncryptedKey])
	require.Equal(t, accountCredentialsSchemaV1, wrapped[accountCredentialsSchemaKey])

	plain, err := decryptAccountCredentials(encryptor, wrapped)
	require.NoError(t, err)
	require.Equal(t, "access-secret", plain["access_token"])
	require.Equal(t, "refresh-secret", plain["refresh_token"])
}

func TestAccountRepositoryEncryptsCredentialsAtRestAndDecryptsOnRead(t *testing.T) {
	ctx := context.Background()
	client := newAccountCredentialSQLiteClient(t)
	repo := newAccountRepositoryWithSQL(client, nil, nil, accountCredentialEncryptorStub{})

	account := &service.Account{
		Name:                 "byo-openai",
		Platform:             service.PlatformOpenAI,
		Type:                 service.AccountTypeOAuth,
		Credentials:          map[string]any{"access_token": "access-secret", "refresh_token": "refresh-secret"},
		CredentialsEncrypted: true,
		Concurrency:          1,
		Priority:             50,
		Status:               service.StatusActive,
		Schedulable:          true,
		AutoPauseOnExpired:   true,
	}
	require.NoError(t, repo.Create(ctx, account))

	raw, err := client.Account.Get(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, raw.CredentialsEncrypted)
	require.True(t, isEncryptedAccountCredentials(raw.Credentials))

	rawJSON, err := json.Marshal(raw.Credentials)
	require.NoError(t, err)
	require.NotContains(t, string(rawJSON), "access-secret")
	require.NotContains(t, string(rawJSON), "refresh-secret")

	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, got.CredentialsEncrypted)
	require.Equal(t, "access-secret", got.Credentials["access_token"])
	require.Equal(t, "refresh-secret", got.Credentials["refresh_token"])
}

func TestAccountRepositoryUpdateCredentialsKeepsEncryptedAtRest(t *testing.T) {
	ctx := context.Background()
	client := newAccountCredentialSQLiteClient(t)
	repo := newAccountRepositoryWithSQL(client, nil, nil, accountCredentialEncryptorStub{})

	account := &service.Account{
		Name:                 "byo-openai-refresh",
		Platform:             service.PlatformOpenAI,
		Type:                 service.AccountTypeOAuth,
		Credentials:          map[string]any{"access_token": "old-secret"},
		CredentialsEncrypted: true,
		Concurrency:          1,
		Priority:             50,
		Status:               service.StatusActive,
		Schedulable:          true,
		AutoPauseOnExpired:   true,
	}
	require.NoError(t, repo.Create(ctx, account))
	require.NoError(t, repo.UpdateCredentials(ctx, account.ID, map[string]any{
		"access_token":  "new-secret",
		"refresh_token": "refresh-secret",
	}))

	raw, err := client.Account.Get(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, raw.CredentialsEncrypted)
	require.True(t, isEncryptedAccountCredentials(raw.Credentials))

	rawJSON, err := json.Marshal(raw.Credentials)
	require.NoError(t, err)
	require.NotContains(t, string(rawJSON), "new-secret")
	require.NotContains(t, string(rawJSON), "refresh-secret")

	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, "new-secret", got.Credentials["access_token"])
	require.Equal(t, "refresh-secret", got.Credentials["refresh_token"])
}

func TestAccountRepositoryListAndGetByOwnerUserID(t *testing.T) {
	ctx := context.Background()
	client := newAccountCredentialSQLiteClient(t)
	repo := newAccountRepositoryWithSQL(client, nil, nil, accountCredentialEncryptorStub{})
	ownerID := int64(101)
	otherOwnerID := int64(202)

	owned := &service.Account{
		Name:                 "owned-byo-openai",
		Platform:             service.PlatformOpenAI,
		Type:                 service.AccountTypeOAuth,
		OwnerUserID:          &ownerID,
		Credentials:          map[string]any{"access_token": "owned-secret"},
		CredentialsEncrypted: true,
		Concurrency:          1,
		Priority:             50,
		Status:               service.StatusActive,
		Schedulable:          true,
		AutoPauseOnExpired:   true,
	}
	otherOwned := &service.Account{
		Name:                 "other-byo-openai",
		Platform:             service.PlatformOpenAI,
		Type:                 service.AccountTypeOAuth,
		OwnerUserID:          &otherOwnerID,
		Credentials:          map[string]any{"access_token": "other-secret"},
		CredentialsEncrypted: true,
		Concurrency:          1,
		Priority:             50,
		Status:               service.StatusActive,
		Schedulable:          true,
		AutoPauseOnExpired:   true,
	}
	require.NoError(t, repo.Create(ctx, owned))
	require.NoError(t, repo.Create(ctx, otherOwned))

	list, page, err := repo.ListByOwnerUserID(ctx, ownerID, pagination.PaginationParams{Page: 1, PageSize: 20, SortBy: "id"})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, list, 1)
	require.Equal(t, owned.ID, list[0].ID)
	require.NotNil(t, list[0].OwnerUserID)
	require.Equal(t, ownerID, *list[0].OwnerUserID)
	require.Equal(t, "owned-secret", list[0].Credentials["access_token"])

	got, err := repo.GetByIDAndOwnerUserID(ctx, owned.ID, ownerID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "owned-secret", got.Credentials["access_token"])

	_, err = repo.GetByIDAndOwnerUserID(ctx, otherOwned.ID, ownerID)
	require.ErrorIs(t, err, service.ErrAccountNotFound)
}

func newAccountCredentialSQLiteClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}
