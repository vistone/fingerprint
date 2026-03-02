package clienthints

import fp "github.com/vistone/fingerprint"

// CHPhase Client Hints 生命周期阶段。
type CHPhase = fp.CHPhase

// CHLifecycleEvent 生命周期事件。
type CHLifecycleEvent = fp.CHLifecycleEvent

// ClientHintsLifecycle 完整生命周期。
type ClientHintsLifecycle = fp.ClientHintsLifecycle

// CHLifecycleManager 生命周期管理器。
type CHLifecycleManager = fp.CHLifecycleManager

const (
	PHASE_INITIAL_REQUEST           = fp.PHASE_INITIAL_REQUEST
	PHASE_SERVER_RESPONSE           = fp.PHASE_SERVER_RESPONSE
	PHASE_SUBSEQUENT_REQUESTS       = fp.PHASE_SUBSEQUENT_REQUESTS
	PHASE_CROSS_ORIGIN_SUB_REQUESTS = fp.PHASE_CROSS_ORIGIN_SUB_REQUESTS
	PHASE_TERMINATED                = fp.PHASE_TERMINATED
)

// NewLifecycleManager 创建生命周期管理器。
func NewLifecycleManager() *CHLifecycleManager {
	return fp.NewCHLifecycleManager()
}
