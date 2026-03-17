# ONNX Runtime Migration Plan

This document defines a safe, phased migration path from the custom Go neural network runtime to ONNX Runtime.

## Why Migrate

- Current handcrafted neural network inference is CPU-only and not optimized.
- Training and inference stacks are split (PyTorch training + custom Go inference), increasing maintenance complexity.
- ONNX Runtime provides a standardized graph runtime, better operator coverage, and optional GPU acceleration.

## Current State (v1.0.26)

- Training script `training/gpu_train.py` now exports:
  - Go-compatible `weights.json` (existing path)
  - Optional ONNX artifacts in a dedicated directory (`--onnx-dir`)
- Exported ONNX files:
  - `encoder.onnx`
  - `classifier.onnx`
  - `detector.onnx`
  - `type_net.onnx`
  - `threat_net.onnx`
  - `action_net.onnx`
  - `manifest.json`
- Gateway async training result includes ONNX artifact summary when files are present.
- ML service supports backend switch:
  - `native` (default)
  - `onnx` (feature flag, with automatic native fallback on runtime errors)

## Runtime Flags

- `FP_ML_INFERENCE_BACKEND=native|onnx`
- `FP_ML_ONNX_MODEL_DIR=/models/onnx`
- `FP_ML_ONNX_PYTHON_BIN=python3`
- `FP_ML_ONNX_PYTHON_SCRIPT=training/onnx_infer.py`
- `FP_ML_ONNX_TIMEOUT_MS=10000`
- `FP_ML_SHADOW_COMPARE_ENABLED=true`
- `FP_ML_SHADOW_SAMPLE_RATE=0.1`
- `FP_ML_SHADOW_METRICS_PATH=/models/shadow_compare.jsonl`
- `FP_ML_CANARY_ENABLED=true`
- `FP_ML_CANARY_RATE=0.05`
- `FP_ML_CANARY_BACKEND=onnx`

## Quick Monitoring

Use the bundled script to watch canary/shadow metrics from the ML stats endpoint:

```bash
INTERVAL_SEC=10 MAX_CHECKS=30 ./scripts/monitor_ml_canary.sh http://localhost:8080
```

## Phased Rollout

1. Phase 1: Dual artifacts (completed)
- Keep current Go inference path as production default.
- Produce ONNX artifacts in every GPU training run.
- Validate shape compatibility and output ranges offline.

2. Phase 2: Shadow inference
- Add ONNX Runtime inference backend behind a feature flag.
- Run native and ONNX inference in parallel for sampled traffic.
- Compare prediction parity:
  - Browser family top-1 agreement
  - Forgery probability delta
  - Threat/action top-1 agreement

3. Phase 3: Gradual traffic switch
- Enable ONNX backend for a small canary percentage.
- Track latency, error rate, and prediction drift.
- Increase rollout only when parity and stability thresholds are met.

4. Phase 4: Default switch
- Make ONNX Runtime the default inference backend.
- Keep native backend as fallback for one release cycle.

## Acceptance Criteria

- P95 inference latency improvement is measurable.
- Prediction parity meets SLO thresholds on validation and shadow traffic.
- No regression in gateway error rate or security decision quality.

## Operational Notes

- Prefer CPUExecutionProvider first for compatibility.
- Enable CUDA/TensorRT providers only in environments with verified runtime libraries.
- Keep `weights.json` export until ONNX path is fully stable and rollback is not required.
