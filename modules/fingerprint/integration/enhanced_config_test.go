package fingerprint_test

import (
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/internal/config"
)

func TestEnhancedConfigCenterBasic(t *testing.T) {
	// 首先initialize默认configuration
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize default config: %v", err)
	}

	// get基础configuration中心并包装
	baseCenter := config.GetConfigCenter()
	center := config.WrapConfigCenter(baseCenter)
	defer center.Close()

	// 测试基本功能 - getconfiguration
	currentConfig := center.Get()
	if currentConfig == nil {
		t.Fatal("Current config should not be nil")
	}

	t.Logf("Successfully retrieved config with version: %s", currentConfig.Metadata.Version)
}

func TestConfigHealthCheck(t *testing.T) {
	// initializeconfiguration
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	baseCenter := config.GetConfigCenter()
	center := config.WrapConfigCenter(baseCenter)
	defer center.Close()

	// wait健康checkcomplete
	time.Sleep(100 * time.Millisecond)

	// get健康state
	health := center.GetHealthStatus()
	t.Logf("Health status: Healthy=%v, Issues=%v", health.Healthy, health.Issues)

	// 测试添加健康check函数
	center.AddHealthCheck(func(cfg *config.ManagedConfig) []string {
		var issues []string
		if cfg.Metadata == nil {
			issues = append(issues, "Metadata is missing")
		}
		return issues
	})

	t.Log("Successfully added custom health check")
}

func TestConfigSubscription(t *testing.T) {
	// initializeconfiguration
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	baseCenter := config.GetConfigCenter()
	center := config.WrapConfigCenter(baseCenter)
	defer center.Close()

	// 订阅configuration变更
	eventCh, err := center.Subscribe("test-subscriber")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// create测试configuration
	currentConfig := center.Get()
	testConfig := *currentConfig
	if testConfig.BehaviorAnalysis == nil {
		testConfig.BehaviorAnalysis = &config.BehaviorAnalysisConfig{}
	}
	testConfig.BehaviorAnalysis.MinRequestsForAnalysis = 15

	// 在goroutine中wait事件
	eventsReceived := make(chan bool, 1)
	go func() {
		timeout := time.After(5 * time.Second)
		select {
		case event := <-eventCh:
			if event.Type == config.ConfigChangeTypeSubscribe {
				// ignore订阅confirm事件
				select {
				case updateEvent := <-eventCh:
					if updateEvent.Type == config.ConfigChangeTypeUpdate {
						eventsReceived <- true
					}
				case <-timeout:
					eventsReceived <- false
				}
			}
		case <-timeout:
			eventsReceived <- false
		}
	}()

	// updateconfiguration
	if err := center.Update(&testConfig, "test update", "test"); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// wait事件receive
	select {
	case received := <-eventsReceived:
		if !received {
			t.Error("Did not receive expected configuration update event")
		}
	case <-time.After(6 * time.Second):
		t.Error("Test timeout waiting for events")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	// initializeconfiguration
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	baseCenter := config.GetConfigCenter()
	center := config.WrapConfigCenter(baseCenter)
	defer center.Close()

	// create多个订阅者
	subscriber1, err := center.Subscribe("subscriber-1")
	if err != nil {
		t.Fatalf("Failed to create subscriber 1: %v", err)
	}

	subscriber2, err := center.Subscribe("subscriber-2")
	if err != nil {
		t.Fatalf("Failed to create subscriber 2: %v", err)
	}

	// updateconfiguration
	currentConfig := center.Get()
	testConfig := *currentConfig
	if testConfig.BehaviorAnalysis == nil {
		testConfig.BehaviorAnalysis = &config.BehaviorAnalysisConfig{}
	}
	testConfig.BehaviorAnalysis.MinRequestsForAnalysis = 20

	if err := center.Update(&testConfig, "multi-subscriber test", "test"); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// verify两个订阅者都能收到事件
	timeout := time.After(2 * time.Second)

	received1 := false
	received2 := false

	for !received1 || !received2 {
		select {
		case event := <-subscriber1:
			if event.Type == config.ConfigChangeTypeUpdate {
				received1 = true
			}
		case event := <-subscriber2:
			if event.Type == config.ConfigChangeTypeUpdate {
				received2 = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for events from subscribers")
		}
	}

	if !received1 {
		t.Error("Subscriber 1 did not receive update event")
	}
	if !received2 {
		t.Error("Subscriber 2 did not receive update event")
	}
}

func TestUnsubscribe(t *testing.T) {
	// initializeconfiguration
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	baseCenter := config.GetConfigCenter()
	center := config.WrapConfigCenter(baseCenter)
	defer center.Close()

	// 订阅
	_, err := center.Subscribe("test-unsubscribe")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// cancel订阅
	if err := center.Unsubscribe("test-unsubscribe"); err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// 尝试再次cancel订阅应该failed
	if err := center.Unsubscribe("test-unsubscribe"); err == nil {
		t.Error("Expected error when unsubscribing non-existent subscriber")
	}

	// updateconfiguration不应该导致panic
	currentConfig := center.Get()
	testConfig := *currentConfig
	if testConfig.BehaviorAnalysis == nil {
		testConfig.BehaviorAnalysis = &config.BehaviorAnalysisConfig{}
	}
	testConfig.BehaviorAnalysis.MinRequestsForAnalysis = 25

	if err := center.Update(&testConfig, "unsubscribe test", "test"); err != nil {
		t.Fatalf("Failed to update config after unsubscribe: %v", err)
	}
}
