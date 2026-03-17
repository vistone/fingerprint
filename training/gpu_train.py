#!/usr/bin/env python3
"""
GPU-accelerated training script for fingerprint ML models.

Called by the Go gateway as a subprocess:
  python3 gpu_train.py --input /tmp/profiles.json --output /models/weights.json [--epochs 200]

Input:  JSON with profile features exported by Go
Output: Go-compatible weights.json for ModelPipeline.LoadWeights()

Progress is written to --progress file as JSON (polled by Go handler).
"""

import argparse
import json
import math
import os
import sys
import time
from pathlib import Path

import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.utils.data import Dataset, DataLoader, TensorDataset

# ── Constants (must match Go ml package exactly) ─────────────────────────
FINGERPRINT_DIM = 30
EMBEDDING_DIM = 32
CROSS_LAYER_DIM = 10
BEHAVIOR_DIM = 8
NUM_BROWSER_FAMILIES = 7
NUM_FORGERY_TYPES = 4
NUM_THREAT_CLASSES = 6
NUM_ACTIONS = 5

DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")


# ── Progress reporting ───────────────────────────────────────────────────
class ProgressReporter:
    def __init__(self, path):
        self.path = path
        self.start = time.time()

    def report(self, phase, epoch, total_epochs, loss, extra=None):
        data = {
            "phase": phase,
            "epoch": epoch,
            "totalEpochs": total_epochs,
            "loss": round(loss, 6),
            "elapsed": round(time.time() - self.start, 1),
            "device": str(DEVICE),
        }
        if extra:
            data.update(extra)
        try:
            with open(self.path, "w") as f:
                json.dump(data, f)
        except OSError:
            pass


# ── Model Definitions (exact match to Go architecture) ───────────────────

class FingerprintEncoder(nn.Module):
    """30 → 32-dim L2-normalized embedding."""
    def __init__(self):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(FINGERPRINT_DIM, 256),
            nn.BatchNorm1d(256),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(256, 128),
            nn.BatchNorm1d(128),
            nn.ReLU(),
            nn.Dropout(0.1),
            nn.Linear(128, 64),
            nn.BatchNorm1d(64),
            nn.ReLU(),
            nn.Linear(64, EMBEDDING_DIM),
        )

    def forward(self, x):
        return F.normalize(self.net(x), p=2, dim=1)


class BrowserClassifier(nn.Module):
    """32-dim embedding → 7-class browser family."""
    def __init__(self):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(EMBEDDING_DIM, 128),
            nn.BatchNorm1d(128),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(128, 64),
            nn.BatchNorm1d(64),
            nn.ReLU(),
            nn.Dropout(0.1),
            nn.Linear(64, NUM_BROWSER_FAMILIES),
        )

    def forward(self, x):
        return self.net(x)


class ForgeryDetectorNet(nn.Module):
    """40-dim → forgery probability (sigmoid)."""
    def __init__(self):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(FINGERPRINT_DIM + CROSS_LAYER_DIM, 128),
            nn.BatchNorm1d(128),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(128, 64),
            nn.BatchNorm1d(64),
            nn.ReLU(),
            nn.Linear(64, 32),
            nn.ReLU(),
            nn.Linear(32, 1),
            nn.Sigmoid(),
        )

    def forward(self, x):
        return self.net(x)


class ForgeryTypeNet(nn.Module):
    """40-dim → 4-class forgery type."""
    def __init__(self):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(FINGERPRINT_DIM + CROSS_LAYER_DIM, 128),
            nn.BatchNorm1d(128),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(128, 64),
            nn.BatchNorm1d(64),
            nn.ReLU(),
            nn.Linear(64, 32),
            nn.ReLU(),
            nn.Linear(32, NUM_FORGERY_TYPES),
        )

    def forward(self, x):
        return self.net(x)


class ThreatNet(nn.Module):
    """45-dim → 6-class threat classification."""
    def __init__(self):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM, 128),
            nn.BatchNorm1d(128),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(128, 64),
            nn.BatchNorm1d(64),
            nn.ReLU(),
            nn.Linear(64, 32),
            nn.ReLU(),
            nn.Linear(32, NUM_THREAT_CLASSES),
        )

    def forward(self, x):
        return self.net(x)


class ActionNet(nn.Module):
    """45-dim → 5-class action recommendation."""
    def __init__(self):
        super().__init__()
        self.net = nn.Sequential(
            nn.Linear(EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM, 128),
            nn.BatchNorm1d(128),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(128, 64),
            nn.BatchNorm1d(64),
            nn.ReLU(),
            nn.Linear(64, 32),
            nn.ReLU(),
            nn.Linear(32, NUM_ACTIONS),
        )

    def forward(self, x):
        return self.net(x)


# ── Feature utilities ────────────────────────────────────────────────────

def compute_cross_layer_features(fp):
    """Compute 10-dim cross-layer features (mirrors Go ComputeCrossLayerFeatures)."""
    batch = fp.shape[0]
    cross = np.zeros((batch, CROSS_LAYER_DIM), dtype=np.float32)
    cross[:, 0] = 1.0 - np.abs(fp[:, 1] - fp[:, 8]) * 2.0
    cross[:, 1] = 1.0 - np.abs(fp[:, 2] - fp[:, 11])
    cross[:, 2] = 1.0 - np.abs(fp[:, 0] - fp[:, 14])
    cross[:, 3] = 1.0 - np.abs(fp[:, 26] - fp[:, 0])
    cross[:, 4] = 1.0 - np.abs(fp[:, 26] - fp[:, 8])
    cross[:, 5] = fp[:, 25]
    both_present = (fp[:, 18] > 0.1) & (fp[:, 19] > 0.1)
    both_absent = (fp[:, 18] < 0.1) & (fp[:, 19] < 0.1)
    cross[:, 6] = np.where(both_present, 1.0, np.where(both_absent, 0.8, 0.2))
    tls13 = fp[:, 0] > 0.8
    cross[:, 7] = np.where(tls13, fp[:, 2], 1.0 - fp[:, 2])
    safe_cipher = np.where(fp[:, 1] > 0, fp[:, 1], 1.0)
    ratio = fp[:, 3] / safe_cipher
    cross[:, 8] = 1.0 - np.abs(ratio - 1.0)
    contradictions = np.sum(cross[:, :9] < 0.3, axis=1)
    cross[:, 9] = contradictions / 9.0
    return np.clip(cross, 0, 1)


def augment_features(features, noise_std):
    noisy = features + np.random.randn(*features.shape).astype(np.float32) * noise_std
    return np.clip(noisy, 0, 1)


# ── Synthetic Profile Expansion ──────────────────────────────────────────

# Browser family feature distribution templates (30-dim ranges)
# [min, max] per dimension for each family
BROWSER_TEMPLATES = {
    0: {  # Chrome
        "tls": ([0.7, 0.9], [0.35, 0.55], [0.3, 0.7], [0.5, 0.75], [1, 1], [1, 1], [0.4, 0.6], [0.05, 0.25]),
        "h2":  ([0.5, 0.7], [0.8, 1.0], [0.5, 0.7], [0.7, 1.0], [0, 0], [0.4, 0.7]),
        "tcp_windows": (128/128, 64240/131072, 1460/2000, 1.0),
        "tcp_mac":     (64/128, 65535/131072, 1460/2000, 1.0),
        "tcp_linux":   (64/128, 64240/131072, 1460/2000, 1.0),
    },
    1: {  # Firefox
        "tls": ([0.7, 0.9], [0.3, 0.5], [0.3, 0.6], [0.4, 0.65], [1, 1], [1, 1], [0.4, 0.6], [0, 0.05]),
        "h2":  ([0.01, 0.02], [0.08, 0.12], [0.5, 0.7], [0.7, 1.0], [0, 0], [0.3, 0.5]),
        "tcp_windows": (128/128, 64240/131072, 1460/2000, 1.0),
        "tcp_mac":     (64/128, 65535/131072, 1460/2000, 1.0),
        "tcp_linux":   (64/128, 64240/131072, 1460/2000, 1.0),
    },
    2: {  # Safari
        "tls": ([0.7, 0.9], [0.3, 0.55], [0.2, 0.5], [0.4, 0.65], [1, 1], [1, 1], [0.3, 0.5], [0, 0.05]),
        "h2":  ([0.5, 0.7], [0.1, 0.3], [0.5, 0.7], [0.5, 0.9], [0, 0.5], [0.3, 0.5]),
        "tcp_mac":     (64/128, 65535/131072, 1460/2000, 1.0),
        "tcp_windows": (128/128, 65535/131072, 1460/2000, 1.0),
        "tcp_linux":   (64/128, 65535/131072, 1460/2000, 1.0),
    },
    3: {  # Edge (Chromium-based)
        "tls": ([0.7, 0.9], [0.35, 0.55], [0.3, 0.65], [0.5, 0.75], [1, 1], [1, 1], [0.4, 0.6], [0.05, 0.25]),
        "h2":  ([0.5, 0.7], [0.8, 1.0], [0.5, 0.7], [0.7, 1.0], [0, 0], [0.4, 0.7]),
        "tcp_windows": (128/128, 64240/131072, 1460/2000, 1.0),
        "tcp_mac":     (64/128, 65535/131072, 1460/2000, 1.0),
        "tcp_linux":   (64/128, 64240/131072, 1460/2000, 1.0),
    },
    4: {  # Opera (Chromium-based)
        "tls": ([0.7, 0.9], [0.35, 0.55], [0.3, 0.6], [0.45, 0.7], [1, 1], [1, 1], [0.4, 0.6], [0.05, 0.2]),
        "h2":  ([0.5, 0.7], [0.8, 1.0], [0.5, 0.7], [0.7, 1.0], [0, 0], [0.4, 0.6]),
        "tcp_windows": (128/128, 64240/131072, 1460/2000, 1.0),
        "tcp_mac":     (64/128, 65535/131072, 1460/2000, 1.0),
        "tcp_linux":   (64/128, 64240/131072, 1460/2000, 1.0),
    },
    5: {  # Brave (Chrome fork, fewer extensions/trackers)
        "tls": ([0.7, 0.9], [0.3, 0.5], [0.3, 0.6], [0.4, 0.65], [1, 1], [1, 1], [0.4, 0.6], [0.05, 0.2]),
        "h2":  ([0.5, 0.7], [0.8, 1.0], [0.5, 0.7], [0.7, 1.0], [0, 0], [0.3, 0.5]),
        "tcp_windows": (128/128, 64240/131072, 1460/2000, 1.0),
        "tcp_mac":     (64/128, 65535/131072, 1460/2000, 1.0),
        "tcp_linux":   (64/128, 64240/131072, 1460/2000, 1.0),
    },
    6: {  # Samsung Internet (Chromium-based, mobile)
        "tls": ([0.7, 0.9], [0.3, 0.5], [0.3, 0.6], [0.45, 0.7], [1, 1], [1, 1], [0.4, 0.6], [0.05, 0.15]),
        "h2":  ([0.4, 0.65], [0.5, 0.9], [0.4, 0.65], [0.5, 0.9], [0, 0], [0.3, 0.5]),
        "tcp_windows": (128/128, 32120/131072, 1460/2000, 1.0),
        "tcp_mac":     (64/128, 32120/131072, 1460/2000, 1.0),
        "tcp_linux":   (64/128, 32120/131072, 1380/2000, 1.0),
    },
}

# JS runtime feature ranges for real browsers (features 18-25)
JS_RUNTIME_PROFILES = {
    "real": {
        18: (0.5, 0.95),   # Canvas hash
        19: (0.3, 0.85),   # WebGL renderer
        20: (0.4, 0.9),    # Audio context
        21: (0.3, 0.8),    # Font enumeration
        22: (0.5, 0.95),   # Plugin count (normalized)
        23: (0.6, 0.95),   # Screen/viewport
        24: (0.0, 0.3),    # Automation detection (low = real)
        25: (0.0, 0.3),    # Consistency anomaly (low = real)
    },
    "headless": {
        18: (0.0, 0.2), 19: (0.0, 0.15), 20: (0.0, 0.1), 21: (0.0, 0.15),
        22: (0.05, 0.15), 23: (0.1, 0.4), 24: (0.7, 1.0), 25: (0.6, 0.9),
    },
    "antidetect": {
        18: (0.3, 0.7), 19: (0.2, 0.6), 20: (0.3, 0.7), 21: (0.2, 0.5),
        22: (0.3, 0.7), 23: (0.4, 0.8), 24: (0.1, 0.4), 25: (0.3, 0.7),
    },
}


def expand_profiles_synthetic(features, labels, target_per_family=800):
    """Expand real profiles synthetically to get more comprehensive coverage.

    For each browser family, generates synthetic variations covering:
    - Different OS platforms (Windows/Mac/Linux TCP/IP characteristics)
    - Version drift (slight TLS/extension variations)
    - JS runtime feature simulation
    - Network condition variations (VPN, mobile, corporate)
    - Regional variations (different header entropy)

    Returns expanded (features, labels) arrays.
    """
    family_indices = {}
    for i, label in enumerate(labels):
        family_indices.setdefault(int(label), []).append(i)

    all_expanded = []
    all_labels = []

    # First, include all originals
    all_expanded.append(features.copy())
    all_labels.append(labels.copy())

    os_types = ["windows", "mac", "linux"]

    for fam, indices in family_indices.items():
        n_real = len(indices)
        n_needed = max(0, target_per_family - n_real)
        if n_needed == 0:
            continue

        template = BROWSER_TEMPLATES.get(fam, BROWSER_TEMPLATES[0])
        real_feats = features[indices]
        syn = np.zeros((n_needed, FINGERPRINT_DIM), dtype=np.float32)

        for i in range(n_needed):
            # Base: sample a real profile from this family
            base = real_feats[i % n_real].copy()

            # 1) OS variation: randomize TCP/IP features (idx 14-17)
            os_choice = np.random.choice(os_types)
            tcp_key = f"tcp_{os_choice}"
            tcp = template.get(tcp_key, template.get("tcp_windows"))
            base[14] = tcp[0] + np.random.randn() * 0.02  # TTL
            base[15] = tcp[1] + np.random.randn() * 0.03  # Window size
            base[16] = tcp[2] + np.random.randn() * 0.01  # MSS
            base[17] = tcp[3]                               # Timestamps

            # 2) Version drift: slight TLS/extension changes
            tls_ranges = template["tls"]
            for dim_idx in range(8):
                lo, hi = tls_ranges[dim_idx]
                # Blend base value with random from range
                alpha = np.random.uniform(0.5, 0.9)
                base[dim_idx] = alpha * base[dim_idx] + (1 - alpha) * np.random.uniform(lo, hi)

            # 3) HTTP/2 variation
            h2_ranges = template["h2"]
            for j, dim_idx in enumerate(range(8, 14)):
                lo, hi = h2_ranges[j]
                alpha = np.random.uniform(0.6, 0.9)
                base[dim_idx] = alpha * base[dim_idx] + (1 - alpha) * np.random.uniform(lo, hi)

            # 4) JS runtime features (idx 18-25) - simulate real browser
            js_prof = JS_RUNTIME_PROFILES["real"]
            for dim_idx in range(18, 26):
                lo, hi = js_prof[dim_idx]
                base[dim_idx] = np.random.uniform(lo, hi)

            # 5) Meta features variation (idx 26-29)
            base[26] = max(0.3, base[26] + np.random.randn() * 0.05)  # UA entropy
            base[27] = max(0.3, base[27] + np.random.randn() * 0.05)  # Profile entropy
            base[28] = np.random.uniform(0.0, 0.15)  # Runtime meta
            base[29] = np.random.uniform(0.0, 0.15)

            # 6) Network condition variation (5% chance VPN/proxy adjustment)
            if np.random.random() < 0.05:
                base[14] += np.random.choice([-0.1, 0.1])  # TTL hop variation
                base[15] *= np.random.uniform(0.7, 1.0)     # Window scaling

            syn[i] = base

        syn = np.clip(syn, 0, 1)
        all_expanded.append(syn)
        all_labels.append(np.full(n_needed, fam, dtype=np.int64))

    # Interpolation between same-family profiles for smoother feature space
    for fam, indices in family_indices.items():
        real_feats = features[indices]
        if len(real_feats) < 2:
            continue
        n_interp = min(500, target_per_family // 3)
        interp = np.zeros((n_interp, FINGERPRINT_DIM), dtype=np.float32)
        for i in range(n_interp):
            idx_a, idx_b = np.random.choice(len(real_feats), 2, replace=False)
            alpha = np.random.beta(2, 2)  # Centered interpolation
            interp[i] = alpha * real_feats[idx_a] + (1 - alpha) * real_feats[idx_b]
            # Add JS features
            js_prof = JS_RUNTIME_PROFILES["real"]
            for dim_idx in range(18, 26):
                lo, hi = js_prof[dim_idx]
                interp[i, dim_idx] = np.random.uniform(lo, hi)
        interp = np.clip(interp, 0, 1)
        all_expanded.append(interp)
        all_labels.append(np.full(n_interp, fam, dtype=np.int64))

    expanded_features = np.concatenate(all_expanded)
    expanded_labels = np.concatenate(all_labels)
    return expanded_features, expanded_labels


def generate_forged_samples(real_features, n):
    """Generate synthetic forged fingerprint samples."""
    forged = np.zeros((n, FINGERPRINT_DIM), dtype=np.float32)
    forgery_types = np.zeros(n, dtype=np.int64)
    for i in range(n):
        strategy = np.random.randint(4)
        s1 = real_features[np.random.randint(len(real_features))]
        s2 = real_features[np.random.randint(len(real_features))]
        if strategy == 0:  # Antidetect: cross-browser mixing
            forged[i, 0:8] = s1[0:8]
            forged[i, 8:14] = s2[8:14]
            forged[i, 14:18] = s1[14:18] if np.random.random() < 0.5 else s2[14:18]
            forged[i, 18] = np.random.random() * 0.3
            forged[i, 19] = np.random.random() * 0.3
            forged[i, 25] = np.random.random() * 0.3 + 0.4
            forged[i, 26] = s1[26]
            forged[i, 27] = s2[27]
            forged[i, 28] = np.random.random() * 0.3 + 0.3
            forgery_types[i] = 2
        elif strategy == 1:  # Headless browser
            forged[i] = s1.copy()
            forged[i, 18:24] = 0
            forged[i, 24] = np.random.random() * 0.3 + 0.7
            forged[i, 25] = np.random.random() * 0.2 + 0.7
            forged[i, 22] = 2.0 / 16.0
            forgery_types[i] = 1
        elif strategy == 2:  # Proxy/MITM
            forged[i] = s1.copy()
            forged[i, 14] += (np.random.random() - 0.5) * 0.3
            forged[i, 15] *= np.random.random() * 0.5 + 0.5
            forged[i, 17] = 0
            forgery_types[i] = 3
        else:  # Generic tool
            alpha = np.random.random()
            forged[i] = alpha * s1 + (1 - alpha) * s2
            forged[i] += np.random.randn(FINGERPRINT_DIM).astype(np.float32) * 0.1
            forgery_types[i] = np.random.choice([1, 2, 3])
    return np.clip(forged, 0, 1), forgery_types


def generate_synthetic_behavior(n, is_forged=False):
    """Generate 8-dim behavior features."""
    behavior = np.zeros((n, BEHAVIOR_DIM), dtype=np.float32)
    if is_forged:
        behavior[:, 0] = np.random.uniform(0.5, 0.8, n)   # switch rate
        behavior[:, 1] = np.random.uniform(0.4, 0.8, n)   # request rate
        behavior[:, 2] = np.random.uniform(0.0, 0.4, n)   # consistency
        behavior[:, 3] = np.random.uniform(0.3, 0.7, n)   # risk trend
        behavior[:, 4] = np.random.uniform(0.1, 0.5, n)   # observations
        behavior[:, 5] = np.random.uniform(0.3, 0.8, n)   # unique FP ratio
        behavior[:, 6] = np.random.uniform(0.0, 0.3, n)   # session duration
        behavior[:, 7] = np.random.uniform(0.4, 0.9, n)   # burst indicator
    else:
        behavior[:, 0] = np.random.uniform(0.0, 0.2, n)
        behavior[:, 1] = np.random.uniform(0.0, 0.3, n)
        behavior[:, 2] = np.random.uniform(0.6, 0.9, n)
        behavior[:, 3] = np.random.uniform(0.0, 0.3, n)
        behavior[:, 4] = np.random.uniform(0.3, 0.8, n)
        behavior[:, 5] = np.random.uniform(0.0, 0.2, n)
        behavior[:, 6] = np.random.uniform(0.3, 0.8, n)
        behavior[:, 7] = np.random.uniform(0.0, 0.2, n)
    behavior += np.random.randn(n, BEHAVIOR_DIM).astype(np.float32) * 0.05
    return np.clip(behavior, 0, 1)


# ── Weight export (Go-compatible JSON) ───────────────────────────────────

def export_sequential_params(model):
    """Extract parameters in Go's SerializedParam order."""
    params = []
    for name, p in model.named_parameters():
        data = p.detach().cpu().numpy()
        if "weight" in name and data.ndim == 2:
            params.append({"shape": list(data.shape), "data": data.flatten().tolist()})
        elif "bias" in name:
            params.append({"shape": [1, data.shape[0]], "data": data.flatten().tolist()})
        elif "running_mean" in name or "running_var" in name:
            continue  # handled below
    # Also export running_mean/var for BatchNorm
    for name, buf in model.named_buffers():
        data = buf.detach().cpu().numpy()
        if "running_mean" in name:
            params.append({"shape": [1, data.shape[0]], "data": data.flatten().tolist()})
        elif "running_var" in name:
            params.append({"shape": [1, data.shape[0]], "data": data.flatten().tolist()})
    return params


def export_sequential_params_ordered(model):
    """Export parameters in exact Go deserialization order:
    For each layer: weight, bias, then for BN: gamma, beta, running_mean, running_var."""
    params = []
    sd = model.state_dict()
    # Go stores: Dense.Weight [out,in], Dense.Bias [1,out], BN.gamma [1,d], BN.beta [1,d], BN.mean [1,d], BN.var [1,d]
    # PyTorch sd has: net.0.weight, net.0.bias, net.1.weight (=gamma), net.1.bias (=beta),
    #                 net.1.running_mean, net.1.running_var, ...
    for key in sd:
        data = sd[key].detach().cpu().numpy()
        if data.ndim == 2:
            params.append({"shape": list(data.shape), "data": data.flatten().tolist()})
        elif data.ndim == 1:
            params.append({"shape": [1, data.shape[0]], "data": data.flatten().tolist()})
        elif data.ndim == 0:
            continue  # num_batches_tracked
    return params


def export_weights(encoder, classifier, detector_net, type_net, threat_net, action_net,
                   metrics_list, output_path):
    """Export all models to Go-compatible ModelWeights JSON."""
    weights = {
        "version": "1.0.17",
        "encoder":     export_sequential_params_ordered(encoder),
        "classifier":  export_sequential_params_ordered(classifier),
        "detector_net": export_sequential_params_ordered(detector_net),
        "type_net":    export_sequential_params_ordered(type_net),
        "threat_net":  export_sequential_params_ordered(threat_net),
        "action_net":  export_sequential_params_ordered(action_net),
        "metrics":     metrics_list,
    }
    with open(output_path, "w") as f:
        json.dump(weights, f)
    size_kb = os.path.getsize(output_path) / 1024
    print(f"  Weights saved: {output_path} ({size_kb:.1f} KB)")


def export_onnx_model(model, dummy_input, output_path, input_name, output_name):
    model.eval()
    output_file = Path(output_path)
    output_file.parent.mkdir(parents=True, exist_ok=True)

    torch.onnx.export(
        model,
        dummy_input,
        str(output_file),
        export_params=True,
        opset_version=17,
        do_constant_folding=True,
        input_names=[input_name],
        output_names=[output_name],
        dynamic_axes={
            input_name: {0: "batch_size"},
            output_name: {0: "batch_size"},
        },
    )


def export_onnx_artifacts(encoder, classifier, detector_net, type_net, threat_net, action_net, onnx_dir):
    onnx_path = Path(onnx_dir)
    onnx_path.mkdir(parents=True, exist_ok=True)

    cpu = torch.device("cpu")
    encoder_cpu = encoder.to(cpu)
    classifier_cpu = classifier.to(cpu)
    detector_cpu = detector_net.to(cpu)
    type_cpu = type_net.to(cpu)
    threat_cpu = threat_net.to(cpu)
    action_cpu = action_net.to(cpu)

    artifacts = [
        ("encoder.onnx", encoder_cpu, torch.randn(1, FINGERPRINT_DIM, device=cpu), "features", "embedding"),
        ("classifier.onnx", classifier_cpu, torch.randn(1, EMBEDDING_DIM, device=cpu), "embedding", "family_logits"),
        ("detector.onnx", detector_cpu, torch.randn(1, FINGERPRINT_DIM + CROSS_LAYER_DIM, device=cpu), "detector_input", "forgery_prob"),
        ("type_net.onnx", type_cpu, torch.randn(1, FINGERPRINT_DIM + CROSS_LAYER_DIM, device=cpu), "detector_input", "type_logits"),
        ("threat_net.onnx", threat_cpu, torch.randn(1, EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM, device=cpu), "threat_input", "threat_logits"),
        ("action_net.onnx", action_cpu, torch.randn(1, EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM, device=cpu), "threat_input", "action_logits"),
    ]

    for file_name, model, dummy, input_name, output_name in artifacts:
        export_onnx_model(
            model=model,
            dummy_input=dummy,
            output_path=onnx_path / file_name,
            input_name=input_name,
            output_name=output_name,
        )

    manifest = {
        "version": "1.0.26",
        "opset": 17,
        "artifacts": [
            {"name": "encoder", "file": "encoder.onnx", "input": [FINGERPRINT_DIM], "output": [EMBEDDING_DIM]},
            {"name": "classifier", "file": "classifier.onnx", "input": [EMBEDDING_DIM], "output": [NUM_BROWSER_FAMILIES]},
            {"name": "detector", "file": "detector.onnx", "input": [FINGERPRINT_DIM + CROSS_LAYER_DIM], "output": [1]},
            {"name": "type_net", "file": "type_net.onnx", "input": [FINGERPRINT_DIM + CROSS_LAYER_DIM], "output": [NUM_FORGERY_TYPES]},
            {"name": "threat_net", "file": "threat_net.onnx", "input": [EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM], "output": [NUM_THREAT_CLASSES]},
            {"name": "action_net", "file": "action_net.onnx", "input": [EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM], "output": [NUM_ACTIONS]},
        ],
    }
    with open(onnx_path / "manifest.json", "w") as f:
        json.dump(manifest, f, indent=2)

    print(f"  ONNX artifacts exported to {onnx_path}")


# ── Training phases ──────────────────────────────────────────────────────

def batch_semihard_triplet_loss(embeddings, labels, margin):
    """Semi-hard triplet mining with L2-normalized embeddings (prevents collapse)."""
    # L2 normalize → all embeddings on unit hypersphere → cannot collapse to zero
    embeddings = F.normalize(embeddings, p=2, dim=1)

    dist_mat = torch.cdist(embeddings, embeddings, p=2)  # [B, B]
    n = embeddings.size(0)
    labels = labels.view(-1)

    same = labels.unsqueeze(0) == labels.unsqueeze(1)  # [B, B]
    diff = ~same
    not_self = ~torch.eye(n, dtype=torch.bool, device=embeddings.device)

    # Hardest positive: max distance among same-class (excl. self)
    ap_dist = dist_mat.clone()
    ap_dist[~(same & not_self)] = -1.0
    hardest_pos, _ = ap_dist.max(dim=1)  # [B]

    # Hardest negative (fallback): closest different-class
    an_dist = dist_mat.clone()
    an_dist[~diff] = float('inf')
    hardest_neg, _ = an_dist.min(dim=1)  # [B]

    # Semi-hard negatives: d(a,n) > d(a,p) AND d(a,n) < d(a,p) + margin
    pos_expanded = hardest_pos.unsqueeze(1)  # [B, 1]
    semi_hard_mask = diff & (dist_mat > pos_expanded) & (dist_mat < pos_expanded + margin)

    sh_dist = dist_mat.clone()
    sh_dist[~semi_hard_mask] = float('inf')
    sh_neg, _ = sh_dist.min(dim=1)  # hardest semi-hard negative

    # Use semi-hard when available, else fall back to hardest neg
    has_sh = semi_hard_mask.sum(dim=1) > 0
    final_neg = torch.where(has_sh, sh_neg, hardest_neg)

    # Valid: have positives and negatives
    has_pos = (same & not_self).sum(dim=1) > 0
    has_neg = diff.sum(dim=1) > 0
    valid = has_pos & has_neg & (hardest_pos > 0)

    if not valid.any():
        return torch.tensor(0.0, device=embeddings.device, requires_grad=True)

    losses = F.relu(hardest_pos[valid] - final_neg[valid] + margin)
    return losses.mean()


def train_encoder(encoder, features, labels, config, progress,
                  classifier=None):
    """Phase 1: End-to-end encoder+classifier joint training with cross-entropy.
    This avoids triplet collapse and provides direct classification signal."""
    print("\n[Phase 1/4] Encoder+Classifier joint training (cross-entropy)")
    encoder.train()
    if classifier is not None:
        classifier.train()

    # Optimize both encoder and classifier together
    params = list(encoder.parameters())
    if classifier is not None:
        params += list(classifier.parameters())
    optimizer = torch.optim.Adam(params, lr=config["lr"])
    scheduler = torch.optim.lr_scheduler.CosineAnnealingWarmRestarts(optimizer, T_0=10, T_mult=2)
    scaler = torch.amp.GradScaler('cuda', enabled=(DEVICE.type == 'cuda'))

    # Moderate augmentation
    all_features = [features]
    all_labels = [labels]
    for noise in [0.01, 0.02, 0.03, 0.04, 0.05, 0.07, 0.09, 0.12]:
        all_features.append(augment_features(features, noise))
        all_labels.append(labels.copy())
    all_features = np.concatenate(all_features)
    all_labels = np.concatenate(all_labels)
    print(f"  Joint training data: {len(all_features):,} samples")

    feat_t = torch.tensor(all_features, device=DEVICE)
    labels_t = torch.tensor(all_labels, device=DEVICE)

    # Train/val split
    n = len(feat_t)
    perm = torch.randperm(n)
    split = int(n * 0.8)
    train_feat, val_feat = feat_t[perm[:split]], feat_t[perm[split:]]
    train_lbl, val_lbl = labels_t[perm[:split]], labels_t[perm[split:]]

    train_ds = TensorDataset(train_feat, train_lbl)
    batch_size = config["batch_size"]
    loader = DataLoader(train_ds, batch_size=batch_size, shuffle=True,
                        drop_last=(len(train_ds) > batch_size),
                        num_workers=0, pin_memory=False)

    epochs = config["epochs"]
    last_loss = 0
    best_acc = 0

    for epoch in range(epochs):
        encoder.train()
        if classifier is not None:
            classifier.train()
        total_loss = 0
        n_batches = 0
        for batch_feat, batch_lbl in loader:
            with torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
                embeddings = encoder(batch_feat)
                if classifier is not None:
                    logits = classifier(embeddings)
                    loss = F.cross_entropy(logits, batch_lbl)
                else:
                    # Fallback: center loss equivalent (shouldn't happen)
                    loss = F.cross_entropy(embeddings, batch_lbl)
            optimizer.zero_grad()
            scaler.scale(loss).backward()
            scaler.step(optimizer)
            scaler.update()
            total_loss += loss.item()
            n_batches += 1
        scheduler.step()
        last_loss = total_loss / max(n_batches, 1)

        # Validation
        encoder.eval()
        if classifier is not None:
            classifier.eval()
        with torch.no_grad(), torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
            val_emb = encoder(val_feat)
            if classifier is not None:
                val_logits = classifier(val_emb)
                val_pred = val_logits.argmax(dim=1)
                acc = (val_pred == val_lbl).float().mean().item()
            else:
                acc = 0.0
        best_acc = max(best_acc, acc)

        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"  Epoch {epoch + 1}/{epochs}  loss={last_loss:.4f}  val_acc={acc:.1%}  batches={n_batches}")
        progress.report("encoder", epoch + 1, epochs, last_loss,
                        {"valAccuracy": round(acc, 4)})

    return last_loss


def train_classifier(classifier, encoder, features, labels, config, progress):
    """Phase 2: Fine-tune classifier with frozen encoder + AMP."""
    print("\n[Phase 2/4] Classifier fine-tuning (encoder frozen)")
    encoder.eval()
    for p in encoder.parameters():
        p.requires_grad = False
    classifier.train()
    optimizer = torch.optim.Adam(classifier.parameters(), lr=config["lr"] * 0.5)
    scaler = torch.amp.GradScaler('cuda', enabled=(DEVICE.type == 'cuda'))

    # Augment and encode
    all_features = [features]
    all_labels = [labels]
    for noise in [0.01, 0.02, 0.03, 0.04, 0.05, 0.07, 0.09, 0.12]:
        all_features.append(augment_features(features, noise))
        all_labels.append(labels.copy())
    all_features = np.concatenate(all_features)
    all_labels = np.concatenate(all_labels)
    print(f"  Classifier data: {len(all_features):,} samples")

    feat_t = torch.tensor(all_features, device=DEVICE)
    labels_t = torch.tensor(all_labels, device=DEVICE)

    with torch.no_grad(), torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
        embeddings = encoder(feat_t)

    # Train/val split
    n = len(embeddings)
    perm = torch.randperm(n)
    split = int(n * 0.8)
    train_emb, val_emb = embeddings[perm[:split]], embeddings[perm[split:]]
    train_lbl, val_lbl = labels_t[perm[:split]], labels_t[perm[split:]]

    train_ds = TensorDataset(train_emb, train_lbl)
    loader = DataLoader(train_ds, batch_size=config["batch_size"], shuffle=True,
                        drop_last=False, num_workers=0, pin_memory=False)

    epochs = config["epochs"]
    best_acc = 0
    last_loss = 0

    for epoch in range(epochs):
        total_loss = 0
        n_batches = 0
        for batch_emb, batch_lbl in loader:
            with torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
                logits = classifier(batch_emb)
                loss = F.cross_entropy(logits, batch_lbl)
            optimizer.zero_grad()
            scaler.scale(loss).backward()
            scaler.step(optimizer)
            scaler.update()
            total_loss += loss.item()
            n_batches += 1
        last_loss = total_loss / max(n_batches, 1)

        # Validation
        classifier.eval()
        with torch.no_grad(), torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
            val_logits = classifier(val_emb)
            val_pred = val_logits.argmax(dim=1)
            acc = (val_pred == val_lbl).float().mean().item()
        classifier.train()
        best_acc = max(best_acc, acc)

        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"  Epoch {epoch + 1}/{epochs}  loss={last_loss:.4f}  val_acc={acc:.1%}  batches={n_batches}")
        progress.report("classifier", epoch + 1, epochs, last_loss,
                        {"valAccuracy": round(acc, 4)})

    return last_loss, best_acc


def train_forgery_detector(detector_net, type_net, features, config, progress):
    """Phase 3: Forgery detection (binary + type classification) + AMP."""
    print("\n[Phase 3/4] Forgery detector")
    n_real = len(features)
    n_forged = int(n_real * config["forgery_ratio"])

    forged_feat, forgery_types = generate_forged_samples(features, n_forged)
    cross_real = compute_cross_layer_features(features)
    cross_forged = compute_cross_layer_features(forged_feat)

    # Combine
    all_feat = np.concatenate([
        np.concatenate([features, cross_real], axis=1),
        np.concatenate([forged_feat, cross_forged], axis=1),
    ])
    all_labels = np.concatenate([np.zeros(n_real), np.ones(n_forged)])
    all_types = np.concatenate([np.zeros(n_real, dtype=np.int64), forgery_types])

    # Augment
    aug_feat = [all_feat]
    aug_labels = [all_labels]
    aug_types = [all_types]
    for noise in [0.01, 0.02, 0.04, 0.06, 0.08, 0.1]:
        aug_feat.append(augment_features(all_feat, noise))
        aug_labels.append(all_labels.copy())
        aug_types.append(all_types.copy())
    all_feat = np.concatenate(aug_feat)
    all_labels = np.concatenate(aug_labels)
    all_types = np.concatenate(aug_types)
    print(f"  Forgery data: {len(all_feat):,} samples")

    feat_t = torch.tensor(all_feat, device=DEVICE)
    label_t = torch.tensor(all_labels, dtype=torch.float32, device=DEVICE)
    type_t = torch.tensor(all_types, device=DEVICE)

    dataset = TensorDataset(feat_t, label_t, type_t)
    loader = DataLoader(dataset, batch_size=config["batch_size"], shuffle=True,
                        drop_last=False, num_workers=0, pin_memory=False)

    detector_net.train()
    type_net.train()
    det_opt = torch.optim.Adam(detector_net.parameters(), lr=config["lr"])
    type_opt = torch.optim.Adam(type_net.parameters(), lr=config["lr"])
    scaler_det = torch.amp.GradScaler('cuda', enabled=(DEVICE.type == 'cuda'))
    scaler_type = torch.amp.GradScaler('cuda', enabled=(DEVICE.type == 'cuda'))

    epochs = config["epochs"]
    last_loss = 0

    for epoch in range(epochs):
        total_loss = 0
        n_batches = 0
        for batch_feat, batch_label, batch_type in loader:
            # Detector — BCE needs float32 (not safe under autocast with sigmoid)
            with torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
                det_out = detector_net(batch_feat).squeeze(-1)
            det_loss = F.binary_cross_entropy(det_out.float(), batch_label.float())
            det_opt.zero_grad()
            scaler_det.scale(det_loss).backward()
            scaler_det.step(det_opt)
            scaler_det.update()

            # Type classifier (only on forged samples)
            forged_mask = batch_label > 0.5
            if forged_mask.any():
                with torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
                    type_out = type_net(batch_feat[forged_mask])
                    type_loss = F.cross_entropy(type_out, batch_type[forged_mask])
                type_opt.zero_grad()
                scaler_type.scale(type_loss).backward()
                scaler_type.step(type_opt)
                scaler_type.update()
                total_loss += (det_loss.item() + type_loss.item())
            else:
                total_loss += det_loss.item()
            n_batches += 1
        last_loss = total_loss / max(n_batches, 1)
        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"  Epoch {epoch + 1}/{epochs}  loss={last_loss:.4f}  batches={n_batches}")
        progress.report("forgery", epoch + 1, epochs, last_loss)

    return last_loss


def train_threat_assessor(threat_net, action_net, encoder, detector_net, type_net,
                          features, labels, config, progress):
    """Phase 4: Threat assessment (threat class + action recommendation) + AMP."""
    print("\n[Phase 4/4] Threat assessor")
    encoder.eval()
    detector_net.eval()
    type_net.eval()

    n_real = len(features)
    n_forged = int(n_real * config["forgery_ratio"])
    forged_feat, _ = generate_forged_samples(features, n_forged)

    all_raw = np.concatenate([features, forged_feat])
    is_forged = np.concatenate([np.zeros(n_real, dtype=bool), np.ones(n_forged, dtype=bool)])

    # Encode through pipeline
    feat_t = torch.tensor(all_raw, device=DEVICE)
    cross = torch.tensor(compute_cross_layer_features(all_raw), device=DEVICE)

    with torch.no_grad(), torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
        embeddings = encoder(feat_t)
        det_input = torch.cat([feat_t, cross], dim=1)
        forg_prob = detector_net(det_input).squeeze(-1)
        forg_type_logits = F.softmax(type_net(det_input), dim=1)

    # Build 45-dim input
    behavior_real = generate_synthetic_behavior(n_real, is_forged=False)
    behavior_forged = generate_synthetic_behavior(n_forged, is_forged=True)
    behavior = np.concatenate([behavior_real, behavior_forged])
    behavior_t = torch.tensor(behavior, device=DEVICE)

    threat_input = torch.cat([embeddings, forg_prob.unsqueeze(1), forg_type_logits, behavior_t], dim=1)

    # Generate labels: real→0 (no threat, allow), forged→random threat/action
    threat_labels = np.zeros(len(all_raw), dtype=np.int64)
    action_labels = np.zeros(len(all_raw), dtype=np.int64)
    for i in range(len(all_raw)):
        if is_forged[i]:
            threat_labels[i] = np.random.choice([1, 2, 3, 4, 5])
            action_labels[i] = np.random.choice([1, 2, 3, 4])
        else:
            threat_labels[i] = 0
            action_labels[i] = 0

    # Augment
    aug_input = [threat_input]
    aug_threat = [threat_labels]
    aug_action = [action_labels]
    for noise in [0.015, 0.025, 0.035, 0.05, 0.065, 0.08, 0.1]:
        noisy = threat_input + torch.randn_like(threat_input) * noise
        aug_input.append(noisy)
        aug_threat.append(threat_labels.copy())
        aug_action.append(action_labels.copy())
    threat_input = torch.cat(aug_input)
    threat_labels = torch.tensor(np.concatenate(aug_threat), device=DEVICE)
    action_labels = torch.tensor(np.concatenate(aug_action), device=DEVICE)
    print(f"  Threat data: {len(threat_input):,} samples")

    dataset = TensorDataset(threat_input, threat_labels, action_labels)
    loader = DataLoader(dataset, batch_size=config["batch_size"], shuffle=True,
                        drop_last=False, num_workers=0, pin_memory=False)

    threat_net.train()
    action_net.train()
    threat_opt = torch.optim.Adam(threat_net.parameters(), lr=config["lr"])
    action_opt = torch.optim.Adam(action_net.parameters(), lr=config["lr"])
    scaler_t = torch.amp.GradScaler('cuda', enabled=(DEVICE.type == 'cuda'))
    scaler_a = torch.amp.GradScaler('cuda', enabled=(DEVICE.type == 'cuda'))

    epochs = config["epochs"]
    last_loss = 0

    for epoch in range(epochs):
        total_loss = 0
        n_batches = 0
        for batch_input, batch_thr, batch_act in loader:
            # Threat
            with torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
                t_out = threat_net(batch_input)
                t_loss = F.cross_entropy(t_out, batch_thr)
            threat_opt.zero_grad()
            scaler_t.scale(t_loss).backward()
            scaler_t.step(threat_opt)
            scaler_t.update()
            # Action
            with torch.amp.autocast('cuda', enabled=(DEVICE.type == 'cuda')):
                a_out = action_net(batch_input)
                a_loss = F.cross_entropy(a_out, batch_act)
            action_opt.zero_grad()
            scaler_a.scale(a_loss).backward()
            scaler_a.step(action_opt)
            scaler_a.update()
            total_loss += (t_loss.item() + a_loss.item())
            n_batches += 1
        last_loss = total_loss / max(n_batches, 1)
        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"  Epoch {epoch + 1}/{epochs}  loss={last_loss:.4f}  batches={n_batches}")
        progress.report("threat", epoch + 1, epochs, last_loss)

    return last_loss


# ── Main ─────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="GPU training for fingerprint ML models")
    parser.add_argument("--input", required=True, help="Path to profile_features.json (exported by Go)")
    parser.add_argument("--output", required=True, help="Path to write weights.json")
    parser.add_argument("--progress", default="/tmp/ml_training_progress.json",
                        help="Path to write progress status JSON")
    parser.add_argument("--epochs", type=int, default=200, help="Training epochs per phase")
    parser.add_argument("--lr", type=float, default=0.001, help="Learning rate")
    parser.add_argument("--batch-size", type=int, default=2048, help="Batch size")
    parser.add_argument("--onnx-dir", default="", help="Optional directory to export ONNX runtime artifacts")
    args = parser.parse_args()

    print("=" * 60)
    print("  Fingerprint ML Training — PyTorch GPU")
    print("=" * 60)
    print(f"Device:  {DEVICE}")
    if torch.cuda.is_available():
        print(f"GPU:     {torch.cuda.get_device_name(0)}")
        print(f"VRAM:    {torch.cuda.get_device_properties(0).total_memory / 1024**3:.1f} GB")
    print(f"Input:   {args.input}")
    print(f"Output:  {args.output}")
    print(f"Epochs:  {args.epochs}")
    print(f"Batch:   {args.batch_size}")

    progress = ProgressReporter(args.progress)
    progress.report("loading", 0, 0, 0)

    # Load profile features exported by Go
    with open(args.input) as f:
        data = json.load(f)
    samples = data["samples"]
    features = np.array([s["features"] for s in samples], dtype=np.float32)
    labels = np.array([s["family_label"] for s in samples], dtype=np.int64)
    browser_types = [s["browser_type"] for s in samples]
    print(f"\nLoaded {len(features)} real profiles, {FINGERPRINT_DIM}-dim features")

    family_counts = {}
    for bt in browser_types:
        family_counts[bt] = family_counts.get(bt, 0) + 1
    for fam, cnt in sorted(family_counts.items()):
        print(f"  {fam:12s} {cnt:3d}")

    # Expand to comprehensive sample library
    print(f"\n--- Synthetic Profile Expansion ---")
    progress.report("expanding", 0, 0, 0)
    expanded_features, expanded_labels = expand_profiles_synthetic(
        features, labels, target_per_family=2000)
    print(f"Expanded: {len(features)} → {len(expanded_features)} samples")
    fam_names = ["chrome", "firefox", "safari", "edge", "opera", "brave", "samsung"]
    for fam_id in range(NUM_BROWSER_FAMILIES):
        n = np.sum(expanded_labels == fam_id)
        print(f"  {fam_names[fam_id]:12s} {n:5d}")

    # Use expanded features for training
    features = expanded_features
    labels = expanded_labels

    config = {
        "epochs": args.epochs,
        "batch_size": args.batch_size,
        "lr": args.lr,
        "triplet_margin": 0.5,
        "forgery_ratio": 1.5,
    }

    # Create models on GPU
    encoder = FingerprintEncoder().to(DEVICE)
    classifier = BrowserClassifier().to(DEVICE)
    detector_net = ForgeryDetectorNet().to(DEVICE)
    type_net = ForgeryTypeNet().to(DEVICE)
    threat_net = ThreatNet().to(DEVICE)
    action_net = ActionNet().to(DEVICE)

    total_params = sum(p.numel() for m in [encoder, classifier, detector_net, type_net, threat_net, action_net]
                       for p in m.parameters())
    print(f"Total parameters: {total_params:,}")
    print(f"AMP (FP16): {'enabled' if DEVICE.type == 'cuda' else 'disabled'}")
    print(f"Training samples (with augmentation): ~{len(features) * 16:,}")

    # torch.compile disabled: container lacks C compiler for Triton backend
    # and benefit is negligible for 131K-param models

    start_time = time.time()
    metrics = []

    # Phase 1: Joint encoder+classifier training (end-to-end)
    enc_loss = train_encoder(encoder, features, labels, config, progress,
                             classifier=classifier)
    metrics.append({"epoch": config["epochs"], "encoder_loss": round(enc_loss, 6)})

    # Phase 2: Classifier
    cls_loss, val_acc = train_classifier(classifier, encoder, features, labels, config, progress)
    metrics.append({"epoch": config["epochs"], "class_loss": round(cls_loss, 6),
                    "val_accuracy": round(val_acc, 4)})

    # Phase 3: Forgery detector
    forg_loss = train_forgery_detector(detector_net, type_net, features, config, progress)
    metrics.append({"epoch": config["epochs"], "forgery_loss": round(forg_loss, 6)})

    # Phase 4: Threat assessor
    threat_loss = train_threat_assessor(threat_net, action_net, encoder, detector_net, type_net,
                                        features, labels, config, progress)
    metrics.append({"epoch": config["epochs"], "threat_loss": round(threat_loss, 6)})

    elapsed = time.time() - start_time
    print(f"\n{'=' * 60}")
    print(f"  Training completed in {elapsed:.1f}s")
    print(f"{'=' * 60}")

    # Export Go-compatible weights
    os.makedirs(os.path.dirname(args.output) or ".", exist_ok=True)
    export_weights(encoder, classifier, detector_net, type_net, threat_net, action_net,
                   metrics, args.output)

    if args.onnx_dir:
        export_onnx_artifacts(encoder, classifier, detector_net, type_net, threat_net, action_net, args.onnx_dir)

    # Final progress
    progress.report("done", config["epochs"], config["epochs"], 0, {
        "totalElapsed": round(elapsed, 1),
        "valAccuracy": round(val_acc, 4),
        "encoderLoss": round(enc_loss, 6),
        "classLoss": round(cls_loss, 6),
        "forgeryLoss": round(forg_loss, 6),
        "threatLoss": round(threat_loss, 6),
    })

    # Quick validation
    print(f"\nValidation:")
    encoder.eval()
    classifier.eval()
    detector_net.eval()
    feat_t = torch.tensor(features[:5], device=DEVICE)
    with torch.no_grad():
        emb = encoder(feat_t)
        cls_out = F.softmax(classifier(emb), dim=1)
        cross = torch.tensor(compute_cross_layer_features(features[:5]), device=DEVICE)
        det_in = torch.cat([feat_t, cross], dim=1)
        forg_prob = detector_net(det_in).squeeze(-1)

    families = ["chrome", "firefox", "safari", "edge", "opera", "brave", "samsung"]
    for i in range(min(5, len(features))):
        pred = families[cls_out[i].argmax().item()]
        conf = cls_out[i].max().item()
        fp = forg_prob[i].item()
        print(f"  {browser_types[i]:12s} → {pred:8s} ({conf:.1%}) forgery: {fp:.1%}")

    print("\nDone.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
