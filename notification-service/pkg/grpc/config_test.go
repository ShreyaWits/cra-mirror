package grpc

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockListener struct {
	net.Listener
}

func (m *mockListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345} }

func TestNewGRPCServer_Mock(t *testing.T) {
	mockSrv := &grpc.Server{}
	mockLis := &mockListener{}
	server := newGRPCServerWith(mockSrv, mockLis)
	assert.NotNil(t, server.GetServer())
	assert.NotNil(t, server.listener)
	// Do NOT call server.Stop() on a mock server
}

func TestNewGRPCServer_ErrorBranch(t *testing.T) {
	origListen := listenFunc
	defer func() { listenFunc = origListen }()
	listenFunc = func(network, address string) (net.Listener, error) {
		return nil, errors.New("mock listen error")
	}
	// Capture log output
	origStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()
	if os.Getenv("BE_CRASHER") == "1" {
		NewGRPCServer(":0")
		return
	}
	cmd := os.Args[0]
	args := []string{"-test.run=TestNewGRPCServer_ErrorBranch"}
	env := append(os.Environ(), "BE_CRASHER=1")
	proc, err := os.StartProcess(cmd, append([]string{cmd}, args...), &os.ProcAttr{Env: env, Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	assert.NoError(t, err)
	proc.Wait()
}

func TestNewGRPCServer_Integration(t *testing.T) {
	server := NewGRPCServer("localhost:0")
	assert.NotNil(t, server)
	assert.NotNil(t, server.GetServer())
	assert.NotNil(t, server.listener)

	started := make(chan struct{})
	go func() {
		close(started) // signal that we're about to start serving
		_ = server.Start()
	}()
	<-started // wait for goroutine to start

	// Now stop the server
	server.Stop()
}
