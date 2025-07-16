package db

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	DefaultBatchSize = 1000
	DefaultWorkers   = 4
)

// BatchInserter handles batch insert operations
type BatchInserter struct {
	db        *sqlx.DB
	batchSize int
	workers   int
	mu        sync.RWMutex
}

// NewBatchInserter creates a new batch inserter
func NewBatchInserter(db *sqlx.DB, batchSize, workers int) *BatchInserter {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	if workers <= 0 {
		workers = DefaultWorkers
	}
	
	return &BatchInserter{
		db:        db,
		batchSize: batchSize,
		workers:   workers,
	}
}

// BatchRecord represents a record to be inserted
type BatchRecord struct {
	Source      string
	OrderID     string
	Date        time.Time
	TotalAmount float64
	RawData     string
}

// BatchInsertRecords inserts records in batches using multiple workers
func (bi *BatchInserter) BatchInsertRecords(records []BatchRecord) error {
	if len(records) == 0 {
		return nil
	}

	// Create batches
	batches := bi.createBatches(records)
	
	// Channel to receive batches
	batchChan := make(chan []BatchRecord, len(batches))
	errorChan := make(chan error, bi.workers)
	
	// Send batches to channel
	for _, batch := range batches {
		batchChan <- batch
	}
	close(batchChan)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < bi.workers; i++ {
		wg.Add(1)
		go bi.worker(&wg, batchChan, errorChan)
	}
	
	// Wait for all workers to complete
	wg.Wait()
	close(errorChan)
	
	// Check for errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}
	
	return nil
}

// createBatches splits records into batches
func (bi *BatchInserter) createBatches(records []BatchRecord) [][]BatchRecord {
	var batches [][]BatchRecord
	
	for i := 0; i < len(records); i += bi.batchSize {
		end := i + bi.batchSize
		if end > len(records) {
			end = len(records)
		}
		batches = append(batches, records[i:end])
	}
	
	return batches
}

// worker processes batches of records
func (bi *BatchInserter) worker(wg *sync.WaitGroup, batchChan <-chan []BatchRecord, errorChan chan<- error) {
	defer wg.Done()
	
	for batch := range batchChan {
		if err := bi.insertBatch(batch); err != nil {
			errorChan <- err
			return
		}
	}
}

// insertBatch inserts a single batch of records
func (bi *BatchInserter) insertBatch(batch []BatchRecord) error {
	if len(batch) == 0 {
		return nil
	}
	
	// Build the SQL query for bulk insert
	query := "INSERT INTO records (source, order_id, date, total_amount, raw_data) VALUES "
	values := make([]interface{}, 0, len(batch)*5)
	placeholders := make([]string, 0, len(batch))
	
	for i, record := range batch {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", 
			i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		values = append(values, record.Source, record.OrderID, record.Date, record.TotalAmount, record.RawData)
	}
	
	query += strings.Join(placeholders, ", ")
	
	// Execute the batch insert
	_, err := bi.db.Exec(query, values...)
	return err
}

// PreparedBatchInserter uses prepared statements for better performance
type PreparedBatchInserter struct {
	db        *sqlx.DB
	stmt      *sqlx.Stmt
	batchSize int
	workers   int
}

// NewPreparedBatchInserter creates a new prepared batch inserter
func NewPreparedBatchInserter(db *sqlx.DB, batchSize, workers int) (*PreparedBatchInserter, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	if workers <= 0 {
		workers = DefaultWorkers
	}
	
	// Prepare the statement
	stmt, err := db.Preparex("INSERT INTO records (source, order_id, date, total_amount, raw_data) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return nil, err
	}
	
	return &PreparedBatchInserter{
		db:        db,
		stmt:      stmt,
		batchSize: batchSize,
		workers:   workers,
	}, nil
}

// Close closes the prepared statement
func (pbi *PreparedBatchInserter) Close() error {
	return pbi.stmt.Close()
}

// BatchInsertRecords inserts records using prepared statements
func (pbi *PreparedBatchInserter) BatchInsertRecords(records []BatchRecord) error {
	if len(records) == 0 {
		return nil
	}
	
	// Create batches
	batches := pbi.createBatches(records)
	
	// Channel to receive batches
	batchChan := make(chan []BatchRecord, len(batches))
	errorChan := make(chan error, pbi.workers)
	
	// Send batches to channel
	for _, batch := range batches {
		batchChan <- batch
	}
	close(batchChan)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < pbi.workers; i++ {
		wg.Add(1)
		go pbi.worker(&wg, batchChan, errorChan)
	}
	
	// Wait for all workers to complete
	wg.Wait()
	close(errorChan)
	
	// Check for errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}
	
	return nil
}

// createBatches splits records into batches
func (pbi *PreparedBatchInserter) createBatches(records []BatchRecord) [][]BatchRecord {
	var batches [][]BatchRecord
	
	for i := 0; i < len(records); i += pbi.batchSize {
		end := i + pbi.batchSize
		if end > len(records) {
			end = len(records)
		}
		batches = append(batches, records[i:end])
	}
	
	return batches
}

// worker processes batches using prepared statements
func (pbi *PreparedBatchInserter) worker(wg *sync.WaitGroup, batchChan <-chan []BatchRecord, errorChan chan<- error) {
	defer wg.Done()
	
	for batch := range batchChan {
		if err := pbi.insertBatch(batch); err != nil {
			errorChan <- err
			return
		}
	}
}

// insertBatch inserts a batch using prepared statements
func (pbi *PreparedBatchInserter) insertBatch(batch []BatchRecord) error {
	if len(batch) == 0 {
		return nil
	}
	
	// Begin transaction for this batch
	tx, err := pbi.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Use prepared statement within transaction
	txStmt := tx.Stmtx(pbi.stmt)
	
	// Insert each record in the batch
	for _, record := range batch {
		_, err := txStmt.Exec(record.Source, record.OrderID, record.Date, record.TotalAmount, record.RawData)
		if err != nil {
			return err
		}
	}
	
	// Commit the transaction
	return tx.Commit()
}

// StreamingBatchInserter handles streaming batch inserts from a channel
type StreamingBatchInserter struct {
	db        *sqlx.DB
	batchSize int
	workers   int
	stmt      *sqlx.Stmt
}

// NewStreamingBatchInserter creates a new streaming batch inserter
func NewStreamingBatchInserter(db *sqlx.DB, batchSize, workers int) (*StreamingBatchInserter, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	if workers <= 0 {
		workers = DefaultWorkers
	}
	
	// Prepare the statement
	stmt, err := db.Preparex("INSERT INTO records (source, order_id, date, total_amount, raw_data) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return nil, err
	}
	
	return &StreamingBatchInserter{
		db:        db,
		batchSize: batchSize,
		workers:   workers,
		stmt:      stmt,
	}, nil
}

// Close closes the prepared statement
func (sbi *StreamingBatchInserter) Close() error {
	return sbi.stmt.Close()
}

// StreamInsertRecords processes records from a channel and inserts them in batches
func (sbi *StreamingBatchInserter) StreamInsertRecords(recordChan <-chan BatchRecord) error {
	// Buffer for collecting records into batches
	batch := make([]BatchRecord, 0, sbi.batchSize)
	
	// Process records from channel
	for record := range recordChan {
		batch = append(batch, record)
		
		// When batch is full, insert it
		if len(batch) >= sbi.batchSize {
			if err := sbi.insertBatch(batch); err != nil {
				return err
			}
			batch = batch[:0] // Reset slice
		}
	}
	
	// Insert any remaining records
	if len(batch) > 0 {
		if err := sbi.insertBatch(batch); err != nil {
			return err
		}
	}
	
	return nil
}

// insertBatch inserts a batch using prepared statements
func (sbi *StreamingBatchInserter) insertBatch(batch []BatchRecord) error {
	if len(batch) == 0 {
		return nil
	}
	
	// Begin transaction for this batch
	tx, err := sbi.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Use prepared statement within transaction
	txStmt := tx.Stmtx(sbi.stmt)
	
	// Insert each record in the batch
	for _, record := range batch {
		_, err := txStmt.Exec(record.Source, record.OrderID, record.Date, record.TotalAmount, record.RawData)
		if err != nil {
			return err
		}
	}
	
	// Commit the transaction
	return tx.Commit()
}
