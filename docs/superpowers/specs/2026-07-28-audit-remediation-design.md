# Audit Remediation Design

## Goal

Remediate the confirmed audit findings in this repository while following the project's existing patterns, restoring reliable repository-level verification, and tightening the highest-risk external attack surfaces by default.

## Scope

This design covers:

1. Restoring root-module verification so `go test ./...` and `go vet ./...` can run without immediate dependency metadata failures.
2. Protecting sensitive Gateway and admin endpoints with the repository's existing API key middleware model.
3. Blocking SSRF targets by default for public scan and admin-triggered outbound fetch flows.
4. Aligning WAF client IP trust rules with Gateway `TrustedProxies` semantics.
5. Fixing confirmed correctness, lifecycle, and concurrency issues in config, pipeline, generator, profiles, and client transport code.
6. Adding focused regression tests for each repaired behavior.

Out of scope:

- Introducing RBAC, session auth, OAuth, or a new security framework.
- Broad architectural refactors unrelated to the confirmed findings.
- Opportunistic cleanup in modules not touched by the audit findings.

## Constraints

- Follow existing project patterns before introducing new abstractions.
- Prefer targeted fixes over large refactors.
- Use TDD for each behavior change: failing test first, then minimal implementation.
- Keep code comments in English only.
- Default to secure behavior for management auth and outbound URL validation.
- Reuse existing `modules/gateway/auth.go` and `TrustedProxies` concepts instead of inventing new parallel models.

## Current State Summary

The repository is a Go multi-module workspace with a root module plus many internal modules under `modules/`. The highest-risk findings cluster into two groups:

1. External attack surface issues in `modules/gateway` and `modules/waf`
2. Internal correctness and lifecycle issues in `modules/internal`, `modules/generator`, `modules/profiles`, and `modules/client`

Repository-level verification is also currently blocked because root `go test ./...` and `go vet ./...` fail early on missing `go.sum` entries for internal module imports used by several `cmd/*` packages. That must be fixed first so later remediation can be validated with confidence.

## Recommended Approach

Use a staged remediation plan with explicit verification gates between stages:

1. Restore verification entrypoints
2. Tighten external boundaries
3. Align proxy trust behavior
4. Repair internal lifecycle and data isolation bugs
5. Run targeted tests, then broader repository verification

This approach minimizes simultaneous blast radius, keeps failures attributable to the active change set, and makes it possible to stop after any stage with a still-reviewable patch series.

## Design

### 1. Repository Verification Baseline

Objective: make root-level verification commands runnable before changing behavior.

Planned design:

- Repair root dependency metadata so `go test ./...` and `go vet ./...` no longer fail immediately due to missing `go.sum` entries.
- Do not treat this as a general dependency upgrade task.
- Keep the dependency fix minimal and scoped to restoring expected repository commands.

Expected outcome:

- Root-level verification becomes a usable gate for the rest of the remediation work.

### 2. Gateway Authentication Enforcement

Objective: ensure management and other sensitive HTTP entrypoints are not reachable without configured API key authentication.

Planned design:

- Reuse the existing `modules/gateway/auth.go` middleware rather than adding handler-local auth checks.
- Apply middleware at route composition boundaries, not inside each handler.
- Keep health-style endpoints explicitly skippable where appropriate.
- Treat missing API key configuration as deny-by-default for sensitive routes rather than silently exposing them.

Design choices:

- Admin UI and `/api/admin/*` routes become protected as a group.
- Sensitive public API routes are reviewed and protected through the same middleware model where applicable.
- Authentication failure responses should remain machine-readable and consistent with the current middleware style.

Expected outcome:

- Admin configuration updates, test request execution, crawler-triggering, and similar capabilities are no longer anonymously reachable.

### 3. Unified SSRF Target Validation

Objective: enforce one outbound target policy across all user-controlled fetch paths.

Planned design:

- Introduce a shared URL validation path for:
  - Gateway `/scan`
  - Admin `client/test`
  - Admin `crawler/crawl`
- Validate both the submitted URL and any followed redirect targets.
- Reject by default:
  - Empty or malformed URLs
  - Non-HTTP(S) schemes
  - Loopback targets
  - RFC1918 private networks
  - Link-local targets
  - Cloud metadata addresses such as `169.254.169.254`
  - Hostnames that resolve into blocked IP ranges

Design choices:

- Keep policy centralized so the same checks and error semantics apply everywhere.
- Prefer explicit allow/deny logic over scattered inline checks.
- Ensure redirects cannot bypass the initial validation decision.

Expected outcome:

- The service can no longer be used as an unauthenticated or admin-assisted SSRF bridge into local, private, or metadata network targets.

### 4. WAF Trusted Proxy Alignment

Objective: stop WAF decisions from trusting spoofable forwarding headers unless the immediate peer is trusted.

Planned design:

- Align WAF client IP extraction with the Gateway model:
  - Only trust `X-Forwarded-For` / `X-Real-IP` when `RemoteAddr` belongs to a configured trusted proxy.
  - Otherwise use the real socket peer address.
- Apply the same semantics across WAF request analysis and network-engine entrypoints.

Design choices:

- Reuse or mirror the Gateway helper semantics rather than keeping a separate interpretation in WAF.
- Keep the implementation focused on trust rules, not on redesigning rate limiting or session identity.

Expected outcome:

- Attackers connecting directly can no longer spoof their source IP to bypass rate limiting, poison behavior analysis, or force false attribution.

### 5. Client Transport TLS Verification

Objective: restore secure default TLS verification in outbound client paths.

Planned design:

- Remove unconditional `InsecureSkipVerify: true` defaults from the main transport paths.
- Preserve compatibility only through explicit, narrowly-scoped configuration if a legitimate test or legacy path requires it.
- Keep certificate verification behavior closer to real browser semantics.

Expected outcome:

- Invalid, intercepted, or self-signed TLS endpoints fail by default unless an intentional and controlled exception path exists.

### 6. UnifiedConfigManager Lifecycle Safety

Objective: eliminate send/close races and unsafe lifecycle transitions in enhanced config features.

Planned design:

- Protect access to the `enhanced` feature set consistently.
- Replace unsafe channel-closure detection patterns with explicit ownership and lifecycle state handling.
- Ensure `EnableEnhancedFeatures`, `DisableEnhancedFeatures`, `Subscribe`, `Unsubscribe`, and `Update` cannot race into `send on closed channel` or consume real events while "checking" channel state.

Expected outcome:

- Config broadcasting and subscription flows become concurrency-safe and predictable.

### 7. TimeoutMiddleware Cancellation Safety

Objective: prevent timeout handling from leaving uncontrolled background work mutating shared state.

Planned design:

- Preserve timeout behavior but ensure timed-out stage execution cannot keep silently mutating shared `StageData`.
- Favor a design that makes cancellation semantics explicit and testable.
- Keep the fix scoped to the middleware contract rather than redesigning the whole pipeline framework.

Expected outcome:

- Stage timeout becomes a true execution boundary instead of merely an early return to the caller.

### 8. Generator and Profile Isolation Fixes

Objective: eliminate data aliasing and batch-generation liveness bugs.

Planned design:

- Fix `GenerateBatch` so duplicate-source avoidance cannot spin indefinitely when the candidate pool is too small.
- Make generated profiles structurally independent from base profiles.
- Ensure registry reads and clone operations do not leak mutable internal state through shallow copies.

Expected outcome:

- Batch generation terminates deterministically.
- Generated and returned profiles can be safely modified by callers without contaminating templates or global registry state.

### 9. Regression Test Strategy

Each fix is paired with focused tests that fail before the implementation change and pass after it.

Primary regression areas:

- Root verification commands compile and run
- Missing or invalid API keys block sensitive routes
- SSRF validation rejects blocked targets and blocked redirects
- WAF only trusts forwarding headers from trusted proxies
- Invalid TLS certificates fail by default
- Config broadcast lifecycle no longer panics under concurrent enable/disable/update flows
- Timeout handling does not leave uncontrolled background work
- Generator batch logic terminates under small candidate sets
- Profile copies are isolated from shared mutable state

## Implementation Order

1. Root verification baseline
2. Gateway auth enforcement
3. Unified SSRF validation
4. WAF trusted proxy alignment
5. Client TLS verification defaults
6. UnifiedConfigManager lifecycle fixes
7. TimeoutMiddleware lifecycle fixes
8. Generator and profile isolation fixes
9. Broader verification pass

## Testing and Verification

Stage-level verification should prefer focused commands first, then expand:

- Package-level `go test` for the touched modules
- Targeted race-sensitive tests where concurrency behavior changes
- Root `go vet ./...`
- Root `go test ./...`

Where repository-level commands are still too broad for immediate iteration, package-scoped tests should be used during development, but the final acceptance target remains successful root verification or a clearly documented residual blocker discovered after the remediation work begins.

## Risks and Mitigations

### Risk: Breaking existing deployments that relied on anonymous admin access

Mitigation:

- This is an intentional security hardening change.
- Keep the authentication model simple and documented.
- Ensure failure mode is explicit and test-covered.

### Risk: Overblocking legitimate outbound fetch targets

Mitigation:

- Limit the default blocklist to clearly dangerous network ranges and schemes.
- Centralize validation so policy changes happen in one place.
- Add tests for allowed public URLs and blocked internal targets.

### Risk: Internal concurrency fixes become accidental refactors

Mitigation:

- Keep each fix driven by one failing test at a time.
- Avoid redesigning unrelated abstractions while repairing lifecycle bugs.

## Success Criteria

The remediation is complete when:

1. Root verification commands no longer fail immediately from dependency metadata gaps.
2. Sensitive Gateway and admin routes require authentication by default.
3. SSRF-prone fetch paths reject blocked targets and redirects consistently.
4. WAF trusts forwarded client IP headers only from trusted proxies.
5. TLS verification is secure by default in client transport paths.
6. Confirmed concurrency and data-isolation bugs have focused regression tests.
7. Final verification provides evidence that the repaired behaviors hold and did not introduce obvious regressions.
