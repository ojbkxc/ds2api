package localtool

type MemoryToolExecutor struct {
	storage MemoryStorage
}

func NewMemoryToolExecutor(storage MemoryStorage) *MemoryToolExecutor {
	return &MemoryToolExecutor{storage: storage}
}

func (e *MemoryToolExecutor) GetDescriptor() ToolDescriptor {
	return ToolDescriptor{
		ID:             "local:memory:memory_save",
		Name:           "memory_save",
		InvocationName: "memory_save",
		Title:          "保存记忆",
		Description:    "保存新的记忆内容",
		InputSchema: ToolDescriptorSchema{
			Type: "object",
			Properties: map[string]JsonValue{
				"type":    map[string]JsonValue{"type": "string", "enum": []string{"user", "feedback", "topic", "reference"}, "description": "记忆类型"},
				"name":    map[string]JsonValue{"type": "string", "description": "记忆名称"},
				"content": map[string]JsonValue{"type": "string", "description": "记忆内容"},
				"tags":    map[string]JsonValue{"type": "array", "items": map[string]JsonValue{"type": "string"}, "description": "标签列表"},
			},
			Required:             []string{"type", "name", "content", "tags"},
			AdditionalProperties: false,
		},
		Execution: ToolDescriptorExecution{
			Mode:    ToolExecutionModeAuto,
			Enabled: true,
			Risk:    ToolRiskLevelLow,
		},
		Provider: ToolProviderIdentity{
			Kind:        ToolProviderKindLocal,
			ID:          "memory",
			DisplayName: "Memory",
			Transport:   ToolTransportKindInProcess,
		},
	}
}

func (e *MemoryToolExecutor) Execute(call ToolCall, context ToolExecutionContext) (*ToolResult, error) {
	startTime := time.Now()
	switch call.Name {
	case "memory_save":
		return e.saveMemory(call, startTime)
	case "memory_update":
		return e.updateMemory(call, startTime)
	case "memory_delete":
		return e.deleteMemory(call, startTime)
	default:
		return &ToolResult{
			Ok:          false,
			Name:        call.Name,
			Summary:     "Unsupported memory tool",
			Error:       &ToolError{Code: "memory_tool_unsupported", Message: "Unsupported memory tool: " + call.Name, Retryable: false},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}
}

func (e *MemoryToolExecutor) saveMemory(call ToolCall, startTime time.Time) (*ToolResult, error) {
	typeVal, ok := call.Payload["type"].(string)
	if !ok {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Invalid payload", Error: &ToolError{Code: "memory_invalid_payload", Message: "type is required", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}
	memoryType := MemoryType(typeVal)
	if memoryType != MemoryTypeUser && memoryType != MemoryTypeFeedback && memoryType != MemoryTypeTopic && memoryType != MemoryTypeReference {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Invalid payload", Error: &ToolError{Code: "memory_invalid_payload", Message: "invalid type", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	name, ok := call.Payload["name"].(string)
	if !ok || name == "" {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Invalid payload", Error: &ToolError{Code: "memory_invalid_payload", Message: "name is required", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	content, ok := call.Payload["content"].(string)
	if !ok || content == "" {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Invalid payload", Error: &ToolError{Code: "memory_invalid_payload", Message: "content is required", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	tags, ok := call.Payload["tags"].([]interface{})
	if !ok {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Invalid payload", Error: &ToolError{Code: "memory_invalid_payload", Message: "tags must be array", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	tagStrings := make([]string, 0, len(tags))
	for _, tag := range tags {
		if s, ok := tag.(string); ok {
			tagStrings = append(tagStrings, s)
		}
	}

	memory, err := e.storage.SaveMemory(NewMemory{
		Type:        memoryType,
		Name:        name,
		Description: name,
		Content:     content,
		Tags:        tagStrings,
		Pinned:      false,
	})
	if err != nil {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Save failed", Error: &ToolError{Code: "memory_save_failed", Message: err.Error(), Retryable: true}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	return &ToolResult{
		Ok:          true,
		Name:        call.Name,
		Summary:     "Memory saved",
		Detail:      name,
		Output:      map[string]interface{}{"id": memory.ID},
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		DurationMs:  time.Since(startTime).Milliseconds(),
	}, nil
}

func (e *MemoryToolExecutor) updateMemory(call ToolCall, startTime time.Time) (*ToolResult, error) {
	idVal, ok := call.Payload["id"].(float64)
	if !ok || idVal <= 0 {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Invalid ID", Error: &ToolError{Code: "memory_invalid_id", Message: "id is required", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}
	id := int(idVal)

	existing, err := e.storage.GetMemoryByID(id)
	if err != nil {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Memory not found", Error: &ToolError{Code: "memory_not_found", Message: err.Error(), Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	typeVal, ok := call.Payload["type"].(string)
	if ok {
		memoryType := MemoryType(typeVal)
		if memoryType == MemoryTypeUser || memoryType == MemoryTypeFeedback || memoryType == MemoryTypeTopic || memoryType == MemoryTypeReference {
			existing.Type = memoryType
		}
	}

	name, ok := call.Payload["name"].(string)
	if ok && name != "" {
		existing.Name = name
		existing.Description = name
	}

	content, ok := call.Payload["content"].(string)
	if ok && content != "" {
		existing.Content = content
	}

	tags, ok := call.Payload["tags"].([]interface{})
	if ok {
		tagStrings := make([]string, 0, len(tags))
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				tagStrings = append(tagStrings, s)
			}
		}
		existing.Tags = tagStrings
	}

	err = e.storage.UpdateMemory(*existing)
	if err != nil {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Update failed", Error: &ToolError{Code: "memory_update_failed", Message: err.Error(), Retryable: true}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	return &ToolResult{
		Ok:          true,
		Name:        call.Name,
		Summary:     "Memory updated",
		Detail:      existing.Name,
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		DurationMs:  time.Since(startTime).Milliseconds(),
	}, nil
}

func (e *MemoryToolExecutor) deleteMemory(call ToolCall, startTime time.Time) (*ToolResult, error) {
	idVal, ok := call.Payload["id"].(float64)
	if !ok || idVal <= 0 {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Invalid ID", Error: &ToolError{Code: "memory_invalid_id", Message: "id is required", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}
	id := int(idVal)

	err := e.storage.DeleteMemory(id)
	if err != nil {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Delete failed", Error: &ToolError{Code: "memory_delete_failed", Message: err.Error(), Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	return &ToolResult{
		Ok:          true,
		Name:        call.Name,
		Summary:     "Memory deleted",
		Detail:      "#" + fmtInt(id),
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		DurationMs:  time.Since(startTime).Milliseconds(),
	}, nil
}
