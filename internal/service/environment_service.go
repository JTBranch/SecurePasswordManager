package service

import "os"

// EnvironmentService provides environment-related helpers
// Usage: env := service.NewEnvironmentService()
//
//	env.IsProduction()
type EnvironmentService struct{}

func NewEnvironmentService() *EnvironmentService {
	return &EnvironmentService{}
}

// IsProduction returns true if running in production (GO_PASSWORD_MANAGER_ENV=prod)
func (e *EnvironmentService) IsProduction() bool {
	return os.Getenv("GO_PASSWORD_MANAGER_ENV") == "prod"
}
