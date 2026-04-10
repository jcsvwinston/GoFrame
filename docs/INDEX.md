# Documentation Map

Reference date: 2026-04-10.
Status: Current.

This file is the canonical entrypoint for GoFrame documentation.

## Start Here

- [QUICKSTART.md](QUICKSTART.md)
- [DEVELOPER_MANUAL.md](DEVELOPER_MANUAL.md)
- [PROJECT_LAYOUT.md](PROJECT_LAYOUT.md)
- [../SPEC.md](../SPEC.md)

## Feature Guides (New)

- [AUTH_GUIDE.md](AUTH_GUIDE.md) — Authentication & Authorization (JWT, sessions, Casbin)
- [TESTING_GUIDE.md](TESTING_GUIDE.md) — Testing strategies and patterns
- [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md) — Production deployment (Docker, K8s, reverse proxy)
- [ERROR_HANDLING.md](ERROR_HANDLING.md) — Error types and HTTP mapping
- [VALIDATION_GUIDE.md](VALIDATION_GUIDE.md) — Input validation and custom rules
- [SIGNALS_GUIDE.md](SIGNALS_GUIDE.md) — Event bus and model hooks
- [MULTISITE_GUIDE.md](MULTISITE_GUIDE.md) — MultiSite & MultiTenant routing
- [RATE_LIMITING_GUIDE.md](RATE_LIMITING_GUIDE.md) — Rate limiting configuration

## Core Engineering References

- [MODELING_MULTI_DATABASE.md](MODELING_MULTI_DATABASE.md)
- [ADMIN_LIVE_RUNTIME_INSPECTOR_SPEC.md](ADMIN_LIVE_RUNTIME_INSPECTOR_SPEC.md)
- [ADMIN_CLUSTER_LAB.md](ADMIN_CLUSTER_LAB.md)
- [CLI_BEST_PRACTICES.md](CLI_BEST_PRACTICES.md)
- [API_CONTRACT_INVENTORY.md](API_CONTRACT_INVENTORY.md)
- [CLI_CONTRACT_MATRIX.md](CLI_CONTRACT_MATRIX.md)
- [CONFIG_KEY_REGISTRY.md](CONFIG_KEY_REGISTRY.md)
- [PLUGIN_SDK.md](PLUGIN_SDK.md)
- [PLUGIN_EXAMPLES.md](PLUGIN_EXAMPLES.md)
- [MAIL_PROVIDERS.md](MAIL_PROVIDERS.md)
- [OBSERVABILITY_BASELINE.md](OBSERVABILITY_BASELINE.md)

## Architecture Decision Records

- [adrs/README.md](adrs/README.md) — ADR-001: stdlib-first, ADR-002: Django-inspired CLI

## Strategy and Governance

- [ENTERPRISE_LONG_TERM_ROADMAP.md](ENTERPRISE_LONG_TERM_ROADMAP.md)
- [ROADMAP_SUPERAR_DJANGO.md](ROADMAP_SUPERAR_DJANGO.md)
- [COMPATIBILITY_SLO.md](COMPATIBILITY_SLO.md)
- [VERSIONING.md](VERSIONING.md)
- [DEPRECATION_TEMPLATE.md](DEPRECATION_TEMPLATE.md)
- [MIGRATION_ASSISTANT_CONVENTIONS.md](MIGRATION_ASSISTANT_CONVENTIONS.md)
- [GO_VERSION_POLICY.md](GO_VERSION_POLICY.md)
- [CI_MATRIX.md](CI_MATRIX.md)
- [RELEASE_CHECKLIST.md](RELEASE_CHECKLIST.md)
- [../CHANGELOG.md](../CHANGELOG.md)

## Validation Reports

- [reports/exploratory_stability.md](reports/exploratory_stability.md)
- [reports/compatibility_harness_latest.md](reports/compatibility_harness_latest.md)
- [reports/dependency_critical_review_2026-04-07.md](reports/dependency_critical_review_2026-04-07.md)
- [reports/release_readiness_2026-04-07.md](reports/release_readiness_2026-04-07.md)

## Precedence Rule

When documents conflict, use this precedence:

1. `README.md`
2. strategy/governance docs listed above
3. detailed implementation docs
4. historical behavior only from git history (not separate phase files)

## Terminology

- External provider binaries: `goframe-plugin-<provider>`
- Legacy mail fallback naming: `goframe-mail-<driver>`
