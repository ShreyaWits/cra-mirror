package mock

import (
	"messaging_service/pkg/confluent"

	"github.com/golang/mock/gomock"
)

// MockConfigurator is a mock implementation of the Configurator interface
type MockConfigurator struct {
	ctrl     *gomock.Controller
	recorder *MockConfiguratorMockRecorder
}

// MockConfiguratorMockRecorder is the mock recorder for MockConfigurator
type MockConfiguratorMockRecorder struct {
	mock *MockConfigurator
}

// NewMockConfigurator creates a new mock instance
func NewMockConfigurator(ctrl *gomock.Controller) *MockConfigurator {
	mock := &MockConfigurator{ctrl: ctrl}
	mock.recorder = &MockConfiguratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigurator) EXPECT() *MockConfiguratorMockRecorder {
	return m.recorder
}

// CreateBaseConfig mocks the CreateBaseConfig method
func (m *MockConfigurator) CreateBaseConfig() confluent.KafkaConfig {
	ret := m.ctrl.Call(m, "CreateBaseConfig")
	ret0, _ := ret[0].(confluent.KafkaConfig)
	return ret0
}

// CreateBaseConfig indicates an expected call of CreateBaseConfig
func (mr *MockConfiguratorMockRecorder) CreateBaseConfig() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBaseConfig", nil)
}

// CreateSubscribeConfig mocks the CreateSubscribeConfig method
func (m *MockConfigurator) CreateSubscribeConfig(topic, groupID string, req interface{}) confluent.KafkaConfig {
	ret := m.ctrl.Call(m, "CreateSubscribeConfig", topic, groupID, req)
	ret0, _ := ret[0].(confluent.KafkaConfig)
	return ret0
}

// CreateSubscribeConfig indicates an expected call of CreateSubscribeConfig
func (mr *MockConfiguratorMockRecorder) CreateSubscribeConfig(topic, groupID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSubscribeConfig", nil, topic, groupID, req)
}
