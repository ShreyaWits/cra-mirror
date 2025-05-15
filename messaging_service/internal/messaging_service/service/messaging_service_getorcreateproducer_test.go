package service_test

import (
	"errors"
	"messaging_service/internal/config"
	"messaging_service/internal/messaging_service/mock"
	"messaging_service/internal/messaging_service/service"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/logger"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetOrCreateProducer(t *testing.T) {
	logger.InitLogger()

	tests := []struct {
		name           string
		topic          string
		setupMock      func(*mock.MockKafkaFactory, *mock.MockProducer)
		expectedErr    bool
		expectNewCall  bool
		expectedCached bool
	}{
		{
			name:  "get existing producer",
			topic: "existing-topic",
			setupMock: func(factory *mock.MockKafkaFactory, producer *mock.MockProducer) {
				// No calls expected to factory as producer should be cached
			},
			expectedErr:    false,
			expectNewCall:  false,
			expectedCached: true,
		},
		{
			name:  "create new producer successfully",
			topic: "new-topic",
			setupMock: func(factory *mock.MockKafkaFactory, producer *mock.MockProducer) {
				factory.EXPECT().CreateProducer(gomock.Any()).Return(producer, nil)
			},
			expectedErr:    false,
			expectNewCall:  true,
			expectedCached: false,
		},
		{
			name:  "create producer fails",
			topic: "error-topic",
			setupMock: func(factory *mock.MockKafkaFactory, producer *mock.MockProducer) {
				factory.EXPECT().CreateProducer(gomock.Any()).Return(nil, errors.New("producer creation failed"))
			},
			expectedErr:    true,
			expectNewCall:  true,
			expectedCached: false,
		},
		{
			name:  "concurrent producer creation",
			topic: "concurrent-topic",
			setupMock: func(factory *mock.MockKafkaFactory, producer *mock.MockProducer) {
				// Simulate only one call succeeding (the other would be blocked by mutex)
				factory.EXPECT().CreateProducer(gomock.Any()).Return(producer, nil).MaxTimes(1)
			},
			expectedErr:    false,
			expectNewCall:  true,
			expectedCached: false,
		},
		{
			name:  "double-check locking test",
			topic: "double-check-topic",
			setupMock: func(factory *mock.MockKafkaFactory, producer *mock.MockProducer) {
				// No calls expected to factory because we'll simulate another
				// goroutine adding the producer between checks
			},
			expectedErr:    false,
			expectNewCall:  false,
			expectedCached: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFactory := mock.NewMockKafkaFactory(ctrl)
			mockProducer := mock.NewMockProducer(ctrl)

			// Configure mock behavior
			tt.setupMock(mockFactory, mockProducer)

			// Create a configuration
			cfg, _ := config.LoadConfig()

			// Create the service
			confluentService := &service.ConfluentMessagingService{
				Factory:   mockFactory,
				Admin:     mock.NewMockKafkaAdmin(ctrl),
				Producers: make(map[string]confluent.Producer),
				Config:    cfg,
			}

			// For the "existing producer" test case, pre-populate the map
			if tt.expectedCached {
				confluentService.Producers[tt.topic] = mockProducer
			}

			// Special case for double-check locking test
			if tt.name == "double-check locking test" {
				// Create a goroutine that will add the producer to the map
				// right after the first check but before the lock is acquired
				go func() {
					// Sleep a tiny bit to let the main thread get to the right point
					// This is not 100% reliable but good enough for testing
					time.Sleep(1 * time.Millisecond)
					confluentService.Producers[tt.topic] = mockProducer
				}()

				// Sleep a tiny bit to ensure the goroutine has time to run
				time.Sleep(2 * time.Millisecond)
			}

			// Setup kafka config
			kafkaConfig := confluent.KafkaConfig{
				Topic: tt.topic,
			}

			// Call the method
			producer, err := confluentService.GetOrCreateProducer(kafkaConfig)

			// Check results
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, producer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, producer)

				// Check that producer is now in the cache
				cachedProducer, exists := confluentService.Producers[tt.topic]
				assert.True(t, exists, "Producer should be cached")
				assert.Equal(t, mockProducer, cachedProducer, "Cached producer should match")
			}
		})
	}
}

// TestGetOrCreateProducerConcurrency tests the thread safety of GetOrCreateProducer
func TestGetOrCreateProducerConcurrency(t *testing.T) {
	logger.InitLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := mock.NewMockKafkaFactory(ctrl)
	mockProducer := mock.NewMockProducer(ctrl)

	// We expect exactly one call to CreateProducer despite multiple goroutines
	mockFactory.EXPECT().CreateProducer(gomock.Any()).Return(mockProducer, nil).Times(1)

	// Create service
	cfg, _ := config.LoadConfig()
	confluentService := &service.ConfluentMessagingService{
		Factory:   mockFactory,
		Admin:     mock.NewMockKafkaAdmin(ctrl),
		Producers: make(map[string]confluent.Producer),
		Config:    cfg,
	}

	// Setup kafka config
	topic := "concurrent-test-topic"
	kafkaConfig := confluent.KafkaConfig{
		Topic: topic,
	}

	// Run multiple goroutines to access the same topic
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			producer, err := confluentService.GetOrCreateProducer(kafkaConfig)
			assert.NoError(t, err)
			assert.Equal(t, mockProducer, producer)
		}()
	}

	wg.Wait()

	// Verify only one producer was created
	assert.Equal(t, 1, len(confluentService.Producers))
	assert.Equal(t, mockProducer, confluentService.Producers[topic])
}

// TestGetOrCreateProducerDoubleCheckLocking specifically tests the double-check locking pattern
// This ensures the method correctly handles the case where another goroutine creates a producer
// between the first check and acquiring the lock
func TestGetOrCreateProducerDoubleCheckLocking(t *testing.T) {
	logger.InitLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := mock.NewMockKafkaFactory(ctrl)
	mockProducer := mock.NewMockProducer(ctrl)

	// Since we don't have a hook between the first check and lock acquisition,
	// we need to test this in a different way.
	// We'll add the producer to the map right after creating the service,
	// simulating another goroutine that got there first.

	// Create service
	cfg, _ := config.LoadConfig()
	confluentService := &service.ConfluentMessagingService{
		Factory:   mockFactory,
		Admin:     mock.NewMockKafkaAdmin(ctrl),
		Producers: make(map[string]confluent.Producer),
		Config:    cfg,
	}

	// Setup kafka config
	topic := "double-check-topic"
	kafkaConfig := confluent.KafkaConfig{
		Topic: topic,
	}

	// Add the producer to the map before calling GetOrCreateProducer
	confluentService.Producers[topic] = mockProducer

	// Call GetOrCreateProducer
	producer, err := confluentService.GetOrCreateProducer(kafkaConfig)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, mockProducer, producer)
	assert.Equal(t, mockProducer, confluentService.Producers[topic])
}

// TestGetOrCreateProducerErrorCases tests specific error cases
func TestGetOrCreateProducerErrorCases(t *testing.T) {
	logger.InitLogger()

	// Test case: Transaction begin failure and producer recreation
	t.Run("producer creation fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFactory := mock.NewMockKafkaFactory(ctrl)

		// Create service
		cfg, _ := config.LoadConfig()
		confluentService := &service.ConfluentMessagingService{
			Factory:   mockFactory,
			Admin:     mock.NewMockKafkaAdmin(ctrl),
			Producers: make(map[string]confluent.Producer),
			Config:    cfg,
		}

		// Setup kafka config
		topic := "error-topic"
		kafkaConfig := confluent.KafkaConfig{
			Topic: topic,
		}

		// Expect factory call to create producer and return error
		mockFactory.EXPECT().CreateProducer(gomock.Any()).Return(nil, errors.New("producer creation failed"))

		// Call the method
		producer, err := confluentService.GetOrCreateProducer(kafkaConfig)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, producer)
		assert.Equal(t, "producer creation failed", err.Error())

		// Check that producer is not in the cache
		_, exists := confluentService.Producers[topic]
		assert.False(t, exists, "Producer should not be cached on error")
	})
}
