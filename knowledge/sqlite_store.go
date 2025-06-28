package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SqliteStore implements Store using SQLite with sqlite-vec extension
type SqliteStore struct {
	db     *gorm.DB
	vecDim int
}

// knowledgeRecord represents the database structure for knowledge items
type knowledgeRecord struct {
	ID        string `gorm:"primaryKey"`
	AgentName string `gorm:"index"`
	Content   string `gorm:"type:text"`
	Metadata  string `gorm:"type:text"` // JSON string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for GORM
func (knowledgeRecord) TableName() string {
	return "knowledge"
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
	if err := db.AutoMigrate(&knowledgeRecord{}); err != nil {
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
		CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_vectors USING vec0(
			knowledge_id TEXT PRIMARY KEY,
			embedding float[%d]
		);
	`, s.vecDim)

	if err := s.db.Exec(createTableSQL).Error; err != nil {
		return errors.Wrapf(err, "failed to create knowledge_vectors table")
	}

	return nil
}

// Store implements Store.Store
func (s *SqliteStore) Store(ctx context.Context, items []KnowledgeItem) error {
	if len(items) == 0 {
		return nil
	}

	// Begin transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	for _, item := range items {
		// Generate ID if not provided
		if item.ID == "" {
			item.ID = uuid.New().String()
		}

		// Serialize metadata
		metadataJSON, err := json.Marshal(item.Metadata)
		if err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to serialize metadata")
		}

		// Create or update knowledge record
		record := knowledgeRecord{
			ID:        item.ID,
			AgentName: item.AgentName,
			Content:   item.Content,
			Metadata:  string(metadataJSON),
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Use Save to create or update
		if err := tx.Save(&record).Error; err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to save knowledge record")
		}

		// Store embedding in vector table
		if len(item.Embedding) > 0 {
			// Delete existing vector (if updating)
			if err := tx.Exec("DELETE FROM knowledge_vectors WHERE knowledge_id = ?", item.ID).Error; err != nil {
				tx.Rollback()
				return errors.Wrapf(err, "failed to delete existing vector")
			}

			// Serialize embedding
			serializedEmbedding, err := sqlite_vec.SerializeFloat32(item.Embedding)
			if err != nil {
				tx.Rollback()
				return errors.Wrapf(err, "failed to serialize embedding")
			}

			// Insert new vector
			insertSQL := `INSERT INTO knowledge_vectors (knowledge_id, embedding) VALUES (?, ?)`
			if err := tx.Exec(insertSQL, item.ID, serializedEmbedding).Error; err != nil {
				tx.Rollback()
				return errors.Wrapf(err, "failed to insert knowledge vector")
			}
		}
	}

	return tx.Commit().Error
}

// Search implements Store.Search
func (s *SqliteStore) Search(ctx context.Context, agentName string, queryEmbedding []float32, limit int) ([]KnowledgeSearchResult, error) {
	// Serialize query embedding
	serializedQuery, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to serialize query embedding")
	}

	// Perform vector similarity search to get knowledge IDs and distances
	searchSQL := `
		SELECT knowledge_id, distance
		FROM knowledge_vectors
		WHERE embedding MATCH ?
		ORDER BY distance
		LIMIT ?
	`

	rows, err := s.db.WithContext(ctx).Raw(searchSQL, serializedQuery, limit*2).Rows() // Get more results for filtering
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

	var records []knowledgeRecord
	if err := s.db.WithContext(ctx).Where("id IN ? AND agent_name = ?", ids, agentName).Find(&records).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to fetch knowledge records")
	}

	// Convert to search results
	var results []KnowledgeSearchResult
	for _, record := range records {
		// Parse metadata
		var metadata map[string]interface{}
		if record.Metadata != "" {
			if err := json.Unmarshal([]byte(record.Metadata), &metadata); err != nil {
				return nil, errors.Wrapf(err, "failed to parse metadata")
			}
		}

		distance := distanceMap[record.ID]
		results = append(results, KnowledgeSearchResult{
			KnowledgeItem: KnowledgeItem{
				ID:        record.ID,
				AgentName: record.AgentName,
				Content:   record.Content,
				Metadata:  metadata,
				CreatedAt: record.CreatedAt,
				UpdatedAt: record.UpdatedAt,
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

// GetByAgent implements Store.GetByAgent
func (s *SqliteStore) GetByAgent(ctx context.Context, agentName string) ([]KnowledgeItem, error) {
	var records []knowledgeRecord
	if err := s.db.WithContext(ctx).Where("agent_name = ?", agentName).Find(&records).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to fetch knowledge records")
	}

	items := make([]KnowledgeItem, 0, len(records))
	for _, record := range records {
		// Parse metadata
		var metadata map[string]interface{}
		if record.Metadata != "" {
			if err := json.Unmarshal([]byte(record.Metadata), &metadata); err != nil {
				return nil, errors.Wrapf(err, "failed to parse metadata")
			}
		}

		items = append(items, KnowledgeItem{
			ID:        record.ID,
			AgentName: record.AgentName,
			Content:   record.Content,
			Metadata:  metadata,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
			// Note: Embedding is not loaded here for performance
		})
	}

	return items, nil
}

// DeleteByAgent implements Store.DeleteByAgent
func (s *SqliteStore) DeleteByAgent(ctx context.Context, agentName string) error {
	// Begin transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get IDs of knowledge to delete
	var ids []string
	if err := tx.Model(&knowledgeRecord{}).Where("agent_name = ?", agentName).Pluck("id", &ids).Error; err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to get knowledge IDs")
	}

	if len(ids) > 0 {
		// Delete from vector table
		if err := tx.Exec("DELETE FROM knowledge_vectors WHERE knowledge_id IN ?", ids).Error; err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to delete vectors")
		}

		// Delete from knowledge table
		if err := tx.Where("agent_name = ?", agentName).Delete(&knowledgeRecord{}).Error; err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "failed to delete knowledge records")
		}
	}

	return tx.Commit().Error
}

// Close implements Store.Close
func (s *SqliteStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
