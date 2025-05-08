package di

import (
	"reflect"
	"testing"
)

func TestInitContainer(t *testing.T) {
	// Initialize the container
	container := InitContainer()

	// Check if the container is initialized correctly
	if container == nil {
		t.Error("Container is nil")
	}

	// Check if the container has the correct type
	if reflect.TypeOf(container).String() != "*di.Container" {
		t.Errorf("Expected container type to be *di.Container, but got %s", reflect.TypeOf(container).String())
	}
}

func TestGetContainer(t *testing.T) {
	// Initialize the container
	container := InitContainer()

	// Get the container
	digContainer := container.GetContainer()

	// Check if the container is returned correctly
	if digContainer == nil {
		t.Error("digContainer is nil")
	}

	// Check if the container has the correct type
	if reflect.TypeOf(digContainer).String() != "*dig.Container" {
		t.Errorf("Expected container type to be *dig.Container, but got %s", reflect.TypeOf(digContainer).String())
	}
}

func TestRegisterService(t *testing.T) {
	// Initialize the container
	container := InitContainer()

	// Define a service
	type MyService struct {
		Value string
	}
	service := &MyService{Value: "test"}

	// Register the service
	container.RegisterService(func() *MyService {
		return service
	})

	// Check if the service is registered correctly
	var invokedValue string
	container.InvokeService(func(s *MyService) {
		invokedValue = s.Value
	})

	// Check if the service is invoked correctly
	if invokedValue != "test" {
		t.Errorf("Expected invoked value to be test, but got %s", invokedValue)
	}
}

func TestInvokeService(t *testing.T) {
	// Initialize the container
	container := InitContainer()

	// Define a service
	type MyService struct {
		Value string
	}
	service := &MyService{Value: "test"}

	// Register the service
	container.RegisterService(func() *MyService {
		return service
	})

	// Invoke the service
	var invokedValue string
	container.InvokeService(func(s *MyService) {
		invokedValue = s.Value
	})

	// Check if the service is invoked correctly
	if invokedValue != "test" {
		t.Errorf("Expected invoked value to be test, but got %s", invokedValue)
	}
}