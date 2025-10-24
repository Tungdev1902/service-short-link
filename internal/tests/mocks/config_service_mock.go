package mocks

import "service-short-link/internal/domain"

type ConfigServiceMock struct {
	GetStringFunc             func(key string) string
	GetIntFunc                func(key string) int
	GetBoolFunc               func(key string) bool
	GetDurationFunc           func(key string) string
	GetAPIKeysFunc            func() []string
	GetServerConfigFunc       func() domain.ServerConfig
	GetDatabaseConfigFunc     func() domain.DatabaseConfig
	GetTestDatabaseConfigFunc func() domain.DatabaseConfig
	GetRedisConfigFunc        func() domain.RedisConfig
}

func (m *ConfigServiceMock) GetString(key string) string {
	if m.GetStringFunc != nil {
		return m.GetStringFunc(key)
	}
	return ""
}
func (m *ConfigServiceMock) GetInt(key string) int {
	if m.GetIntFunc != nil {
		return m.GetIntFunc(key)
	}
	return 0
}
func (m *ConfigServiceMock) GetBool(key string) bool {
	if m.GetBoolFunc != nil {
		return m.GetBoolFunc(key)
	}
	return false
}
func (m *ConfigServiceMock) GetDuration(key string) string {
	if m.GetDurationFunc != nil {
		return m.GetDurationFunc(key)
	}
	return ""
}
func (m *ConfigServiceMock) GetAPIKeys() []string {
	if m.GetAPIKeysFunc != nil {
		return m.GetAPIKeysFunc()
	}
	return nil
}
func (m *ConfigServiceMock) GetServerConfig() domain.ServerConfig {
	if m.GetServerConfigFunc != nil {
		return m.GetServerConfigFunc()
	}
	return domain.ServerConfig{}
}
func (m *ConfigServiceMock) GetDatabaseConfig() domain.DatabaseConfig {
	if m.GetDatabaseConfigFunc != nil {
		return m.GetDatabaseConfigFunc()
	}
	return domain.DatabaseConfig{}
}
func (m *ConfigServiceMock) GetTestDatabaseConfig() domain.DatabaseConfig {
	if m.GetTestDatabaseConfigFunc != nil {
		return m.GetTestDatabaseConfigFunc()
	}
	return domain.DatabaseConfig{}
}
func (m *ConfigServiceMock) GetRedisConfig() domain.RedisConfig {
	if m.GetRedisConfigFunc != nil {
		return m.GetRedisConfigFunc()
	}
	return domain.RedisConfig{}
}
