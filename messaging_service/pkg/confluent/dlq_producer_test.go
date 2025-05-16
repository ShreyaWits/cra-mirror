package confluent_test

import (
	"context"
	"errors"
	"testing"

	"messaging_service/internal/modules/message_broker/mock"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/logger"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewDLQProducer(t *testing.T) {
	tests := []struct {
		name       string
		cfg        confluent.KafkaConfig
		shouldFail bool
	}{
		{
			name:       "valid config",
			cfg:        confluent.NewDefaultKafkaConfig([]string{"localhost:9092"}, "test-topic"),
			shouldFail: false,
		},
		{
			name:       "pruducer not genrated",
			cfg:        confluent.NewDefaultKafkaConfig([]string{"localhost:9092"}, "test-topic"),
			shouldFail: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// dlqProducer,err := confluent.NewFactory().CreateDLQProducer(tc.cfg);
			dlqProducer, err := confluent.NewDLQProducer(tc.cfg)
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, dlqProducer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dlqProducer)
			}
		})
	}
}

func TestSendToDLQ(t *testing.T) {
	logger.InitLogger()

	type fields struct {
		mockSetup func(p *mock.MockKafkaProducerInterface)
	}

	tests := []struct {
		name          string
		fields        fields
		expectErr     bool
		failureReason string
	}{
		{
			name: "successful DLQ send",
			fields: fields{
				mockSetup: func(p *mock.MockKafkaProducerInterface) {
					p.EXPECT().Produce(gomock.Any(), nil).Return(nil)
					p.EXPECT().Flush(gomock.Any()).Return(0)
				},
			},
			expectErr:     false,
			failureReason: "test-failure",
		},
		{
			name: "DLQ produce failure",
			fields: fields{
				mockSetup: func(p *mock.MockKafkaProducerInterface) {
					p.EXPECT().Produce(gomock.Any(), nil).Return(errors.New("produce error"))
				},
			},
			expectErr:     true,
			failureReason: "test-failure",
		},
		{
			name: "flush not complete",
			fields: fields{
				mockSetup: func(p *mock.MockKafkaProducerInterface) {
					p.EXPECT().Produce(gomock.Any(), nil).Return(nil)
					p.EXPECT().Flush(gomock.Any()).Return(1) // one message still in queue
				},
			},
			expectErr:     false,
			failureReason: "flush-warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProducer := mock.NewMockKafkaProducerInterface(ctrl)
			tt.fields.mockSetup(mockProducer)

			cfg := confluent.NewDefaultKafkaConfig([]string{"localhost:9092"}, "test-topic")

			// 💡 INJECT THE MOCK PRODUCER MANUALLY
			dlqProducer := &confluent.DLQProducerImpl{
				Producer: mockProducer,
				Topic:    cfg.Topic,
				Config:   cfg,
			}

			err := dlqProducer.SendToDLQ(context.Background(), "msg-id", []byte("value"), nil, tt.failureReason)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestDLQClose(t *testing.T) {
	logger.InitLogger()
	tests := []struct {
		name      string
		mockSetup func(p *mock.MockKafkaProducerInterface)
	}{
		{
			name: "close flush and shutdown cleanly",
			mockSetup: func(p *mock.MockKafkaProducerInterface) {
				p.EXPECT().Flush(gomock.Any()).Return(0)
				p.EXPECT().Close()
			},
		},
		{
			name: "close with remaining messages",
			mockSetup: func(p *mock.MockKafkaProducerInterface) {
				p.EXPECT().Flush(gomock.Any()).Return(5) // simulate incomplete flush
				p.EXPECT().Close()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProducer := mock.NewMockKafkaProducerInterface(ctrl)
			tc.mockSetup(mockProducer)

			cfg := confluent.NewDefaultKafkaConfig([]string{"localhost:9092"}, "test-topic")
			b := confluent.DefaultDLQConfig()

			// 💡 INJECT THE MOCK PRODUCER MANUALLY
			dlqProducer := &confluent.DLQProducerImpl{
				Producer: mockProducer,
				Topic:    b.TopicSuffix + cfg.Topic,
				Config:   cfg,
			}

			err := dlqProducer.Close()
			assert.NoError(t, err)
		})
	}
}
