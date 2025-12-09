package shell

// MockExecutor for testing
type MockExecutor struct {
	ExecuteFunc          func(name string, args ...string) ([]byte, error)
	ExecuteWithStdioFunc func(name string, args ...string) error
	LookPathFunc         func(name string) (string, error)
}

func (m *MockExecutor) Execute(name string, args ...string) ([]byte, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(name, args...)
	}
	return []byte("mock output"), nil
}

func (m *MockExecutor) ExecuteWithStdio(name string, args ...string) error {
	if m.ExecuteWithStdioFunc != nil {
		return m.ExecuteWithStdioFunc(name, args...)
	}
	return nil
}

func (m *MockExecutor) LookPath(name string) (string, error) {
	if m.LookPathFunc != nil {
		return m.LookPathFunc(name)
	}
	return "/usr/bin/" + name, nil
}
