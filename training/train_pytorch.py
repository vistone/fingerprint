#!/usr/bin/env python3
"""
PyTorch GPU training for fingerprint analysis models.

Trains 4 models matching the Go inference architecture exactly:
  1. FingerprintEncoder: 30→256→BN→ReLU→Drop→128→BN→ReLU→Drop→64→BN→ReLU→32 + L2 norm
  2. BrowserClassifier:  32→128→BN→ReLU→Drop→64→BN→ReLU→Drop→7 (softmax)
  3. ForgeryDetector:    40→128→BN→ReLU→Drop→64→BN→ReLU→32→ReLU→1(sigmoid) / 4(softmax)
  4. ThreatAssessor:     45→128→BN→ReLU→Drop→64→BN→ReLU→32→ReLU→6(softmax) / 5(softmax)

Exports weights in Go-compatible JSON format.
"""

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
from torch.utils.data import DataLoader, Dataset, WeightedRandomSampler

# ── Constants (match Go) ─────────────────────────────────────────────────
FINGERPRINT_DIM = 30
EMBEDDING_DIM = 32
CROSS_LAYER_DIM = 10
BEHAVIOR_DIM = 8
NUM_BROWSER_FAMILIES = 7
NUM_FORGERY_TYPES = 4
NUM_THREAT_CLASSES = 6
NUM_ACTIONS = 5

DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")

# ── Model Definitions ────────────────────────────────────────────────────

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
        raw = self.net(x)
        return F.normalize(raw, p=2, dim=1)


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
        return self.net(x)  # logits (softmax applied in loss)


class ForgeryDetectorNet(nn.Module):
    """40-dim → forgery prob (1) sigmoid."""
    def __init__(self):
        super().__init__()
        input_dim = FINGERPRINT_DIM + CROSS_LAYER_DIM  # 40
        self.net = nn.Sequential(
            nn.Linear(input_dim, 128),
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
        input_dim = FINGERPRINT_DIM + CROSS_LAYER_DIM  # 40
        self.net = nn.Sequential(
            nn.Linear(input_dim, 128),
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
        return self.net(x)  # logits


class ThreatNet(nn.Module):
    """45-dim → 6-class threat."""
    def __init__(self):
        super().__init__()
        input_dim = EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM  # 45
        self.net = nn.Sequential(
            nn.Linear(input_dim, 128),
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
    """45-dim → 5-class action."""
    def __init__(self):
        super().__init__()
        input_dim = EMBEDDING_DIM + 1 + NUM_FORGERY_TYPES + BEHAVIOR_DIM  # 45
        self.net = nn.Sequential(
            nn.Linear(input_dim, 128),
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


# ── Data Utilities ───────────────────────────────────────────────────────

def load_profiles(path: str):
    """Load exported profile features JSON."""
    with open(path) as f:
        data = json.load(f)
    samples = data["samples"]
    features = np.array([s["features"] for s in samples], dtype=np.float32)
    labels = np.array([s["family_label"] for s in samples], dtype=np.int64)
    browser_types = [s["browser_type"] for s in samples]
    return features, labels, browser_types


def compute_cross_layer_features(fp: np.ndarray) -> np.ndarray:
    """Compute 10-dim cross-layer features (mirrors Go ComputeCrossLayerFeatures)."""
    batch = fp.shape[0]
    cross = np.zeros((batch, CROSS_LAYER_DIM), dtype=np.float32)

    cross[:, 0] = 1.0 - np.abs(fp[:, 1] - fp[:, 8]) * 2.0
    cross[:, 1] = 1.0 - np.abs(fp[:, 2] - fp[:, 11])
    cross[:, 2] = 1.0 - np.abs(fp[:, 0] - fp[:, 14])
    cross[:, 3] = 1.0 - np.abs(fp[:, 26] - fp[:, 0])
    cross[:, 4] = 1.0 - np.abs(fp[:, 26] - fp[:, 8])
    cross[:, 5] = fp[:, 25]

    # Canvas-WebGL consistency
    both_present = (fp[:, 18] > 0.1) & (fp[:, 19] > 0.1)
    both_absent = (fp[:, 18] < 0.1) & (fp[:, 19] < 0.1)
    cross[:, 6] = np.where(both_present, 1.0, np.where(both_absent, 0.8, 0.2))

    # Cipher order anomaly
    tls13 = fp[:, 0] > 0.8
    cross[:, 7] = np.where(tls13, fp[:, 2], 1.0 - fp[:, 2])

    # Extension anomaly
    safe_cipher = np.where(fp[:, 1] > 0, fp[:, 1], 1.0)
    ratio = fp[:, 3] / safe_cipher
    cross[:, 8] = 1.0 - np.abs(ratio - 1.0)

    # Contradiction count
    contradictions = np.sum(cross[:, :9] < 0.3, axis=1)
    cross[:, 9] = contradictions / 9.0

    return np.clip(cross, 0, 1)


def augment_features(features: np.ndarray, noise_std: float) -> np.ndarray:
    """Add Gaussian noise for data augmentation."""
    noisy = features + np.random.randn(*features.shape).astype(np.float32) * noise_std
    return np.clip(noisy, 0, 1)


def generate_forged_samples(real_features: np.ndarray, n: int) -> tuple:
    """Generate forged fingerprint samples with multiple strategies."""
    forged = np.zeros((n, FINGERPRINT_DIM), dtype=np.float32)
    forgery_types = np.zeros(n, dtype=np.int64)  # 0=real,1=headless,2=antidetect,3=proxy

    for i in range(n):
        strategy = np.random.randint(4)
        s1 = real_features[np.random.randint(len(real_features))]
        s2 = real_features[np.random.randint(len(real_features))]

        if strategy == 0:  # Cross-browser mixing (antidetect)
            forged[i, 0:8] = s1[0:8]
            forged[i, 8:14] = s2[8:14]
            forged[i, 14:18] = s1[14:18] if np.random.random() < 0.5 else s2[14:18]
            forged[i, 18] = np.random.random() * 0.3
            forged[i, 19] = np.random.random() * 0.3
            forged[i, 25] = np.random.random() * 0.3 + 0.4
            forged[i, 26] = s1[26]
            forged[i, 27] = s2[27]
            forged[i, 28] = np.random.random() * 0.3 + 0.3
            forgery_types[i] = 2  # antidetect

        elif strategy == 1:  # Headless browser
            forged[i] = s1.copy()
            forged[i, 18:24] = 0  # no JS features
            forged[i, 21] = np.random.random() * 0.05
            forged[i, 22] = np.random.random() * 0.2
            forged[i, 24] = 0.25
            forged[i, 25] = np.random.random() * 0.3 + 0.7
            forged[i, 28] = np.random.random() * 0.2 + 0.5
            forgery_types[i] = 1  # headless

        elif strategy == 2:  # Proxy/MITM
            forged[i] = s1.copy()
            forged[i, 14] = 0.5 + np.random.random() * 0.5
            forged[i, 15] = np.random.random() * 0.3
            forged[i, 16] = np.random.random() * 0.4
            forged[i, 17] = 0
            forged[i, 29] = np.random.random() * 0.2 + 0.3
            forgery_types[i] = 3  # proxy

        else:  # Generic noise tool
            mask = np.random.random(FINGERPRINT_DIM) < 0.5
            forged[i] = np.where(mask, s1, s2)
            forged[i] += np.random.randn(FINGERPRINT_DIM).astype(np.float32) * 0.1
            forged[i, 28] = np.random.random() * 0.3 + 0.2
            forgery_types[i] = np.random.choice([1, 2, 3])

        # Final noise
        forged[i] += np.random.randn(FINGERPRINT_DIM).astype(np.float32) * 0.02

    forged = np.clip(forged, 0, 1)
    return forged, forgery_types


def generate_synthetic_behavior(n: int, is_forged: np.ndarray) -> np.ndarray:
    """Generate synthetic behavioral features matching Go generateSyntheticBehavior."""
    behavior = np.zeros((n, BEHAVIOR_DIM), dtype=np.float32)
    for i in range(n):
        if is_forged[i]:
            behavior[i, 0] = np.random.random() * 0.3 + 0.5
            behavior[i, 1] = np.random.random() * 0.4 + 0.4
            behavior[i, 2] = np.random.random() * 0.4
            behavior[i, 3] = np.random.random() * 0.3 + 0.5
            behavior[i, 4] = np.random.random() * 0.3
            behavior[i, 5] = np.random.random() * 0.3 + 0.5
            behavior[i, 6] = np.random.random() * 0.3
            behavior[i, 7] = np.random.random() * 0.3 + 0.3
        else:
            behavior[i, 0] = np.random.random() * 0.2
            behavior[i, 1] = np.random.random() * 0.3
            behavior[i, 2] = np.random.random() * 0.3 + 0.6
            behavior[i, 3] = np.random.random() * 0.3
            behavior[i, 4] = np.random.random() * 0.3 + 0.4
            behavior[i, 5] = np.random.random() * 0.3
            behavior[i, 6] = np.random.random() * 0.3 + 0.5
            behavior[i, 7] = np.random.random() * 0.2
    behavior += np.random.randn(n, BEHAVIOR_DIM).astype(np.float32) * 0.05
    return np.clip(behavior, 0, 1)


def generate_threat_labels(features, forgery_probs, forgery_types):
    """Generate rule-based threat labels matching Go generateThreatLabel."""
    n = len(features)
    threat_labels = np.zeros(n, dtype=np.int64)
    action_labels = np.zeros(n, dtype=np.int64)

    for i in range(n):
        if forgery_probs[i] > 0.7:
            ft = forgery_types[i]
            if ft == 1:  # headless
                threat_labels[i] = 1  # bot
                action_labels[i] = 4  # block
            elif ft == 2:  # antidetect
                threat_labels[i] = 2  # fingerprint_spoof
                action_labels[i] = 3  # throttle
            elif ft == 3:  # proxy
                threat_labels[i] = 5  # evasion
                action_labels[i] = 2  # challenge
            else:
                threat_labels[i] = 2
                action_labels[i] = 3
        elif features[i, 28] > 0.3:  # tool marker
            threat_labels[i] = 1  # bot
            action_labels[i] = 2  # challenge
        elif features[i, 29] > 0.3:  # behavior anomaly
            threat_labels[i] = 4  # behavioral_anomaly
            action_labels[i] = 1  # monitor
        else:
            threat_labels[i] = 0  # none
            action_labels[i] = 0  # allow

    return threat_labels, action_labels


# ── Export to Go format ──────────────────────────────────────────────────

def export_sequential_params(model: nn.Module) -> list:
    """
    Export nn.Sequential parameters in the exact order Go expects:
    For each layer:
      - Linear: weight [out×in], bias [1×out]
      - BatchNorm1d: gamma [1×dim], beta [1×dim], running_mean [1×dim], running_var [1×dim]
    """
    params = []
    for layer in model.net:
        if isinstance(layer, nn.Linear):
            # Go DenseLayer stores Weight as [outDim × inDim], Bias as [1 × outDim]
            w = layer.weight.detach().cpu().numpy()  # [out, in]
            b = layer.bias.detach().cpu().numpy()     # [out]
            params.append({"shape": list(w.shape), "data": w.flatten().tolist()})
            params.append({"shape": [1, len(b)], "data": b.tolist()})
        elif isinstance(layer, nn.BatchNorm1d):
            dim = layer.num_features
            gamma = layer.weight.detach().cpu().numpy()      # [dim]
            beta = layer.bias.detach().cpu().numpy()          # [dim]
            run_mean = layer.running_mean.detach().cpu().numpy()  # [dim]
            run_var = layer.running_var.detach().cpu().numpy()    # [dim]
            params.append({"shape": [1, dim], "data": gamma.tolist()})
            params.append({"shape": [1, dim], "data": beta.tolist()})
            params.append({"shape": [1, dim], "data": run_mean.tolist()})
            params.append({"shape": [1, dim], "data": run_var.tolist()})
    return params


def export_weights(encoder, classifier, detector_net, type_net, threat_net, action_net,
                   metrics_list, output_path):
    """Export all models to Go-compatible ModelWeights JSON."""
    weights = {
        "version": "1.0.16",
        "encoder":     export_sequential_params(encoder),
        "classifier":  export_sequential_params(classifier),
        "detector_net": export_sequential_params(detector_net),
        "type_net":    export_sequential_params(type_net),
        "threat_net":  export_sequential_params(threat_net),
        "action_net":  export_sequential_params(action_net),
        "metrics":     metrics_list,
    }
    with open(output_path, "w") as f:
        json.dump(weights, f)
    size_kb = os.path.getsize(output_path) / 1024
    print(f"  Weights saved: {output_path} ({size_kb:.1f} KB)")


# ── Training ─────────────────────────────────────────────────────────────

def train_encoder(encoder, features, labels, config):
    """Phase 1: Triplet loss training with semi-hard negative mining."""
    print("\n[Phase 1] Encoder training (triplet loss + hard negatives)")
    encoder.train()
    optimizer = torch.optim.Adam(encoder.parameters(), lr=config["lr"])
    scheduler = torch.optim.lr_scheduler.CosineAnnealingWarmRestarts(optimizer, T_0=10, T_mult=2)

    # Build augmented dataset
    all_features = [features]
    all_labels = [labels]
    for noise in [0.025, 0.04, 0.05, 0.05, 0.06, 0.075, 0.1, 0.125]:
        aug = augment_features(features, noise)
        all_features.append(aug)
        all_labels.append(labels.copy())
    all_features = np.concatenate(all_features)
    all_labels = np.concatenate(all_labels)

    # Group by family
    family_indices = {}
    for i, label in enumerate(all_labels):
        family_indices.setdefault(int(label), []).append(i)

    features_t = torch.tensor(all_features, device=DEVICE)
    n_samples = len(all_features)

    for epoch in range(config["epochs"]):
        total_loss = 0
        n_batches = 0

        # Shuffle within families
        for fam in family_indices:
            np.random.shuffle(family_indices[fam])

        for fam, indices in family_indices.items():
            if len(indices) < 2:
                continue

            # Get negative indices (all other families)
            neg_indices = []
            for other_fam, other_idx in family_indices.items():
                if other_fam != fam:
                    neg_indices.extend(other_idx)
            if not neg_indices:
                continue

            # Create mini-batches of triplets
            bs = min(config["batch_size"], len(indices) // 2)
            if bs < 2:
                bs = 2

            for start in range(0, len(indices) - bs, bs):
                anchor_idx = indices[start:start + bs]
                pos_idx = [indices[np.random.randint(len(indices))] for _ in range(bs)]

                # Semi-hard negative mining
                with torch.no_grad():
                    anchor_emb = encoder(features_t[anchor_idx])
                    # Sample candidates and pick closest
                    n_cand = min(32, len(neg_indices))
                    cand_idx = np.random.choice(neg_indices, n_cand, replace=False)
                    cand_emb = encoder(features_t[cand_idx])

                    # For each anchor, find closest negative
                    dists = torch.cdist(anchor_emb, cand_emb)  # [bs, n_cand]
                    neg_select = dists.argmin(dim=1).cpu().numpy()
                    neg_idx = [cand_idx[j] for j in neg_select]

                optimizer.zero_grad()
                a_emb = encoder(features_t[anchor_idx])
                p_emb = encoder(features_t[pos_idx])
                n_emb = encoder(features_t[neg_idx])

                loss = F.triplet_margin_loss(a_emb, p_emb, n_emb,
                                             margin=config["triplet_margin"])
                loss.backward()
                torch.nn.utils.clip_grad_norm_(encoder.parameters(), 5.0)
                optimizer.step()

                total_loss += loss.item()
                n_batches += 1

        scheduler.step()
        avg_loss = total_loss / max(1, n_batches)
        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"    Epoch {epoch+1:3d}/{config['epochs']}: loss={avg_loss:.4f} lr={optimizer.param_groups[0]['lr']:.6f}")

    encoder.eval()
    return avg_loss


def train_classifier(classifier, encoder, features, labels, config):
    """Phase 2: Browser classifier training (cross-entropy)."""
    print("\n[Phase 2] Browser classifier training (cross-entropy)")
    classifier.train()
    encoder.eval()

    optimizer = torch.optim.Adam(classifier.parameters(), lr=config["lr"])
    scheduler = torch.optim.lr_scheduler.CosineAnnealingLR(optimizer, T_max=config["epochs"], eta_min=config["lr"] * 0.01)

    # Augment data
    all_features = [features]
    all_labels = [labels]
    for noise in [0.025, 0.04, 0.05, 0.05, 0.06, 0.075, 0.1, 0.125]:
        all_features.append(augment_features(features, noise))
        all_labels.append(labels.copy())
    all_features = np.concatenate(all_features)
    all_labels = np.concatenate(all_labels)

    # Split train/val
    n = len(all_features)
    perm = np.random.permutation(n)
    split = int(n * 0.8)
    train_idx, val_idx = perm[:split], perm[split:]

    features_t = torch.tensor(all_features, device=DEVICE)
    labels_t = torch.tensor(all_labels, device=DEVICE)

    # Class weights for imbalanced data
    class_counts = np.bincount(all_labels, minlength=NUM_BROWSER_FAMILIES)
    weights = 1.0 / np.maximum(class_counts, 1).astype(np.float32)
    weights = weights / weights.sum() * NUM_BROWSER_FAMILIES
    class_weights = torch.tensor(weights, device=DEVICE)

    best_val_acc = 0
    for epoch in range(config["epochs"]):
        classifier.train()
        np.random.shuffle(train_idx)
        total_loss = 0
        n_batches = 0

        for start in range(0, len(train_idx), config["batch_size"]):
            batch_idx = train_idx[start:start + config["batch_size"]]
            if len(batch_idx) < 2:
                continue
            with torch.no_grad():
                emb = encoder(features_t[batch_idx])
            logits = classifier(emb)
            loss = F.cross_entropy(logits, labels_t[batch_idx], weight=class_weights)

            optimizer.zero_grad()
            loss.backward()
            torch.nn.utils.clip_grad_norm_(classifier.parameters(), 5.0)
            optimizer.step()

            total_loss += loss.item()
            n_batches += 1

        scheduler.step()

        # Validation
        classifier.eval()
        with torch.no_grad():
            val_emb = encoder(features_t[val_idx])
            val_logits = classifier(val_emb)
            val_pred = val_logits.argmax(dim=1)
            val_acc = (val_pred == labels_t[val_idx]).float().mean().item()
            best_val_acc = max(best_val_acc, val_acc)

        avg_loss = total_loss / max(1, n_batches)
        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"    Epoch {epoch+1:3d}/{config['epochs']}: loss={avg_loss:.4f} val_acc={val_acc:.1%} lr={optimizer.param_groups[0]['lr']:.6f}")

    classifier.eval()
    print(f"    Best val accuracy: {best_val_acc:.1%}")
    return avg_loss, best_val_acc


def train_forgery_detector(detector_net, type_net, features, config):
    """Phase 3: Forgery detector training (BCE + CE)."""
    print("\n[Phase 3] Forgery detector training (BCE + CE)")
    detector_net.train()
    type_net.train()

    all_params = list(detector_net.parameters()) + list(type_net.parameters())
    optimizer = torch.optim.Adam(all_params, lr=config["lr"])
    scheduler = torch.optim.lr_scheduler.CosineAnnealingLR(optimizer, T_max=config["epochs"], eta_min=config["lr"] * 0.01)

    # Augment real samples
    aug_real = [features]
    for noise in [0.03, 0.05, 0.05, 0.07]:
        aug_real.append(augment_features(features, noise))
    real_features = np.concatenate(aug_real)

    n_real = len(real_features)
    n_forged = int(n_real * config["forgery_ratio"])

    for epoch in range(config["epochs"]):
        # Generate fresh forged samples each epoch (diverse negative sampling)
        forged_features, forged_types = generate_forged_samples(features, n_forged)

        # Combine
        all_fp = np.concatenate([real_features, forged_features])
        all_cross = compute_cross_layer_features(all_fp)
        all_input = np.concatenate([all_fp, all_cross], axis=1)

        is_forged = np.concatenate([np.zeros(n_real), np.ones(n_forged)]).astype(np.float32)
        # Type labels: 0 for real, actual type for forged
        type_labels = np.concatenate([np.zeros(n_real, dtype=np.int64), forged_types])

        # Shuffle
        perm = np.random.permutation(len(all_input))
        all_input = all_input[perm]
        is_forged = is_forged[perm]
        type_labels = type_labels[perm]

        input_t = torch.tensor(all_input, device=DEVICE)
        target_t = torch.tensor(is_forged, device=DEVICE)
        type_t = torch.tensor(type_labels, device=DEVICE)

        total_loss = 0
        n_batches = 0
        bs = config["batch_size"]

        for start in range(0, len(all_input), bs):
            end = min(start + bs, len(all_input))
            if end - start < 2:
                continue
            batch_in = input_t[start:end]
            batch_target = target_t[start:end]
            batch_type = type_t[start:end]

            # Detector loss (binary)
            det_out = detector_net(batch_in).squeeze(-1)
            det_loss = F.binary_cross_entropy(det_out, batch_target)

            # Type loss (only on forged samples, need ≥2 for BatchNorm)
            forged_mask = batch_target > 0.5
            type_loss = torch.tensor(0.0, device=DEVICE)
            if forged_mask.sum() >= 2:
                type_out = type_net(batch_in[forged_mask])
                type_loss = F.cross_entropy(type_out, batch_type[forged_mask])

            loss = det_loss + 0.5 * type_loss

            optimizer.zero_grad()
            loss.backward()
            torch.nn.utils.clip_grad_norm_(all_params, 5.0)
            optimizer.step()

            total_loss += loss.item()
            n_batches += 1

        scheduler.step()
        avg_loss = total_loss / max(1, n_batches)
        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"    Epoch {epoch+1:3d}/{config['epochs']}: loss={avg_loss:.4f}")

    detector_net.eval()
    type_net.eval()
    return avg_loss


def train_threat_assessor(threat_net, action_net, encoder, detector_net, type_net,
                          features, labels, config):
    """Phase 4: Threat assessor training (CE + synthetic behavior)."""
    print("\n[Phase 4] Threat assessor training (CE + synthetic behavior)")
    threat_net.train()
    action_net.train()
    encoder.eval()
    detector_net.eval()
    type_net.eval()

    all_params = list(threat_net.parameters()) + list(action_net.parameters())
    optimizer = torch.optim.Adam(all_params, lr=config["lr"])
    scheduler = torch.optim.lr_scheduler.CosineAnnealingLR(optimizer, T_max=config["epochs"], eta_min=config["lr"] * 0.01)

    # Augment: real + forged samples
    aug_real = [features]
    aug_labels = [labels]
    for noise in [0.03, 0.05, 0.05, 0.07]:
        aug_real.append(augment_features(features, noise))
        aug_labels.append(labels.copy())
    real_features = np.concatenate(aug_real)
    real_labels = np.concatenate(aug_labels)
    n_real = len(real_features)

    n_forged = int(n_real * config["forgery_ratio"])
    forged_features, forged_types = generate_forged_samples(features, n_forged)

    all_fp = np.concatenate([real_features, forged_features])
    all_cross = compute_cross_layer_features(all_fp)
    is_forged = np.concatenate([np.zeros(n_real, dtype=bool), np.ones(n_forged, dtype=bool)])

    # Pre-compute embeddings and forgery results
    fp_t = torch.tensor(all_fp, device=DEVICE)
    cross_t = torch.tensor(all_cross, device=DEVICE)

    with torch.no_grad():
        embeddings = encoder(fp_t)
        det_input = torch.cat([fp_t, cross_t], dim=1)
        forgery_probs = detector_net(det_input).squeeze(-1)
        forgery_type_logits = type_net(det_input)
        forgery_type_probs = F.softmax(forgery_type_logits, dim=1)

    embeddings_np = embeddings.cpu().numpy()
    forgery_probs_np = forgery_probs.cpu().numpy()
    forgery_types_pred = forgery_type_logits.argmax(dim=1).cpu().numpy()
    forgery_type_probs_np = forgery_type_probs.cpu().numpy()

    # Generate behavior features
    behavior = generate_synthetic_behavior(len(all_fp), is_forged)

    # Generate threat/action labels
    threat_labels, action_labels = generate_threat_labels(all_fp, forgery_probs_np, forgery_types_pred)

    # Build threat assessor input: embedding(32) + forgery_prob(1) + type_probs(4) + behavior(8) = 45
    assessor_input = np.concatenate([
        embeddings_np,
        forgery_probs_np.reshape(-1, 1),
        forgery_type_probs_np,
        behavior,
    ], axis=1).astype(np.float32)

    input_t = torch.tensor(assessor_input, device=DEVICE)
    threat_t = torch.tensor(threat_labels, device=DEVICE)
    action_t = torch.tensor(action_labels, device=DEVICE)

    n = len(assessor_input)
    for epoch in range(config["epochs"]):
        threat_net.train()
        action_net.train()
        perm = np.random.permutation(n)
        total_loss = 0
        n_batches = 0
        bs = config["batch_size"]

        for start in range(0, n, bs):
            idx = perm[start:start + bs]
            if len(idx) < 2:
                continue
            batch_in = input_t[idx]

            t_logits = threat_net(batch_in)
            a_logits = action_net(batch_in)
            t_loss = F.cross_entropy(t_logits, threat_t[idx])
            a_loss = F.cross_entropy(a_logits, action_t[idx])
            loss = t_loss + 0.5 * a_loss

            optimizer.zero_grad()
            loss.backward()
            torch.nn.utils.clip_grad_norm_(all_params, 5.0)
            optimizer.step()

            total_loss += loss.item()
            n_batches += 1

        scheduler.step()
        avg_loss = total_loss / max(1, n_batches)
        if (epoch + 1) % 20 == 0 or epoch == 0:
            print(f"    Epoch {epoch+1:3d}/{config['epochs']}: loss={avg_loss:.4f}")

    threat_net.eval()
    action_net.eval()
    return avg_loss


# ── Main ─────────────────────────────────────────────────────────────────

def main():
    print("=" * 60)
    print("  Fingerprint ML Training — PyTorch GPU")
    print("=" * 60)
    print(f"Device: {DEVICE}")
    if torch.cuda.is_available():
        print(f"GPU:    {torch.cuda.get_device_name(0)}")
        print(f"VRAM:   {torch.cuda.get_device_properties(0).total_memory / 1024**3:.1f} GB")

    # Load data
    data_path = Path(__file__).parent.parent / "training" / "profile_features.json"
    if not data_path.exists():
        data_path = Path("training/profile_features.json")
    print(f"\nLoading profiles from {data_path}")
    features, labels, browser_types = load_profiles(str(data_path))
    print(f"  {len(features)} profiles, {FINGERPRINT_DIM}-dim features")

    family_counts = {}
    for bt in browser_types:
        family_counts[bt] = family_counts.get(bt, 0) + 1
    for fam, cnt in sorted(family_counts.items()):
        print(f"    {fam:12s} {cnt:3d}")

    # Training config
    config = {
        "epochs": 200,
        "batch_size": 64,
        "lr": 0.001,
        "triplet_margin": 1.0,
        "forgery_ratio": 1.5,
    }
    print(f"\nTraining config: {config}")

    # Create models
    encoder = FingerprintEncoder().to(DEVICE)
    classifier = BrowserClassifier().to(DEVICE)
    detector_net = ForgeryDetectorNet().to(DEVICE)
    type_net = ForgeryTypeNet().to(DEVICE)
    threat_net = ThreatNet().to(DEVICE)
    action_net = ActionNet().to(DEVICE)

    total_params = sum(p.numel() for m in [encoder, classifier, detector_net, type_net, threat_net, action_net]
                       for p in m.parameters())
    print(f"Total parameters: {total_params:,}")

    start_time = time.time()
    metrics = []

    # Phase 1: Encoder
    enc_loss = train_encoder(encoder, features, labels, config)
    metrics.append({"epoch": config["epochs"], "encoder_loss": enc_loss})

    # Phase 2: Classifier
    cls_loss, val_acc = train_classifier(classifier, encoder, features, labels, config)
    metrics.append({"epoch": config["epochs"], "class_loss": cls_loss, "val_accuracy": val_acc})

    # Phase 3: Forgery detector
    forg_loss = train_forgery_detector(detector_net, type_net, features, config)
    metrics.append({"epoch": config["epochs"], "forgery_loss": forg_loss})

    # Phase 4: Threat assessor
    threat_loss = train_threat_assessor(threat_net, action_net, encoder, detector_net, type_net,
                                        features, labels, config)
    metrics.append({"epoch": config["epochs"], "threat_loss": threat_loss})

    elapsed = time.time() - start_time
    print(f"\n{'=' * 60}")
    print(f"  Training completed in {elapsed:.1f}s")
    print(f"{'=' * 60}")

    # Export weights
    output_dir = Path("models")
    output_dir.mkdir(exist_ok=True)

    weights_path = output_dir / "weights.json"
    export_weights(encoder, classifier, detector_net, type_net, threat_net, action_net,
                   metrics, str(weights_path))

    # Also save PyTorch checkpoint
    torch.save({
        "encoder": encoder.state_dict(),
        "classifier": classifier.state_dict(),
        "detector_net": detector_net.state_dict(),
        "type_net": type_net.state_dict(),
        "threat_net": threat_net.state_dict(),
        "action_net": action_net.state_dict(),
        "config": config,
        "metrics": metrics,
    }, str(output_dir / "checkpoint.pt"))
    print(f"  PyTorch checkpoint: {output_dir / 'checkpoint.pt'}")

    # Validation
    print(f"\n{'=' * 60}")
    print("  Validation")
    print(f"{'=' * 60}")
    encoder.eval()
    classifier.eval()
    detector_net.eval()
    type_net.eval()
    threat_net.eval()
    action_net.eval()

    feat_t = torch.tensor(features[:5], device=DEVICE)
    with torch.no_grad():
        emb = encoder(feat_t)
        cls_out = F.softmax(classifier(emb), dim=1)
        cross = torch.tensor(compute_cross_layer_features(features[:5]), device=DEVICE)
        det_in = torch.cat([feat_t, cross], dim=1)
        forg_prob = detector_net(det_in).squeeze(-1)

    families = ["chrome", "firefox", "safari", "edge", "opera", "brave", "samsung"]
    for i in range(5):
        pred_fam = families[cls_out[i].argmax().item()]
        conf = cls_out[i].max().item()
        fprob = forg_prob[i].item()
        print(f"  Profile {browser_types[i]:12s} → predicted: {pred_fam:8s} ({conf:.1%}) forgery: {fprob:.1%}")

    print("\nDone.")


if __name__ == "__main__":
    main()
