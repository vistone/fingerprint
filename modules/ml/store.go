// Package ml — store.go provides persistent, versioned model storage.
//
// ┌─────────────────────────────────────────────────────────────────────┐
// │                  ModelStore — Persistent Model Library              │
// │                                                                     │
// │  Versioned model snapshots on disk:                                 │
// │    <baseDir>/                                                       │
// │      manifest.json         ← index of all versions                  │
// │      v1/weights.json       ← snapshot 1                             │
// │      v2/weights.json       ← snapshot 2 (evolved from v1)          │
// │      ...                                                            │
// │      latest/weights.json   ← symlink to newest version             │
// │                                                                     │
// │  Auto-load on startup: pipeline.LoadFromStore(store)                │
// │  Auto-save after train: store.Save(pipeline, metadata)              │
// │  Incremental evolution: trainer.Evolve(pipeline, newProfiles)       │
// │                                                                     │
// │  Keeps at most MaxVersions snapshots; prunes oldest automatically.  │
// └─────────────────────────────────────────────────────────────────────┘
package ml

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// ModelStore — versioned model persistence
// =========================================================================

// StoreConfig holds model store configuration.
type StoreConfig struct {
	BaseDir     string // root directory for model storage
	MaxVersions int    // maximum number of versions to keep (0 = unlimited)
}

// DefaultStoreConfig returns sensible defaults.
func DefaultStoreConfig(baseDir string) *StoreConfig {
	return &StoreConfig{
		BaseDir:     baseDir,
		MaxVersions: 10,
	}
}

// ModelManifest is the index of all stored model versions.
type ModelManifest struct {
	Versions []ModelVersion `json:"versions"`
}

// ModelVersion describes a single stored model snapshot.
type ModelVersion struct {
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	ParentVer   int       `json:"parent_version,omitempty"` // 0 = trained from scratch
	Epochs      int       `json:"epochs"`
	TrainLoss   float64   `json:"train_loss,omitempty"`
	ValAccuracy float64   `json:"val_accuracy,omitempty"`
	Description string    `json:"description,omitempty"`
}

// ModelStore manages versioned model snapshots on disk.
type ModelStore struct {
	config   *StoreConfig
	manifest ModelManifest
	mu       sync.Mutex
}

// NewModelStore opens or creates a model store at the configured directory.
func NewModelStore(config *StoreConfig) (*ModelStore, error) {
	if config == nil || config.BaseDir == "" {
		return nil, fmt.Errorf("model store base directory is required")
	}

	if err := os.MkdirAll(config.BaseDir, 0750); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}

	s := &ModelStore{config: config}

	if err := s.loadManifest(); err != nil {
		// No manifest yet — start fresh.
		s.manifest = ModelManifest{}
	}

	return s, nil
}

// Latest returns the latest model version, or nil if the store is empty.
func (s *ModelStore) Latest() *ModelVersion {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.manifest.Versions) == 0 {
		return nil
	}
	v := s.manifest.Versions[len(s.manifest.Versions)-1]
	return &v
}

// VersionCount returns the number of stored versions.
func (s *ModelStore) VersionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.manifest.Versions)
}

// Save persists the current pipeline weights as a new version.
func (s *ModelStore) Save(pipeline *ModelPipeline, description string, metrics *TrainingMetrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextVer := 1
	parentVer := 0
	if len(s.manifest.Versions) > 0 {
		latest := s.manifest.Versions[len(s.manifest.Versions)-1]
		nextVer = latest.Version + 1
		parentVer = latest.Version
	}

	// Create version directory
	versionDir := filepath.Join(s.config.BaseDir, fmt.Sprintf("v%d", nextVer))
	if err := os.MkdirAll(versionDir, 0750); err != nil {
		return fmt.Errorf("create version directory: %w", err)
	}

	// Save weights
	weightsPath := filepath.Join(versionDir, "weights.json")
	if err := pipeline.SaveWeights(weightsPath); err != nil {
		return fmt.Errorf("save weights: %w", err)
	}

	// Record version metadata
	mv := ModelVersion{
		Version:     nextVer,
		CreatedAt:   time.Now().UTC(),
		ParentVer:   parentVer,
		Description: description,
	}
	if metrics != nil {
		mv.Epochs = metrics.Epoch
		mv.TrainLoss = metrics.EncoderLoss + metrics.ClassLoss + metrics.ForgeryLoss + metrics.ThreatLoss
		mv.ValAccuracy = metrics.ValAccuracy
	}
	s.manifest.Versions = append(s.manifest.Versions, mv)

	// Write manifest
	if err := s.saveManifest(); err != nil {
		return fmt.Errorf("update manifest: %w", err)
	}

	// Prune old versions
	if s.config.MaxVersions > 0 && len(s.manifest.Versions) > s.config.MaxVersions {
		s.pruneOldVersions()
	}

	return nil
}

// Load restores pipeline weights from a specific version.
func (s *ModelStore) Load(pipeline *ModelPipeline, version int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	weightsPath := filepath.Join(s.config.BaseDir, fmt.Sprintf("v%d", version), "weights.json")
	return pipeline.LoadWeights(weightsPath)
}

// LoadLatest restores the most recent version. Returns false if store is empty.
func (s *ModelStore) LoadLatest(pipeline *ModelPipeline) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.manifest.Versions) == 0 {
		return false, nil
	}

	latest := s.manifest.Versions[len(s.manifest.Versions)-1]
	weightsPath := filepath.Join(s.config.BaseDir, fmt.Sprintf("v%d", latest.Version), "weights.json")
	if err := pipeline.LoadWeights(weightsPath); err != nil {
		return false, err
	}
	return true, nil
}

// ListVersions returns all stored versions, oldest first.
func (s *ModelStore) ListVersions() []ModelVersion {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ModelVersion, len(s.manifest.Versions))
	copy(out, s.manifest.Versions)
	return out
}

// =========================================================================
// Internal helpers
// =========================================================================

func (s *ModelStore) manifestPath() string {
	return filepath.Join(s.config.BaseDir, "manifest.json")
}

func (s *ModelStore) loadManifest() error {
	data, err := os.ReadFile(s.manifestPath())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.manifest)
}

func (s *ModelStore) saveManifest() error {
	data, err := json.MarshalIndent(&s.manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.manifestPath(), data, 0600)
}

func (s *ModelStore) pruneOldVersions() {
	for len(s.manifest.Versions) > s.config.MaxVersions {
		oldest := s.manifest.Versions[0]
		versionDir := filepath.Join(s.config.BaseDir, fmt.Sprintf("v%d", oldest.Version))
		os.RemoveAll(versionDir) // best-effort cleanup
		s.manifest.Versions = s.manifest.Versions[1:]
	}
	// Best-effort manifest update after pruning
	_ = s.saveManifest()
}

// =========================================================================
// Incremental evolution — fine-tune from existing weights
// =========================================================================

// EvolveConfig configures incremental model evolution.
type EvolveConfig struct {
	Epochs       int     // fine-tuning epochs (typically 5-20, much less than full training)
	LearningRate float64 // lower LR for fine-tuning (e.g. 0.0001)
	AugmentNoise float64 // data augmentation noise
}

// DefaultEvolveConfig returns sensible defaults for incremental evolution.
func DefaultEvolveConfig() *EvolveConfig {
	return &EvolveConfig{
		Epochs:       10,
		LearningRate: 0.0001,
		AugmentNoise: 0.01,
	}
}

// Evolve performs incremental fine-tuning on the pipeline's existing weights.
// Unlike TrainFromProfiles (full training from scratch), Evolve:
//   - Starts from current weights (not random initialization)
//   - Uses a lower learning rate for stability
//   - Runs fewer epochs
//   - Preserves previously learned patterns while adapting to new data
//
// This implements the "continuously evolving" model the user described:
// the model library is saved, and only fine-tuned — never retrained from scratch
// unless explicitly requested.
func (t *NeuralTrainer) Evolve(registry *profiles.ProfileRegistry, config *EvolveConfig) (*TrainingMetrics, error) {
	if !t.Pipeline.Trained() {
		return nil, fmt.Errorf("cannot evolve untrained model; run TrainFromProfiles first")
	}
	if config == nil {
		config = DefaultEvolveConfig()
	}

	allProfiles := registry.GetAll()
	if len(allProfiles) == 0 {
		return nil, fmt.Errorf("no profiles available for evolution")
	}

	// Override trainer config for fine-tuning
	saved := *t.Config
	t.Config.Epochs = config.Epochs
	t.Config.LearningRate = config.LearningRate
	t.Config.AugmentNoise = config.AugmentNoise
	defer func() { *t.Config = saved }()

	trainSet, valSet := t.buildTrainingData(allProfiles)

	// Phase 1: Fine-tune encoder (fewer epochs, lower LR)
	if err := t.trainEncoder(trainSet); err != nil {
		return nil, fmt.Errorf("encoder fine-tuning failed: %w", err)
	}

	// Phase 2: Fine-tune classifier
	if err := t.trainClassifier(trainSet, valSet); err != nil {
		return nil, fmt.Errorf("classifier fine-tuning failed: %w", err)
	}

	// Phase 3: Fine-tune forgery detector
	if err := t.trainForgeryDetector(trainSet); err != nil {
		return nil, fmt.Errorf("forgery detector fine-tuning failed: %w", err)
	}

	// Phase 4: Fine-tune threat assessor
	if err := t.trainThreatAssessor(trainSet); err != nil {
		return nil, fmt.Errorf("threat assessor fine-tuning failed: %w", err)
	}

	// Return the latest training metrics
	var latest *TrainingMetrics
	if len(t.Metrics) > 0 {
		m := t.Metrics[len(t.Metrics)-1]
		latest = &m
	}

	return latest, nil
}

// =========================================================================
// Pipeline convenience methods for store integration
// =========================================================================

// LoadFromStore loads the latest model from the store.
// Returns true if a model was loaded, false if the store is empty (no error).
func (p *ModelPipeline) LoadFromStore(store *ModelStore) (bool, error) {
	return store.LoadLatest(p)
}

// SaveToStore saves the current weights to the store as a new version.
func (p *ModelPipeline) SaveToStore(store *ModelStore, description string, metrics *TrainingMetrics) error {
	return store.Save(p, description, metrics)
}

// =========================================================================
// Full workflow: load → evolve → save
// =========================================================================

// EvolveAndSave is a convenience method that performs the full evolution cycle:
// 1. Fine-tune existing weights with new/updated profiles
// 2. Save evolved model as a new version in the store
//
// Returns the new model version number and training metrics.
func (t *NeuralTrainer) EvolveAndSave(
	registry *profiles.ProfileRegistry,
	store *ModelStore,
	evolveConfig *EvolveConfig,
) (int, *TrainingMetrics, error) {
	metrics, err := t.Evolve(registry, evolveConfig)
	if err != nil {
		return 0, nil, err
	}

	versions := store.ListVersions()
	parentVer := 0
	if len(versions) > 0 {
		parentVer = versions[len(versions)-1].Version
	}

	description := fmt.Sprintf("evolved from v%d", parentVer)
	if err := store.Save(t.Pipeline, description, metrics); err != nil {
		return 0, metrics, fmt.Errorf("save evolved model: %w", err)
	}

	newVersions := store.ListVersions()
	newVer := 0
	if len(newVersions) > 0 {
		newVer = newVersions[len(newVersions)-1].Version
	}

	return newVer, metrics, nil
}

// TrainAndSave performs full training from scratch and saves to the store.
// Use this only for initial training; prefer EvolveAndSave for subsequent updates.
func (t *NeuralTrainer) TrainAndSave(
	registry *profiles.ProfileRegistry,
	store *ModelStore,
) (int, error) {
	if err := t.TrainFromProfiles(registry); err != nil {
		return 0, err
	}

	var metrics *TrainingMetrics
	if len(t.Metrics) > 0 {
		m := t.Metrics[len(t.Metrics)-1]
		metrics = &m
	}

	if err := store.Save(t.Pipeline, "initial training", metrics); err != nil {
		return 0, fmt.Errorf("save trained model: %w", err)
	}

	versions := store.ListVersions()
	if len(versions) > 0 {
		return versions[len(versions)-1].Version, nil
	}
	return 1, nil
}

// =========================================================================
// Sorted version helpers
// =========================================================================

// SortVersions sorts model versions by version number ascending.
func SortVersions(versions []ModelVersion) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version < versions[j].Version
	})
}
