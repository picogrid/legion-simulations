package core

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/picogrid/legion-simulations/pkg/client"
	"github.com/picogrid/legion-simulations/pkg/logger"
	"github.com/picogrid/legion-simulations/pkg/models"
)

// UpdateBuffer manages batched updates to Legion API
type UpdateBuffer struct {
	client        *client.Legion
	orgID         string
	updates       map[uuid.UUID]*EntityUpdate
	maxBatchSize  int
	flushInterval time.Duration
	lastFlush     time.Time
	mu            sync.Mutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// EntityUpdate represents a pending update for an entity
type EntityUpdate struct {
	EntityID     uuid.UUID
	Position     *models.GeomPoint
	Status       *string
	Metadata     map[string]interface{}
	LastModified time.Time
}

// UpdateStats tracks update statistics
type UpdateStats struct {
	TotalUpdates     int64
	BatchesSent      int64
	UpdatesSent      int64
	UpdatesFailed    int64
	AverageBatchSize float64
	LastBatchTime    time.Time
	LastError        error
}

// NewUpdateBuffer creates a new update buffer
func NewUpdateBuffer(client *client.Legion, orgID string, maxBatchSize int, flushInterval time.Duration) *UpdateBuffer {
	return &UpdateBuffer{
		client:        client,
		orgID:         orgID,
		updates:       make(map[uuid.UUID]*EntityUpdate),
		maxBatchSize:  maxBatchSize,
		flushInterval: flushInterval,
		lastFlush:     time.Now(),
		stopChan:      make(chan struct{}),
	}
}

// Start begins the automatic flush goroutine
func (ub *UpdateBuffer) Start(ctx context.Context) {
	ub.wg.Add(1)
	go func() {
		defer ub.wg.Done()

		ticker := time.NewTicker(ub.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ub.stopChan:
				return
			case <-ticker.C:
				if err := ub.Flush(ctx); err != nil {
					logger.Errorf("Error flushing updates: %v", err)
				}
			}
		}
	}()
}

// Stop stops the update buffer
func (ub *UpdateBuffer) Stop() {
	close(ub.stopChan)
	ub.wg.Wait()
}

// QueuePositionUpdate queues a position update
func (ub *UpdateBuffer) QueuePositionUpdate(entityID uuid.UUID, position *models.GeomPoint) {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	update, exists := ub.updates[entityID]
	if !exists {
		update = &EntityUpdate{
			EntityID: entityID,
			Metadata: make(map[string]interface{}),
		}
		ub.updates[entityID] = update
	}

	update.Position = position
	update.LastModified = time.Now()

	// Check if we should flush
	if len(ub.updates) >= ub.maxBatchSize {
		go func() {
			ctx := context.Background()
			if err := ub.Flush(ctx); err != nil {
				logger.Errorf("Error auto-flushing updates: %v", err)
			}
		}()
	}
}

// QueueStatusUpdate queues a status update
func (ub *UpdateBuffer) QueueStatusUpdate(entityID uuid.UUID, status string) {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	update, exists := ub.updates[entityID]
	if !exists {
		update = &EntityUpdate{
			EntityID: entityID,
			Metadata: make(map[string]interface{}),
		}
		ub.updates[entityID] = update
	}

	update.Status = &status
	update.LastModified = time.Now()
}

// QueueMetadataUpdate queues a metadata update
func (ub *UpdateBuffer) QueueMetadataUpdate(entityID uuid.UUID, key string, value interface{}) {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	update, exists := ub.updates[entityID]
	if !exists {
		update = &EntityUpdate{
			EntityID: entityID,
			Metadata: make(map[string]interface{}),
		}
		ub.updates[entityID] = update
	}

	update.Metadata[key] = value
	update.LastModified = time.Now()
}

// Flush sends all pending updates to Legion
func (ub *UpdateBuffer) Flush(ctx context.Context) error {
	ub.mu.Lock()

	if len(ub.updates) == 0 {
		ub.mu.Unlock()
		return nil
	}

	// Copy updates and clear buffer
	updates := make(map[uuid.UUID]*EntityUpdate)
	for k, v := range ub.updates {
		updates[k] = v
	}
	ub.updates = make(map[uuid.UUID]*EntityUpdate)
	ub.lastFlush = time.Now()

	ub.mu.Unlock()

	// Process updates
	var wg sync.WaitGroup
	errChan := make(chan error, len(updates))

	// Limit concurrent API calls
	semaphore := make(chan struct{}, 10)

	for entityID, update := range updates {
		wg.Add(1)
		go func(id uuid.UUID, u *EntityUpdate) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := ub.sendUpdate(ctx, id, u); err != nil {
				errChan <- err

				// Re-queue failed update
				ub.mu.Lock()
				ub.updates[id] = u
				ub.mu.Unlock()
			}
		}(entityID, update)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		logger.Errorf("Failed to send %d/%d updates", len(errors), len(updates))
		return errors[0] // Return first error
	}

	logger.Infof("Successfully flushed %d updates", len(updates))
	return nil
}

// sendUpdate sends a single update to Legion
func (ub *UpdateBuffer) sendUpdate(ctx context.Context, entityID uuid.UUID, update *EntityUpdate) error {
	// Update position if changed
	if update.Position != nil {
		req := &models.CreateEntityLocationRequest{
			Position: update.Position,
		}

		orgCtx := client.WithOrgID(ctx, ub.orgID)
		if _, err := ub.client.CreateEntityLocation(orgCtx, entityID.String(), req); err != nil {
			return err
		}
	}

	// Update status if changed
	if update.Status != nil {
		// TODO: Implement status update when API supports it
		// For now, we'll include it in metadata
		update.Metadata["status"] = *update.Status
	}

	// Update metadata if changed
	if len(update.Metadata) > 0 {
		// TODO: Implement metadata update when API supports partial updates
		// For now, log the metadata changes
		logger.Debugf("Metadata updates for entity %s: %v", entityID, update.Metadata)
	}

	return nil
}

// GetStats returns current buffer statistics
func (ub *UpdateBuffer) GetStats() UpdateStats {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	return UpdateStats{
		TotalUpdates:  int64(len(ub.updates)),
		LastBatchTime: ub.lastFlush,
	}
}

// ForceFlush immediately flushes all pending updates
func (ub *UpdateBuffer) ForceFlush(ctx context.Context) error {
	return ub.Flush(ctx)
}

// GetPendingCount returns the number of pending updates
func (ub *UpdateBuffer) GetPendingCount() int {
	ub.mu.Lock()
	defer ub.mu.Unlock()
	return len(ub.updates)
}
