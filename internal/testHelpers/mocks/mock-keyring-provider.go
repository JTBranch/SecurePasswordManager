package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockKeyringProvider struct {
	mock.Mock
}

func (m *MockKeyringProvider) Set(service, key, value string) error {
	args := m.Called(service, key, value)
	return args.Error(0)
}

func (m *MockKeyringProvider) Get(service, key string) (string, error) {
	args := m.Called(service, key)
	return args.String(0), args.Error(1)
}

func (m *MockKeyringProvider) Delete(service, key string) error {
	args := m.Called(service, key)
	return args.Error(0)
}
