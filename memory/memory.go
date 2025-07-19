package memory

type (
	Memory struct {
		Key    string       `json:"key" jsonschema:"description=The key of the memory. This is the key of the memory."`
		Value  string       `json:"value" jsonschema:"description=The value of the memory. This is the content of the memory."`
		Source MemorySource `json:"source" jsonschema:"description=The source of the memory. This is the origin of the memory."`
		Tags   []string     `json:"tags,omitempty" jsonschema:"description=The tags of the memory. This is the metadata of the memory."`

		Embedding []float32 `json:"-"`
	}

	MemorySource = string

	// ScoredMemory holds a memory with its similarity score
	ScoredMemory struct {
		Memory *Memory `json:"memory" jsonschema:"description=The memory that was found"`
		Score  float64 `json:"score" jsonschema:"description=The similarity score of the memory to the query (0.0~1.0)"`
	}
)

const (
	MemorySourceUser  MemorySource = "user"
	MemorySourceAgent MemorySource = "agent"
)
