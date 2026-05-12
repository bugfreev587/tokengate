package setup

import (
	"os"
	"strings"
	"testing"
)

func TestDecideAdminBootstrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		totalUsers int64
		adminUsers int64
		should     bool
		reason     string
	}{
		{
			name:       "empty database should create admin",
			totalUsers: 0,
			adminUsers: 0,
			should:     true,
			reason:     adminBootstrapReasonEmptyDatabase,
		},
		{
			name:       "admin exists should skip",
			totalUsers: 10,
			adminUsers: 1,
			should:     false,
			reason:     adminBootstrapReasonAdminExists,
		},
		{
			name:       "users exist without admin should skip",
			totalUsers: 5,
			adminUsers: 0,
			should:     false,
			reason:     adminBootstrapReasonUsersExistWithoutAdmin,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := decideAdminBootstrap(tc.totalUsers, tc.adminUsers)
			if got.shouldCreate != tc.should {
				t.Fatalf("shouldCreate=%v, want %v", got.shouldCreate, tc.should)
			}
			if got.reason != tc.reason {
				t.Fatalf("reason=%q, want %q", got.reason, tc.reason)
			}
		})
	}
}

func TestSetupDefaultAdminConcurrency(t *testing.T) {
	t.Run("simple mode admin uses higher concurrency", func(t *testing.T) {
		t.Setenv("RUN_MODE", "simple")
		if got := setupDefaultAdminConcurrency(); got != simpleModeAdminConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, simpleModeAdminConcurrency)
		}
	})

	t.Run("standard mode keeps existing default", func(t *testing.T) {
		t.Setenv("RUN_MODE", "standard")
		if got := setupDefaultAdminConcurrency(); got != defaultUserConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, defaultUserConcurrency)
		}
	})
}

func TestWriteConfigFileKeepsDefaultUserConcurrency(t *testing.T) {
	t.Setenv("RUN_MODE", "simple")
	t.Setenv("DATA_DIR", t.TempDir())

	if err := writeConfigFile(&SetupConfig{}); err != nil {
		t.Fatalf("writeConfigFile() error = %v", err)
	}

	data, err := os.ReadFile(GetConfigFilePath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), "user_concurrency: 5") {
		t.Fatalf("config missing default user concurrency, got:\n%s", string(data))
	}
}

func TestApplyDatabaseURLFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://railway:secret@db.railway.internal:6543/tokengate?sslmode=require")

	cfg := DatabaseConfig{}
	if err := applyDatabaseURLFromEnv(&cfg); err != nil {
		t.Fatalf("applyDatabaseURLFromEnv() error = %v", err)
	}

	if cfg.Host != "db.railway.internal" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "db.railway.internal")
	}
	if cfg.Port != 6543 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 6543)
	}
	if cfg.User != "railway" {
		t.Fatalf("User = %q, want %q", cfg.User, "railway")
	}
	if cfg.Password != "secret" {
		t.Fatalf("Password = %q, want %q", cfg.Password, "secret")
	}
	if cfg.DBName != "tokengate" {
		t.Fatalf("DBName = %q, want %q", cfg.DBName, "tokengate")
	}
	if cfg.SSLMode != "require" {
		t.Fatalf("SSLMode = %q, want %q", cfg.SSLMode, "require")
	}
}

func TestApplyRedisURLFromEnv(t *testing.T) {
	t.Setenv("REDIS_URL", "rediss://:secret@redis.railway.internal:6380/5")

	cfg := RedisConfig{}
	if err := applyRedisURLFromEnv(&cfg); err != nil {
		t.Fatalf("applyRedisURLFromEnv() error = %v", err)
	}

	if cfg.Host != "redis.railway.internal" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "redis.railway.internal")
	}
	if cfg.Port != 6380 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 6380)
	}
	if cfg.Password != "secret" {
		t.Fatalf("Password = %q, want %q", cfg.Password, "secret")
	}
	if cfg.DB != 5 {
		t.Fatalf("DB = %d, want %d", cfg.DB, 5)
	}
	if !cfg.EnableTLS {
		t.Fatalf("EnableTLS = false, want true")
	}
}
