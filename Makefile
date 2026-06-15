.PHONY: build build-backend build-frontend build-datamanagementd test test-backend test-frontend test-frontend-critical test-public-release-regression test-p0-compatibility test-p0-canary test-p0-canary-selftest test-datamanagementd secret-scan

PNPM ?= pnpm

FRONTEND_CRITICAL_VITEST := \
	src/views/auth/__tests__/LinuxDoCallbackView.spec.ts \
	src/views/auth/__tests__/WechatCallbackView.spec.ts \
	src/views/user/__tests__/PaymentView.spec.ts \
	src/views/user/__tests__/PaymentResultView.spec.ts \
	src/components/user/profile/__tests__/ProfileInfoCard.spec.ts \
	src/views/admin/__tests__/SettingsView.spec.ts \
	src/views/admin/__tests__/LaunchReadinessView.spec.ts \
	src/__tests__/integration/admin-launch-readiness.spec.ts

FRONTEND_PUBLIC_RELEASE_REGRESSION_VITEST := \
	src/views/admin/__tests__/LaunchReadinessView.spec.ts \
	src/__tests__/integration/admin-launch-readiness.spec.ts

# 一键编译前后端
build: build-backend build-frontend

# 编译后端（复用 backend/Makefile）
build-backend:
	@$(MAKE) -C backend build

# 编译前端（需要已安装依赖）
build-frontend:
	@$(PNPM) --dir frontend run build

# 编译 datamanagementd（宿主机数据管理进程）
build-datamanagementd:
	@cd datamanagement && go build -o datamanagementd ./cmd/datamanagementd

# 运行测试（后端 + 前端）
test: test-backend test-frontend

test-backend:
	@$(MAKE) -C backend test

test-frontend:
	@$(PNPM) --dir frontend run lint:check
	@$(PNPM) --dir frontend run typecheck
	@$(MAKE) test-frontend-critical

test-frontend-critical:
	@$(PNPM) --dir frontend exec vitest run $(FRONTEND_CRITICAL_VITEST)

test-public-release-regression:
	@$(MAKE) -C backend test-public-release-regression
	@$(PNPM) --dir frontend exec vitest run $(FRONTEND_PUBLIC_RELEASE_REGRESSION_VITEST)

test-p0-compatibility:
	@tools/tokengate_p0_compatibility_suite.sh

test-p0-canary:
	@tools/tokengate_p0_canary.sh

test-p0-canary-selftest:
	@tools/tokengate_p0_canary_test.sh
	@tools/tokengate_regression_workflow_test.sh

test-datamanagementd:
	@cd datamanagement && go test ./...

secret-scan:
	@python3 tools/secret_scan.py
