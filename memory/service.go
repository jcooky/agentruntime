package memory

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/firebase/genkit/go/genkit"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/entity"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type (
	Service interface {
		// Knowledge management methods
		IndexKnowledge(ctx context.Context, agentName string, knowledge []map[string]any) error
		RetrieveRelevantKnowledge(ctx context.Context, agentName string, query string, limit int) ([]string, error)
		DeleteAgentKnowledge(ctx context.Context, agentName string) error
		Close() error
	}
	SqliteService struct {
		db           *gorm.DB
		embedder     Embedder
		vecExtLoaded bool
	}

	// Embedder interface for generating embeddings
	Embedder interface {
		Embed(ctx context.Context, texts ...string) ([][]float32, error)
	}

	KnowledgeChunk struct {
		Content  string
		Metadata map[string]any
	}
)

var (
	_ Service = (*SqliteService)(nil)
)

func NewService(ctx context.Context, conf *config.MemoryConfig, logger *slog.Logger, genkit *genkit.Genkit) (Service, error) {
	if !conf.SqliteEnabled {
		return nil, errors.New("sqlite memory service is not enabled. Please check your configuration.")
	}
	if conf.SqlitePath == "" {
		return nil, errors.New("sqlite memory service path is not configured. Please check your configuration.")
	} else if _, err := os.Stat(conf.SqlitePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(conf.SqlitePath), 0755); err != nil {
			return nil, errors.Wrapf(err, "failed to create sqlite directory at %s", conf.SqlitePath)
		} else {
			logger.Info("created sqlite directory", slog.String("path", conf.SqlitePath))
		}
	}

	// Initialize sqlite-vec extension using Go bindings (only if vector functionality is enabled)
	if conf.VectorEnabled {
		sqlite_vec.Auto()
	}

	db, err := gorm.Open(
		sqlite.Open(
			fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=on", conf.SqlitePath),
		),
		&gorm.Config{},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open sqlite database at %s", conf.SqlitePath)
	}

	// Auto-migrate GORM entities
	if err := db.AutoMigrate(&entity.Knowledge{}); err != nil {
		return nil, errors.Wrapf(err, "failed to auto-migrate sqlite database at %s", conf.SqlitePath)
	}

	// Create embedder for RAG functionality
	embedder := NewGenkitEmbedder(genkit)

	service := &SqliteService{
		db:           db,
		embedder:     embedder,
		vecExtLoaded: conf.VectorEnabled,
	}

	// Verify sqlite-vec is working and create vector table (only if vector functionality is enabled)
	if conf.VectorEnabled {
		if err := service.verifyAndCreateVectorTable(); err != nil {
			return nil, errors.Wrapf(err, "failed to initialize sqlite-vec")
		}
	}

	return service, nil
}

func (s *SqliteService) Close() error {
	if sqlDB, err := s.db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			return errors.Wrapf(err, "failed to close database connection")
		}
	}
	return nil
}

// initializeVectorFunctionality initializes sqlite-vec functionality on-demand
func (s *SqliteService) initializeVectorFunctionality() error {
	// Initialize sqlite-vec extension
	sqlite_vec.Auto()

	// Verify and create vector table
	if err := s.verifyAndCreateVectorTable(); err != nil {
		return errors.Wrapf(err, "failed to initialize vector table")
	}

	// Mark vector functionality as loaded
	s.vecExtLoaded = true
	return nil
}

// verifyAndCreateVectorTable verifies sqlite-vec is working and creates the vector table
func (s *SqliteService) verifyAndCreateVectorTable() error {
	// Verify sqlite-vec is loaded by checking vec_version() using GORM's connection pool
	var sqliteVersion, vecVersion string
	err := s.db.Raw("SELECT sqlite_version(), vec_version()").Row().Scan(&sqliteVersion, &vecVersion)
	if err != nil {
		return errors.Wrapf(err, "sqlite-vec extension not properly loaded")
	}

	// Create sqlite-vec virtual table
	return s.createVectorTable()
}

// createVectorTable creates the sqlite-vec virtual table for vector operations
func (s *SqliteService) createVectorTable() error {
	createTableSQL := `
		CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_vectors USING vec0(
			knowledge_id INTEGER,
			agent_name TEXT,
			content TEXT,
			embedding float[1536]
		);
	`

	if err := s.db.Exec(createTableSQL).Error; err != nil {
		return errors.Wrapf(err, "failed to create knowledge_vectors virtual table")
	}

	return nil
}

// IndexKnowledge indexes knowledge documents for an agent
func (s *SqliteService) IndexKnowledge(ctx context.Context, agentName string, knowledge []map[string]any) error {
	if s.embedder == nil {
		// Return error instead of silently failing - this indicates a configuration issue
		return errors.New("embedder is not available - check OpenAI API key configuration. Knowledge indexing requires a valid OpenAI API key")
	}

	// Automatically initialize vector functionality if not already done but knowledge is being indexed
	if !s.vecExtLoaded {
		if err := s.initializeVectorFunctionality(); err != nil {
			return errors.Wrapf(err, "failed to initialize vector functionality for knowledge indexing")
		}
	}

	// First, delete existing knowledge for this agent
	if err := s.DeleteAgentKnowledge(ctx, agentName); err != nil {
		return errors.Wrapf(err, "failed to delete existing knowledge")
	}

	if len(knowledge) == 0 {
		return nil
	}

	// Process knowledge into text chunks
	chunks := s.processKnowledge(knowledge)
	if len(chunks) == 0 {
		return nil
	}

	// Extract text content for embedding
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	// Generate embeddings
	embeddings, err := s.embedder.Embed(ctx, texts...)
	if err != nil {
		return errors.Wrapf(err, "failed to generate embeddings")
	}

	if len(embeddings) != len(chunks) {
		return errors.Errorf("embedding count mismatch: got %d, expected %d", len(embeddings), len(chunks))
	}

	// Save chunks with embeddings using GORM
	db := s.db.WithContext(ctx)

	// Begin transaction for both GORM and sqlite-vec operations
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var knowledgeIDs []uint
	for i, chunk := range chunks {
		var embeddingJSON datatypes.JSONType[[]float32]
		if err := embeddingJSON.Scan(embeddings[i]); err != nil {
			return errors.Wrapf(err, "failed to scan embedding")
		}

		var metadataJSON datatypes.JSONType[map[string]any]
		if err := metadataJSON.Scan(chunk.Metadata); err != nil {
			return errors.Wrapf(err, "failed to scan metadata")
		}

		knowledgeEntity := &entity.Knowledge{
			AgentName: agentName,
			Content:   chunk.Content,
			Metadata:  metadataJSON,
			Embedding: embeddingJSON,
		}

		if err := tx.Create(knowledgeEntity).Error; err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to save knowledge chunk")
		}

		knowledgeIDs = append(knowledgeIDs, knowledgeEntity.ID)
	}

	// Store in sqlite-vec vector table for fast similarity search
	if err := s.indexInVectorTable(ctx, tx, agentName, chunks, embeddings, knowledgeIDs); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to index in vector table")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.Wrapf(err, "failed to commit transaction")
	}

	return nil
}

// indexInVectorTable stores vectors in sqlite-vec virtual table using Go bindings within a transaction
func (s *SqliteService) indexInVectorTable(ctx context.Context, tx *gorm.DB, agentName string, chunks []KnowledgeChunk, embeddings [][]float32, knowledgeIDs []uint) error {
	// Delete existing vectors for this agent using GORM transaction
	if err := tx.Exec("DELETE FROM knowledge_vectors WHERE agent_name = ?", agentName).Error; err != nil {
		return errors.Wrapf(err, "failed to delete existing vectors")
	}

	// Insert new vectors using GORM transaction
	for i, chunk := range chunks {
		// Use sqlite-vec Go bindings to serialize the embedding
		serializedEmbedding, err := sqlite_vec.SerializeFloat32(embeddings[i])
		if err != nil {
			return errors.Wrapf(err, "failed to serialize embedding")
		}

		// Insert using GORM's transaction
		if err := tx.Exec(`
			INSERT INTO knowledge_vectors (knowledge_id, agent_name, content, embedding)
			VALUES (?, ?, ?, ?)
		`, knowledgeIDs[i], agentName, chunk.Content, serializedEmbedding).Error; err != nil {
			return errors.Wrapf(err, "failed to insert knowledge vector")
		}
	}

	return nil
}

// RetrieveRelevantKnowledge retrieves relevant knowledge chunks based on query
func (s *SqliteService) RetrieveRelevantKnowledge(ctx context.Context, agentName string, query string, limit int) ([]string, error) {
	if s.embedder == nil {
		// Gracefully handle when no embedder is available
		return nil, nil
	}

	if !s.vecExtLoaded {
		// If vector functionality is not loaded and no knowledge has been indexed yet, return empty results
		return nil, nil
	}

	// Generate embedding for the query
	embeddings, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate query embedding")
	}

	if len(embeddings) == 0 {
		return nil, errors.Errorf("no embedding generated for query")
	}

	queryEmbedding := embeddings[0]

	return s.retrieveWithSqliteVec(ctx, agentName, queryEmbedding, limit)
}

// retrieveWithSqliteVec uses sqlite-vec for fast vector similarity search with Go bindings
func (s *SqliteService) retrieveWithSqliteVec(ctx context.Context, agentName string, queryEmbedding []float32, limit int) ([]string, error) {
	// Use sqlite-vec Go bindings to serialize the query embedding
	serializedQuery, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to serialize query embedding")
	}

	// Perform vector similarity search using sqlite-vec with GORM's connection pool
	searchSQL := `
		SELECT content, distance
		FROM knowledge_vectors
		WHERE agent_name = ? AND embedding MATCH ?
		ORDER BY distance
		LIMIT ?
	`

	rows, err := s.db.WithContext(ctx).Raw(searchSQL, agentName, serializedQuery, limit).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute sqlite-vec similarity search")
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var content string
		var distance float64
		if err := rows.Scan(&content, &distance); err != nil {
			return nil, errors.Wrapf(err, "failed to scan result row")
		}
		results = append(results, content)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "error iterating result rows")
	}

	return results, nil
}

// DeleteAgentKnowledge removes all knowledge for an agent
func (s *SqliteService) DeleteAgentKnowledge(ctx context.Context, agentName string) error {
	db := s.db.WithContext(ctx)

	// Delete from GORM table
	if err := db.Where("agent_name = ?", agentName).Delete(&entity.Knowledge{}).Error; err != nil {
		return errors.Wrapf(err, "failed to delete agent knowledge from GORM")
	}

	// Delete from sqlite-vec table (only if vector functionality is loaded)
	if s.vecExtLoaded {
		if err := db.Exec("DELETE FROM knowledge_vectors WHERE agent_name = ?", agentName).Error; err != nil {
			return errors.Wrapf(err, "failed to delete agent knowledge from vector table")
		}
	}

	return nil
}

// processKnowledge converts knowledge maps into indexable text chunks
func (s *SqliteService) processKnowledge(knowledge []map[string]any) []KnowledgeChunk {
	var chunks []KnowledgeChunk

	for _, item := range knowledge {
		// Convert the knowledge item to a searchable text representation
		content := s.extractTextFromKnowledge(item)
		if content == "" {
			continue
		}

		chunks = append(chunks, KnowledgeChunk{
			Content:  content,
			Metadata: item,
		})
	}

	return chunks
}

// extractTextFromKnowledge extracts searchable text from a knowledge map
func (s *SqliteService) extractTextFromKnowledge(item map[string]any) string {
	var textParts []string

	// Common text fields to extract (in priority order)
	textFields := []string{"content", "description", "title", "summary", "text", "name"}

	// First, look for standard text fields
	var foundStandardFields []string
	for _, field := range textFields {
		if value, exists := item[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				foundStandardFields = append(foundStandardFields, str)
			}
		}
	}

	// If we found standard text fields, use them
	if len(foundStandardFields) > 0 {
		textParts = foundStandardFields
	} else {
		// If no standard text fields found, try to extract from all string values
		// Sort keys for deterministic ordering
		var keys []string
		for k := range item {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := item[key]
			if str, ok := value.(string); ok && str != "" {
				textParts = append(textParts, fmt.Sprintf("%s: %s", key, str))
			}
		}
	}

	return strings.Join(textParts, " ")
}
