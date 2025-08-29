package config

type KnowledgeConfig struct {
	// Core Database Settings
	// SqliteEnabled controls whether SQLite knowledge service is activated
	// Default: true
	SqliteEnabled bool `json:"sqliteEnabled,omitempty"`

	// SqlitePath specifies the file path for the SQLite database
	// Default: ~/.agentruntime/knowledge.db
	SqlitePath string `json:"sqlitePath,omitempty"`

	// Vector Search Settings
	// VectorEnabled controls whether vector search functionality is enabled
	// Requires OpenAI API key for embeddings
	// Default: true
	VectorEnabled bool `json:"vectorEnabled,omitempty"`

	// Rerank Settings
	// RerankEnabled controls whether to use LLM-based reranking after vector search
	// This improves search accuracy by evaluating semantic relevance
	// Default: true
	RerankEnabled bool `json:"rerankEnabled,omitempty"`

	// RerankModel specifies which LLM model to use for reranking
	// Supports OpenAI models (e.g., "gpt-4o-mini", "gpt-4") or other providers
	// Default: "gpt-5-mini"
	RerankModel string `json:"rerankModel,omitempty"`

	// RerankTopK sets the maximum number of results to return after reranking
	// This is the final number of results presented to the user
	// Default: 10
	RerankTopK int `json:"rerankTopK,omitempty"`

	// RetrievalFactor determines how many candidates to retrieve for reranking
	// Actual retrieval count = RerankTopK × RetrievalFactor
	// Higher values give reranker more options but increase latency
	// Default: 3 (retrieves 3x the final result count)
	RetrievalFactor int `json:"retrievalFactor,omitempty"`

	// UseBatchRerank controls whether to use batch processing for reranking
	// Batch processing evaluates all candidates in a single LLM call (more efficient)
	// Individual processing makes separate calls for each candidate (more accurate)
	// Default: true
	UseBatchRerank bool `json:"useBatchRerank,omitempty"`

	// Query Rewrite Settings
	// QueryRewriteEnabled controls whether to use query rewriting for better search
	// Default: false
	QueryRewriteEnabled bool `json:"queryRewriteEnabled,omitempty"`

	// QueryRewriteStrategy specifies which rewriting strategy to use
	// Options: "hyde" (Hypothetical Document Embeddings), "expansion", "multi", "none"
	// Default: "hyde"
	QueryRewriteStrategy string `json:"queryRewriteStrategy,omitempty"`

	// QueryRewriteModel specifies which LLM model to use for query rewriting
	// Default: same as RerankModel
	QueryRewriteModel string `json:"queryRewriteModel,omitempty"`

	// PDFExtractionMethod specifies which method to use for PDF extraction
	// Options: "llm" (use LLM to extract text), "library" (use standard PDF library to extract text)
	// Default: "llm"
	PDFExtractionMethod string `json:"pdfExtractionMethod,omitempty"`

	// PDFExtractionTextModel specifies which LLM model to use for PDF extraction text
	// Default: "anthropic/claude-4-sonnet"
	PDFExtractionTextModel string `json:"pdfExtractionTextModel,omitempty"`
}

// NewKnowledgeConfig creates a new KnowledgeConfig with sensible defaults
// These defaults can be overridden by environment variables
func NewKnowledgeConfig() *KnowledgeConfig {
	return &KnowledgeConfig{
		// Core Database Settings
		SqliteEnabled: true,
		SqlitePath:    ":memory:",

		// Vector Search Settings
		VectorEnabled: true,

		// Rerank Settings
		RerankEnabled:   true,
		RerankModel:     "openai/gpt-5-mini",
		RerankTopK:      10,
		RetrievalFactor: 3,    // Retrieve 3x candidates for reranking
		UseBatchRerank:  true, // Use batch reranker for better performance

		// Query Rewrite Settings
		QueryRewriteEnabled:  false, // Disabled by default
		QueryRewriteStrategy: "hyde",
		QueryRewriteModel:    "openai/gpt-5-mini",
		// QueryRewriteModel will default to RerankModel if not set

		PDFExtractionTextModel: "openai/gpt-5-mini",
		PDFExtractionMethod:    "library",
	}
}
