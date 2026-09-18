package internal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	violationBuffer  *ViolationBuffer
	bufferRetryCount = 3
	bufferRetryDelay = time.Millisecond * 100
)

// ViolationBuffer provides async buffered violation recording
type ViolationBuffer struct {
	ch     chan *ViolationLog
	done   chan struct{}
	wg     sync.WaitGroup
	db     *sqlx.DB
	size   int
	errors chan error
}

// NewViolationBuffer creates a new violation buffer for async writes
func NewViolationBuffer(db *sqlx.DB, bufferSize int) *ViolationBuffer {
	return &ViolationBuffer{
		ch:     make(chan *ViolationLog, bufferSize),
		done:   make(chan struct{}),
		db:     db,
		size:   bufferSize,
		errors: make(chan error, bufferSize),
	}
}

// Start begins the async violation write worker
func (vb *ViolationBuffer) Start() {
	vb.wg.Add(1)
	go func() {
		defer vb.wg.Done()
		// Batch write violations every 100ms or when buffer reaches 500 items
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		batch := make([]*ViolationLog, 0, 500)

		for {
			select {
			case <-vb.done:
				// Flush remaining batch on shutdown
				if len(batch) > 0 {
					vb.batchInsert(batch)
				}
				return
			case v := <-vb.ch:
				if v == nil {
					return
				}
				batch = append(batch, v)
				if len(batch) >= 500 {
					vb.batchInsert(batch)
					batch = batch[:0]
				}
			case <-ticker.C:
				if len(batch) > 0 {
					vb.batchInsert(batch)
					batch = batch[:0]
				}
			}
		}
	}()
}

// Stop gracefully shuts down the violation buffer
func (vb *ViolationBuffer) Stop() error {
	close(vb.done)
	vb.wg.Wait()
	close(vb.ch)
	close(vb.errors)
	return nil
}

// batchInsert inserts multiple violations in a single transaction
func (vb *ViolationBuffer) batchInsert(violations []*ViolationLog) {
	if len(violations) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := vb.db.BeginTx(ctx, nil)
	if err != nil {
		select {
		case vb.errors <- fmt.Errorf("[violation] batch insert begin tx failed: %w", err):
		default:
		}
		return
	}

	stmt, err := tx.PrepareContext(ctx, sqlInsertViolation)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("[violation] rollback after prepare error failed: %v", rbErr)
		}
		select {
		case vb.errors <- fmt.Errorf("[violation] batch insert prepare failed: %w", err):
		default:
		}
		return
	}
	defer func() {
		if cerr := stmt.Close(); cerr != nil {
			log.Printf("[violation] close stmt error: %v", cerr)
		}
	}()

	for _, v := range violations {
		_, err := stmt.ExecContext(ctx, v.ID, v.Username, v.RemoteIP, v.Action, v.Reason, v.Timestamp)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("[violation] rollback after exec error failed: %v", rbErr)
			}
			select {
			case vb.errors <- fmt.Errorf("[violation] batch insert exec failed: %w", err):
			default:
			}
			return
		}
	}

	if err := tx.Commit(); err != nil {
		select {
		case vb.errors <- fmt.Errorf("[violation] batch insert commit failed: %w", err):
		default:
		}
	}
}

// Write sends a violation to the async buffer (non-blocking with retry)
func (vb *ViolationBuffer) Write(violation *ViolationLog) error {
	for attempt := 0; attempt < bufferRetryCount; attempt++ {
		select {
		case vb.ch <- violation:
			return nil
		default:
			if attempt < bufferRetryCount-1 {
				time.Sleep(bufferRetryDelay)
			}
		}
	}
	return fmt.Errorf("[violation] buffer full after %d attempts", bufferRetryCount)
}
