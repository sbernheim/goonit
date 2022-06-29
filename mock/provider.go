// Provider is a single source for whatever mocks your tests need.
//
// It uses Gomock under the hood. See: https://github.com/golang/mock
//
// Extend Provider and BaseProvider to add constructor methods that return whatever mocks your tests
// need and store mock instances in your implementation of the Test interface.
//
package mock

// Generate mocks for external modules using mockgen and a `go:generate` annotation like the one for
// MockLogger below.
//
//go:generate mockgen -destination=mockLogger.go -package=mock github.com/go-logr/logr Logger

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
)

type Provider interface {
	Controller() *gomock.Controller
	Logger() *MockLogger
	Finish()
}

type BaseProvider struct {
	t *testing.T
	c *gomock.Controller
}

func NewProvider(t *testing.T) Provider {
	return &BaseProvider{
		t: t,
		c: gomock.NewController(t),
	}
}

func (p *BaseProvider) Controller() *gomock.Controller {
	return p.c
}

func (p *BaseProvider) Logger() *MockLogger {
	return NewMockLogger(p.c)
}

func (p *BaseProvider) Finish() {
	p.c.Finish()
}
