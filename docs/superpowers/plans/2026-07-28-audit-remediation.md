# Audit Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore repository-level verification, lock down the highest-risk external attack surfaces, and fix the confirmed internal correctness and concurrency bugs identified in the audit.

**Architecture:** Execute the remediation in narrow, test-first slices. First restore root verification, then harden gateway and WAF boundaries, then repair internal lifecycle and data-isolation bugs. Each task ends with focused tests before broader verification.

**Tech Stack:** Go 1.25 workspace, standard library `net/http`, package-scoped Go tests, root `go test ./...`, root `go vet ./...`

## Global Constraints

- Follow existing project patterns before introducing new abstractions.
- Prefer targeted fixes over large refactors.
- Use TDD for each behavior change: failing test first, then minimal implementation.
- Keep code comments in English only.
- Default to secure behavior for management auth and outbound URL validation.
- Reuse existing `modules/gateway/auth.go` and `TrustedProxies` concepts instead of inventing new parallel models.

---

### Task 1: Restore Root Verification Baseline

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Test: root verification commands

**Interfaces:**
- Consumes: existing root module dependency graph
- Produces: runnable root `go test ./...` and `go vet ./...` entrypoints, or a reduced and well-defined remaining blocker

- [ ] **Step 1: Write the failing verification expectation**

Document the current failure by running:

```bash
go test ./...
go vet ./...
```

Expected: both commands fail immediately with missing `go.sum` entries for internal module imports used by `cmd/*`.

- [ ] **Step 2: Make the dependency metadata fix with the smallest possible scope**

Update root module dependency metadata so the missing internal-module `go.sum` entries are recorded without turning this into a dependency upgrade task.

- [ ] **Step 3: Re-run root verification**

Run:

```bash
go test ./...
go vet ./...
```

Expected: the prior missing-`go.sum` failures are gone. If new failures appear, record them and continue only if they are genuine code/test issues rather than metadata breakage.

### Task 2: Protect Admin and Sensitive Gateway Routes

**Files:**
- Modify: `modules/gateway/gateway_analyze.go`
- Modify: `modules/gateway/web/handler.go`
- Reuse/Modify if needed: `modules/gateway/auth.go`
- Test: `modules/gateway/auth_test.go` or a new focused route-auth test file near gateway/web

**Interfaces:**
- Consumes: `NewAPIKeyAuth(keys, skipPaths).Middleware(next)`
- Produces: authenticated wrapper for admin and other sensitive HTTP routes

- [ ] **Step 1: Write a failing test for unauthenticated admin access**

Add a test that registers the admin routes behind the gateway HTTP composition and asserts an unauthenticated request to `/admin` or `/api/admin/config` is rejected.

- [ ] **Step 2: Run the focused auth test and verify it fails for the right reason**

Run a package-scoped test command for the new test.

Expected: route is currently reachable or not protected by auth middleware.

- [ ] **Step 3: Apply auth middleware at route composition boundaries**

Protect admin UI and `/api/admin/*` routes as a group using the existing API key middleware model. Keep health-style routes explicit in skip paths where appropriate.

- [ ] **Step 4: Re-run focused gateway auth tests**

Expected: missing key is rejected, valid key still succeeds, skip paths remain reachable.

### Task 3: Add Unified SSRF Target Validation

**Files:**
- Modify: `modules/gateway/gateway_scanner.go`
- Modify: `modules/gateway/gateway_fetch.go`
- Modify: `modules/gateway/web/handler_operations.go`
- Modify: `modules/gateway/web/handler_pages.go` or crawler-related handler file
- Create or Modify: a shared gateway URL validation helper in `modules/gateway/`
- Test: scanner, admin client test, and crawler tests near existing gateway test files

**Interfaces:**
- Consumes: user-controlled URL input for `/scan`, `client/test`, and `crawler/crawl`
- Produces: shared validation function that accepts only safe outbound HTTP(S) targets and rejects blocked redirects

- [ ] **Step 1: Write one failing scanner test for a blocked target**

Add a test asserting `/scan` rejects a blocked target such as `http://127.0.0.1` or `http://169.254.169.254`.

- [ ] **Step 2: Run the test and confirm it fails because the target is currently accepted**

- [ ] **Step 3: Implement shared outbound target validation**

The validator should reject:
- empty or malformed URLs
- non-HTTP(S) schemes
- loopback targets
- RFC1918 private IPs
- link-local addresses
- metadata address `169.254.169.254`
- hostnames resolving into blocked ranges

- [ ] **Step 4: Extend tests to admin `client/test`, crawler entrypoint, and blocked redirects**

- [ ] **Step 5: Re-run the focused gateway SSRF tests**

Expected: blocked targets and blocked redirects fail consistently; a normal public test URL path still passes.

### Task 4: Align WAF Client IP Trust Rules

**Files:**
- Modify: `modules/waf/waf_analyze.go`
- Modify: `modules/waf/engine_network.go`
- Modify if necessary: shared config or helper wiring for trusted proxies
- Test: `modules/waf/waf_test.go` and/or a new focused WAF proxy-trust test file

**Interfaces:**
- Consumes: request `RemoteAddr`, `X-Forwarded-For`, `X-Real-IP`, trusted proxy config
- Produces: consistent client IP extraction that only trusts forwarded headers from trusted proxies

- [ ] **Step 1: Write a failing test for spoofed `X-Forwarded-For` from an untrusted peer**

- [ ] **Step 2: Run the WAF test and verify the current implementation trusts the spoofed header**

- [ ] **Step 3: Implement trusted-proxy-aware client IP extraction**

Use the gateway semantics: trust forwarding headers only when the immediate peer is trusted.

- [ ] **Step 4: Add a positive test for a trusted proxy path**

- [ ] **Step 5: Re-run focused WAF tests**

### Task 5: Restore Secure TLS Verification Defaults

**Files:**
- Modify: `modules/client/transport.go`
- Test: add/update tests in `modules/client/`

**Interfaces:**
- Consumes: outbound TLS transport creation in HTTP/2, HTTP/1.1, and compatibility paths
- Produces: default certificate verification behavior that does not disable TLS verification globally

- [ ] **Step 1: Write a failing test against an invalid/self-signed certificate path**

- [ ] **Step 2: Run the client transport test and verify it currently succeeds when it should fail**

- [ ] **Step 3: Remove unconditional `InsecureSkipVerify: true` defaults**

- [ ] **Step 4: Re-run focused client transport tests**

Expected: invalid certificates fail by default; legitimate paths still work.

### Task 6: Fix UnifiedConfigManager Channel Lifecycle Races

**Files:**
- Modify: `modules/internal/config/unified.go`
- Test: `modules/internal/config/unified_test.go`

**Interfaces:**
- Consumes: `EnableEnhancedFeatures`, `DisableEnhancedFeatures`, `Subscribe`, `Unsubscribe`, `Update`
- Produces: concurrency-safe feature lifecycle and event broadcasting

- [ ] **Step 1: Write a failing concurrency test**

Target a race-prone sequence such as concurrent `Update()` and `DisableEnhancedFeatures()` or repeated enable/disable/subscribe flows that currently panic or drop into unsafe close/send behavior.

- [ ] **Step 2: Run the focused config test and confirm the failure mode**

- [ ] **Step 3: Implement explicit ownership and safe lifecycle transitions**

Remove channel-state probing by receive, and make send/close ordering safe.

- [ ] **Step 4: Re-run focused config tests**

### Task 7: Fix TimeoutMiddleware Cancellation Safety

**Files:**
- Modify: `modules/internal/pipeline/pipeline.go`
- Test: add/update tests near `modules/internal/pipeline/`

**Interfaces:**
- Consumes: `TimeoutMiddleware.Process(ctx, stageName, data, next)`
- Produces: timeout behavior that does not leave uncontrolled work mutating shared state after timeout

- [ ] **Step 1: Write a failing timeout regression test**

Use a stage that ignores cancellation long enough to expose the post-timeout behavior.

- [ ] **Step 2: Run the focused pipeline test and verify the current behavior is unsafe**

- [ ] **Step 3: Implement the smallest contract-safe fix**

- [ ] **Step 4: Re-run focused pipeline tests**

### Task 8: Fix Generator Liveness and Profile Isolation

**Files:**
- Modify: `modules/generator/generator.go`
- Modify: `modules/generator/generator_tools.go`
- Modify: `modules/profiles/profile.go`
- Modify: `modules/profiles/profile_safe.go`
- Test: `modules/generator/generator_test.go`, `modules/profiles/profile_test.go`, and adjacent focused tests

**Interfaces:**
- Consumes: `GenerateBatch`, `mutateProfile`, `ProfileRegistry.Get`, `ProfileRegistry.GetAll`, `(*ClientProfile).Clone`
- Produces: deterministic batch generation and structurally independent returned/generated profiles

- [ ] **Step 1: Write a failing test for duplicate-source batch generation with too-small candidate pool**

- [ ] **Step 2: Run the generator test and confirm it exposes the liveness issue**

- [ ] **Step 3: Implement a deterministic duplicate-handling strategy**

Avoid `i--`-style retry logic that can spin forever when uniqueness is impossible.

- [ ] **Step 4: Write a failing test for mutation isolation**

Assert that editing a generated or retrieved profile does not mutate the template or registry-owned original.

- [ ] **Step 5: Run the isolation tests and confirm the current shallow-copy behavior fails**

- [ ] **Step 6: Implement deep-copy isolation for mutable nested fields actually exposed by the public API**

- [ ] **Step 7: Re-run focused generator and profiles tests**

### Task 9: Final Verification Sweep

**Files:**
- No intended code changes; verification only

**Interfaces:**
- Consumes: all prior task outputs
- Produces: final evidence of behavior correctness and residual blockers, if any remain

- [ ] **Step 1: Run touched-package tests**

Run package-scoped tests for:

```bash
go test ./modules/gateway/... ./modules/waf/... ./modules/client/... ./modules/internal/... ./modules/generator/... ./modules/profiles/...
```

- [ ] **Step 2: Run root static verification**

Run:

```bash
go vet ./...
```

- [ ] **Step 3: Run root test verification**

Run:

```bash
go test ./...
```

- [ ] **Step 4: Summarize remaining blockers, if any**

If anything still fails, classify whether it is:
- a regression introduced by this work
- an unrelated pre-existing failure
- an environment or data dependency
