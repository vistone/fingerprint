#!/usr/bin/env python3
"""
ONNX Runtime inference bridge for Fingerprint ML pipeline.

Reads JSON from stdin:
{
  "samples": [
    {
      "features": [...30],
      "cross_features": [...10],
      "behavior": [...8]
    }
  ]
}

Outputs JSON to stdout:
{
  "results": [
    {
      "embedding": [...32],
      "family_logits": [...7],
      "forgery_prob": 0.12,
      "type_logits": [...4],
      "threat_logits": [...6],
      "action_logits": [...5]
    }
  ]
}
"""

import argparse
import json
import sys
from pathlib import Path

import numpy as np
import onnxruntime as ort


def load_manifest(path: Path):
    with path.open() as f:
        return json.load(f)


def create_sessions(model_dir: Path):
    providers = ort.get_available_providers()
    ordered = []
    if "CUDAExecutionProvider" in providers:
        ordered.append("CUDAExecutionProvider")
    ordered.append("CPUExecutionProvider")

    return {
        "encoder": ort.InferenceSession(str(model_dir / "encoder.onnx"), providers=ordered),
        "classifier": ort.InferenceSession(str(model_dir / "classifier.onnx"), providers=ordered),
        "detector": ort.InferenceSession(str(model_dir / "detector.onnx"), providers=ordered),
        "type_net": ort.InferenceSession(str(model_dir / "type_net.onnx"), providers=ordered),
        "threat_net": ort.InferenceSession(str(model_dir / "threat_net.onnx"), providers=ordered),
        "action_net": ort.InferenceSession(str(model_dir / "action_net.onnx"), providers=ordered),
    }


def run_inference(sessions, sample):
    features = np.asarray(sample["features"], dtype=np.float32).reshape(1, -1)
    cross = np.asarray(sample["cross_features"], dtype=np.float32).reshape(1, -1)
    behavior = np.asarray(sample.get("behavior", []), dtype=np.float32).reshape(1, -1)
    if behavior.shape[1] < 8:
        padded = np.zeros((1, 8), dtype=np.float32)
        padded[:, :behavior.shape[1]] = behavior
        behavior = padded
    elif behavior.shape[1] > 8:
        behavior = behavior[:, :8]

    detector_input = np.concatenate([features, cross], axis=1).astype(np.float32)

    embedding = sessions["encoder"].run(None, {"features": features})[0]
    family_logits = sessions["classifier"].run(None, {"embedding": embedding})[0]
    forgery_prob = sessions["detector"].run(None, {"detector_input": detector_input})[0]
    type_logits = sessions["type_net"].run(None, {"detector_input": detector_input})[0]

    threat_input = np.concatenate(
        [embedding, forgery_prob.reshape(1, 1), type_logits, behavior],
        axis=1,
    ).astype(np.float32)
    threat_logits = sessions["threat_net"].run(None, {"threat_input": threat_input})[0]
    action_logits = sessions["action_net"].run(None, {"threat_input": threat_input})[0]

    return {
        "embedding": embedding.reshape(-1).astype(float).tolist(),
        "family_logits": family_logits.reshape(-1).astype(float).tolist(),
        "forgery_prob": float(forgery_prob.reshape(-1)[0]),
        "type_logits": type_logits.reshape(-1).astype(float).tolist(),
        "threat_logits": threat_logits.reshape(-1).astype(float).tolist(),
        "action_logits": action_logits.reshape(-1).astype(float).tolist(),
    }


def main():
    parser = argparse.ArgumentParser(description="ONNX runtime inference bridge")
    parser.add_argument("--manifest", required=True, help="Path to ONNX manifest.json")
    args = parser.parse_args()

    manifest_path = Path(args.manifest)
    if not manifest_path.exists():
        raise FileNotFoundError(f"manifest not found: {manifest_path}")

    model_dir = manifest_path.parent
    _ = load_manifest(manifest_path)
    sessions = create_sessions(model_dir)

    payload = json.load(sys.stdin)
    samples = payload.get("samples", [])
    results = [run_inference(sessions, sample) for sample in samples]
    json.dump({"results": results}, sys.stdout)


if __name__ == "__main__":
    try:
        main()
    except Exception as err:
        print(json.dumps({"error": str(err)}))
        sys.exit(1)
