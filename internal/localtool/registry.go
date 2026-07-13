package localtool

var DefaultRegistry = NewToolRegistry()
var DefaultMemoryStorage = NewInMemoryStorage()

func init() {
	DefaultRegistry.Register(NewWebSearchExecutor())
	DefaultRegistry.Register(NewWebFetchExecutor())
	DefaultRegistry.Register(NewMemoryToolExecutor(DefaultMemoryStorage))
}

func Execute(call ToolCall, context ToolExecutionContext) (*ToolResult, error) {
	executor, ok := DefaultRegistry.Get(call.Name)
	if !ok {
		return nil, &ToolNotFoundError{Name: call.Name}
	}
	return executor.Execute(call, context)
}

func HasTool(name string) bool {
	return DefaultRegistry.Has(name)
}

func ListTools() []ToolDescriptor {
	return DefaultRegistry.List()
}

func GetMemoryStorage() MemoryStorage {
	return DefaultMemoryStorage
}

type ToolNotFoundError struct{ Name string }

func (e *ToolNotFoundError) Error() string {
	return "tool not found: " + e.Name
}
