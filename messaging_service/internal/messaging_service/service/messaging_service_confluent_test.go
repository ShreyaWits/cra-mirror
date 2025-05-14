package service_test

import (
	"context"
	pb "cra-protos/messaging_service"
	"errors"
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/mock"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/logger"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestTopicExists(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		exists    bool
		expectErr bool
	}{
		{
			name:      "topic exists",
			topic:     "existing-topic",
			exists:    true,
			expectErr: false,
		},

		{
			name:      "topic does not exist",
			topic:     "missing-topic",
			exists:    false,
			expectErr: false,
		},
		{
			name:      "error listing topics",
			topic:     "any-topic",
			exists:    false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAdmin := mock.NewMockKafkaAdmin(ctrl)
			service := &service.ConfluentMessagingService{
				Factory: mock.NewMockKafkaFactory(ctrl),
				Admin:   mockAdmin,
			}

			// Set up mock expectations based on the test case
			switch tt.name {
			case "topic exists":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{"existing-topic"}, nil)
			case "topic does not exist":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{"other-topic"}, nil)
			case "error listing topics":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return(nil, errors.New("listing error"))
			}

			exists, err := service.TopicExists(context.Background(), tt.topic)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.exists, exists)
			}
		})
	}
}

func TestConsumeMessage(t *testing.T) {
	logger.InitLogger()
	tests := []struct {
		name      string
		topic     string
		group     string
		expectErr bool
	}{
		// {
		// 	name:      "consume message successfully",
		// 	topic:     "consume-topic",
		// 	group:     "consume-group",
		// 	expectErr: false,
		// },
		// {
		// 	name:      "topic does not exist",
		// 	topic:     "missing-topic",
		// 	group:     "consume-group",
		// 	expectErr: true,
		// },
		{
			name:      "error creating consumer",
			topic:     "error-topic",
			group:     "consume-group",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// Create a valid configuration
			cfg, _ := config.LoadConfig()
			mockAdmin := mock.NewMockKafkaAdmin(ctrl)
			mockFactory := mock.NewMockKafkaFactory(ctrl)
			// mockConsumer := mock.NewMockConsumer(ctrl)
			service := &service.ConfluentMessagingService{
				Factory: mockFactory,
				Admin:   mockAdmin,
				Config:  cfg,
			}
			stream := &mockSubscribeV1Server{
				ctx: context.Background(), // or context.WithCancel for controlled shutdown
			}
			switch tt.name {
			// case "consume message successfully":
			// 	mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{tt.topic}, nil)
			// 	mockFactory.EXPECT().CreateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockConsumer, nil)
			// 	// Simulate consumer Start and cancel after mock call
			// 	mockConsumer.EXPECT().Start(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
			// 		stream.ctx.Done()
			// 		return nil
			// 	})

			// Expect Subscribe call
			case "topic does not exist":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{tt.topic}, nil)
				mockFactory.EXPECT().CreateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("creation error"))
				mockFactory.EXPECT().CreateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("creation error"))
				mockFactory.EXPECT().CreateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("creation error"))
			case "error creating consumer":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{tt.topic}, nil)
				mockFactory.EXPECT().CreateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("creation error"))
				mockFactory.EXPECT().CreateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("creation error"))
				mockFactory.EXPECT().CreateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("creation error"))
			}
			configurator := confluent.NewConfigurator([]string{"localhost:9092"}, cfg)
			kafkaConfig := configurator.CreateSubscribeConfig(tt.topic, tt.group, &pb.SubscribeRequest{Topic: tt.topic, GroupId: tt.group})

			// Call the CreateTopic method

			err := service.ConsumeMessage(stream, kafkaConfig)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

type mockSubscribeV1Server struct {
	grpc.ServerStream
	ctx     context.Context
	sent    []*pb.KafkaMessage
	sendErr error
}

func (m *mockSubscribeV1Server) Send(msg *pb.KafkaMessage) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, msg)
	return nil
}

func (m *mockSubscribeV1Server) Context() context.Context {
	return m.ctx
}

func TestPublishMessage(t *testing.T) {
	logger.InitLogger()
	tests := []struct {
		name      string
		topic     string
		value     map[string]string
		cfg       interface{}
		expectErr bool
	}{
		{
			name:      "publish message successfully",
			topic:     "publish-topic",
			value:     map[string]string{"key": "value"},
			cfg:       confluent.KafkaConfig{Topic: "publish-topic"},
			expectErr: false,
		},
		{
			name:      "topic does not exist",
			topic:     "missing-topic",
			value:     map[string]string{"key": "value"},
			cfg:       confluent.KafkaConfig{Topic: "missing-topic"},
			expectErr: true,
		},
		{
			name:      "error publishing message",
			topic:     "error-topic",
			value:     map[string]string{"key": "value"},
			cfg:       confluent.KafkaConfig{Topic: "error-topic"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAdmin := mock.NewMockKafkaAdmin(ctrl)
			mockProducer := mock.NewMockProducer(ctrl)
			service := &service.ConfluentMessagingService{
				Factory: mock.NewMockKafkaFactory(ctrl),
				Admin:   mockAdmin,
				Producers: map[string]confluent.Producer{
					tt.topic: mockProducer,
				},
			}

			switch tt.name {
			case "publish message successfully":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{tt.topic}, nil)
				mockAdmin.EXPECT().HasActiveConsumers(gomock.Any(), tt.topic).Return(true, nil)
				mockProducer.EXPECT().WriteWithRetry(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "topic does not exist":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{"other-topic"}, nil)
			case "error publishing message":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{tt.topic}, nil)
				mockAdmin.EXPECT().HasActiveConsumers(gomock.Any(), tt.topic).Return(true, nil)
				mockProducer.EXPECT().WriteWithRetry(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("publish error"))
			}

			err := service.PublishMessage(context.Background(), tt.cfg, &pb.PublishRequest{Topic: tt.topic, Value: tt.value})
			if tt.expectErr {
				assert.Error(t, err)
			} else {

				assert.Nil(t, err)
			}
		})
	}
}

func TestCreateTopic(t *testing.T) {
	logger.InitLogger()
	tests := []struct {
		name      string
		topic     string
		expectErr bool
	}{
		{
			name:      "create topic successfully",
			topic:     "new1-topic",
			expectErr: false,
		},
		{
			name:      "topic already exists",
			topic:     "existing-topic",
			expectErr: true,
		},
		{
			name:      "error creating topic",
			topic:     "error-topic",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// Create a valid configuration
			cfg, _ := config.LoadConfig()
			mockAdmin := mock.NewMockKafkaAdmin(ctrl)
			service := &service.ConfluentMessagingService{
				Factory: mock.NewMockKafkaFactory(ctrl),
				Admin:   mockAdmin,
				Config:  cfg,
			}

			// Set expectations based on the test case
			switch tt.name {
			case "create topic successfully":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{"any-topic"}, nil)
				mockAdmin.EXPECT().CreateTopic(gomock.Any(), tt.topic, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				// mockAdmin.EXPECT().CreateTopic(gomock.Any(), tt.topic, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				// mockAdmin.EXPECT().CreateTopic(gomock.Any(), tt.topic, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "topic already exists":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{tt.topic}, nil)
			case "error creating topic":
				mockAdmin.EXPECT().ListTopics(gomock.Any()).Return([]string{"any-topic"}, nil)
				mockAdmin.EXPECT().CreateTopic(gomock.Any(), tt.topic, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("creation error"))
				mockAdmin.EXPECT().CreateTopic(gomock.Any(), tt.topic, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("creation error"))
				mockAdmin.EXPECT().CreateTopic(gomock.Any(), tt.topic, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("creation error"))
			}

			kafkaConfig := confluent.NewConfigurator([]string{"localhost:9092"}, cfg)
			t_cfg := kafkaConfig.CreateTopicConfig(tt.topic, &pb.CreateTopicRequest{Topic: tt.topic})

			// Call the CreateTopic method
			err := service.CreateTopic(context.Background(), &pb.CreateTopicRequest{Topic: tt.topic}, t_cfg)

			// Assert the expected outcome
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
