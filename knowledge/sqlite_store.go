//go:build !without_sqlite

package knowledge

import (
	"context"
	"fmt"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SqliteStore implements Store using SQLite with sqlite-vec extension
type SqliteStore struct {
	db     *gorm.DB
	vecDim int
}

// KnowledgeRecord represents the database structure for knowledge items
type SqliteKnowledgeRecord struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Metadata datatypes.JSONType[map[string]any]

	Documents []*SqliteDocumentRecord `gorm:"foreignKey:KnowledgeRecordID"`
}

// TableName specifies the table name for GORM
func (SqliteKnowledgeRecord) TableName() string {
	return "knowledges"
}

type SqliteDocumentRecord struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Content       datatypes.JSONType[Content]
	EmbeddingText string
	Metadata      datatypes.JSONType[map[string]any]

	KnowledgeRecordID string
	KnowledgeRecord   *SqliteKnowledgeRecord `gorm:"foreignKey:KnowledgeRecordID"`
}

func (SqliteDocumentRecord) TableName() string {
	return "documents"
}

// NewSqliteStore creates a new SQLite-based knowledge store
func NewSqliteStore(dbPath string, dimension int) (*SqliteStore, error) {
	// Initialize sqlite-vec extension
	sqlite_vec.Auto()

	// Open database connection
	db, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=on", dbPath)),
		&gorm.Config{},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open sqlite database")
	}

	store := &SqliteStore{
		db:     db,
		vecDim: dimension,
	}

	// Auto-migrate the knowledge table
	if err := db.AutoMigrate(&SqliteKnowledgeRecord{}, &SqliteDocumentRecord{}); err != nil {
		return nil, errors.Wrapf(err, "failed to migrate knowledge table")
	}

	// Create vector table
	if err := store.createVectorTable(); err != nil {
		return nil, err
	}

	return store, nil
}

// createVectorTable creates the sqlite-vec virtual table
func (s *SqliteStore) createVectorTable() error {
	// Verify sqlite-vec is loaded
	var sqliteVersion, vecVersion string
	err := s.db.Raw("SELECT sqlite_version(), vec_version()").Row().Scan(&sqliteVersion, &vecVersion)
	if err != nil {
		return errors.Wrapf(err, "sqlite-vec extension not properly loaded")
	}

	// Create virtual table for vectors
	createTableSQL := fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS document_vectors USING vec0(
			document_id TEXT PRIMARY KEY,
			embedding float[%d]
		);
	`, s.vecDim)

	if err := s.db.Exec(createTableSQL).Error; err != nil {
		return errors.Wrapf(err, "failed to create document_vectors table")
	}

	return nil
}

// Store implements Store.Store
func (s *SqliteStore) Store(ctx context.Context, knowledge *Knowledge) error {
	if len(knowledge.Documents) == 0 {
		return nil
	}

	// Begin transaction
	tx := s.db.WithContext(ctx)
	if err := tx.Transaction(func(tx *gorm.DB) error {
		if knowledge.ID == "" {
			knowledge.ID = uuid.NewString()
		}
		record := SqliteKnowledgeRecord{
			ID:        knowledge.ID,
			Metadata:  datatypes.NewJSONType(knowledge.Metadata),
			Documents: make([]*SqliteDocumentRecord, 0, len(knowledge.Documents)),
		}

		if err := tx.Save(&record).Error; err != nil {
			return errors.Wrapf(err, "failed to save knowledge record")
		}

		for _, item := range knowledge.Documents {
			if item.ID == "" {
				item.ID = uuid.NewString()
			}

			// Create or update knowledge record
			docRecord := SqliteDocumentRecord{
				ID:                item.ID,
				EmbeddingText:     item.EmbeddingText,
				Metadata:          datatypes.NewJSONType(item.Metadata),
				KnowledgeRecordID: knowledge.ID, // Set the foreign key
				Content:           datatypes.NewJSONType(item.Content),
			}

			// Use Save to create or update
			if err := tx.Save(&docRecord).Error; err != nil {
				return errors.Wrapf(err, "failed to save document record")
			}

			// Store embedding in vector table
			if len(item.Embeddings) > 0 {
				// Delete existing vector (if updating)
				if err := tx.Exec("DELETE FROM document_vectors WHERE document_id = ?", item.ID).Error; err != nil {
					return errors.Wrapf(err, "failed to delete existing vector")
				}

				// Serialize embedding
				serializedEmbedding, err := sqlite_vec.SerializeFloat32(item.Embeddings)
				if err != nil {
					return errors.Wrapf(err, "failed to serialize embedding")
				}

				// Insert new vector
				insertSQL := "INSERT INTO document_vectors (document_id, embedding) VALUES (?, ?)"
				if err := tx.Exec(insertSQL, item.ID, serializedEmbedding).Error; err != nil {
					return errors.Wrapf(err, "failed to insert knowledge vector")
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Search implements Store.Search
func (s *SqliteStore) Search(ctx context.Context, queryEmbedding []float32, limit int, allowedKnowledgeIds []string) ([]KnowledgeSearchResult, error) {
	var allowedDocumentIds []string

	// Get document IDs from allowed knowledge IDs
	if len(allowedKnowledgeIds) > 0 {
		if err := s.db.WithContext(ctx).
			Model(&SqliteDocumentRecord{}).
			Where("knowledge_record_id IN ?", allowedKnowledgeIds).
			Pluck("id", &allowedDocumentIds).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get document IDs from knowledge IDs")
		}

		// If no documents found for the allowed knowledge IDs, return empty results
		if len(allowedDocumentIds) == 0 {
			return []KnowledgeSearchResult{}, nil
		}
	}

	// Validate embedding
	if len(queryEmbedding) == 0 {
		return []KnowledgeSearchResult{}, nil
	}

	// Serialize query embedding
	serializedQuery, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to serialize query embedding")
	}

	// Perform vector similarity search to get knowledge IDs and distances
	var searchSQL string
	var args []interface{}

	if len(allowedDocumentIds) > 0 {
		// Filter by allowed document IDs if specified
		searchSQL = `
			SELECT document_id, distance
			FROM document_vectors
			WHERE embedding MATCH ? AND document_id IN ?
			ORDER BY distance
			LIMIT ?
		`
		args = []interface{}{serializedQuery, allowedDocumentIds, limit * 2}
	} else {
		// No filtering - search all documents
		searchSQL = `
			SELECT document_id, distance
			FROM document_vectors
			WHERE embedding MATCH ?
			ORDER BY distance
			LIMIT ?
		`
		args = []interface{}{serializedQuery, limit * 2}
	}

	rows, err := s.db.WithContext(ctx).Raw(searchSQL, args...).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute search query")
	}
	defer rows.Close()

	// Collect IDs and distances
	type searchResult struct {
		ID       string
		Distance float32
	}
	var searchResults []searchResult

	for rows.Next() {
		var result searchResult
		if err := rows.Scan(&result.ID, &result.Distance); err != nil {
			return nil, errors.Wrapf(err, "failed to scan result row")
		}
		searchResults = append(searchResults, result)
	}

	if len(searchResults) == 0 {
		return []KnowledgeSearchResult{}, nil
	}

	// Get knowledge records for the found IDs
	var ids []string
	distanceMap := make(map[string]float32)
	for _, result := range searchResults {
		ids = append(ids, result.ID)
		distanceMap[result.ID] = result.Distance
	}

	var records []SqliteDocumentRecord
	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to fetch knowledge records")
	}

	// Convert to search results
	var results []KnowledgeSearchResult
	for _, record := range records {
		// Parse metadata
		metadata := record.Metadata.Data()
		if metadata == nil {
			metadata = map[string]any{
				"knowledge_id": record.KnowledgeRecordID,
			}
		}

		distance := distanceMap[record.ID]
		results = append(results, KnowledgeSearchResult{
			Document: &Document{
				ID:            record.ID,
				Content:       record.Content.Data(),
				Metadata:      metadata,
				EmbeddingText: record.EmbeddingText,
			},
			Score: 1.0 - distance, // Convert distance to similarity score
		})
	}

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetKnowledgeById implements Store.GetKnowledgeById
func (s *SqliteStore) GetKnowledgeById(ctx context.Context, knowledgeId string) (*Knowledge, error) {
	var record SqliteKnowledgeRecord
	if err := s.db.WithContext(ctx).Preload("Documents").First(&record, "id = ?", knowledgeId).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to fetch knowledge records")
	}

	knowledge := &Knowledge{
		ID:        record.ID,
		Metadata:  record.Metadata.Data(),
		Documents: make([]*Document, 0, len(record.Documents)),
	}

	for _, document := range record.Documents {
		// Parse metadata
		metadata := document.Metadata.Data()
		if metadata == nil {
			metadata = map[string]any{
				"knowledge_id": record.ID,
			}
		}

		knowledge.Documents = append(knowledge.Documents, &Document{
			ID:            document.ID,
			Content:       document.Content.Data(),
			Metadata:      metadata,
			EmbeddingText: document.EmbeddingText,
		})
	}

	return knowledge, nil
}

// DeleteKnowledgeById implements Store.DeleteKnowledgeById
func (s *SqliteStore) DeleteKnowledgeById(ctx context.Context, knowledgeId string) error {
	// Begin transaction
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var documentIds []string
		if err := tx.Model(&SqliteDocumentRecord{}).Where("knowledge_record_id = ?", knowledgeId).Pluck("id", &documentIds).Error; err != nil {
			return errors.Wrapf(err, "failed to get knowledge record")
		}

		if len(documentIds) > 0 {
			// Delete from vector table
			if err := tx.Exec("DELETE FROM document_vectors WHERE document_id IN ?", documentIds).Error; err != nil {
				return errors.Wrapf(err, "failed to delete vectors")
			}

			// Delete from knowledge table
			if err := tx.Delete(&SqliteDocumentRecord{}, "id IN ?", documentIds).Error; err != nil {
				return errors.Wrapf(err, "failed to delete knowledge records")
			}
		}

		if err := tx.Delete(&SqliteKnowledgeRecord{}, "id = ?", knowledgeId).Error; err != nil {
			return errors.Wrapf(err, "failed to delete knowledge record")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Close implements Store.Close
func (s *SqliteStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
