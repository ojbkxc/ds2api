package localtool

var DefaultRegistry = NewToolRegistry()

func init() {
	DefaultRegistry.Register(NewWebSearchExecutor())
	DefaultRegistry.Register(NewWebFetchExecutor())
}

func Execute(call ToolCall) (*ToolResult, error) {
	executor, ok := DefaultRegistry.Get(call.Name)
	if !ok {
		return nil, &ToolNotFoundError{Name: call.Name}
	}
	return executor.Execute(call)
}

func HasTool(name string) bool {
	return DefaultRegistry.Has(name)
}

func ListTools() []ToolDescriptor {
	return DefaultRegistry.List()
}

type ToolNotFoundError struct{ Name string }

func (e *ToolNotFoundError) Error() string {
	return "tool not found: " + e.Name
}