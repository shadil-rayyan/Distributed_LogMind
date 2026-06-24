package ingestion

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"logmind/internal/config"
	"logmind/internal/detection"
	"logmind/internal/domain"
	"logmind/internal/observability"
	"logmind/internal/storage"
)

var batchPool = sync.Pool{
	New: func() interface{} {
		// Initialize with a default slice. It will grow as needed.
		slice := make([]domain.Log, 0, 1000)
		return &slice
	},
}

type WorkerPool struct {
	cfg        *config.Config
	repo       storage.LogRepository
	engine     *detection.Engine
	logChannel <-chan domain.Log
	wg         sync.WaitGroup
}

func NewWorkerPool(cfg *config.Config, repo storage.LogRepository, engine *detection.Engine, logChannel <-chan domain.Log) *WorkerPool {
	return &WorkerPool{
		cfg:        cfg,
		repo:       repo,
		engine:     engine,
		logChannel: logChannel,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.cfg.MaxWorkerPool; i++ {
		wp.wg.Add(1)
		go wp.batchWorker(i)
	}
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func (wp *WorkerPool) batchWorker(workerID int) {
	defer wp.wg.Done()

	batchPtr := batchPool.Get().(*[]domain.Log)
	// Reset length to 0 to reuse the allocated capacity
	*batchPtr = (*batchPtr)[:0]
	
	ticker := time.NewTicker(wp.cfg.BatchTimeout)
	defer ticker.Stop()

	flush := func() {
		if len(*batchPtr) == 0 {
			return
		}
		
		start := time.Now()
		
		var err error
		// Issue #2: Implement retry logic with exponential backoff
		retries := 3
		backoff := 50 * time.Millisecond
		for i := 0; i < retries; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = wp.repo.InsertBatch(ctx, *batchPtr)
			cancel() // Issue #5: Immediate cancel to avoid leaking context inside loops/repeated calls
			
			if err == nil {
				break
			}
			slog.Warn("failed to flush batch to db, retrying...", "error", err, "workerID", workerID, "attempt", i+1)
			time.Sleep(backoff)
			backoff *= 2
		}

		if err != nil {
			slog.Error("failed to flush batch to db after retries, data dropped", "error", err, "workerID", workerID)
		} else {
			// Record metrics after successful write
			for _, l := range *batchPtr {
				if l.Level == "error" {
					wp.engine.RecordError(l.Service, l.Timestamp)
				}
			}
			duration := time.Since(start).Seconds()
			observability.BatchFlushDuration.Observe(duration)
			slog.Debug("batch flushed", "worker_id", workerID, "batch_size", len(*batchPtr), "duration_s", duration)
		}

		// Issue #3: Always decrement queue depth for all items processed, regardless of success/failure
		for range *batchPtr {
			observability.QueueDepth.Dec()
		}
		
		*batchPtr = (*batchPtr)[:0]
	}

	for {
		select {
		case logData, ok := <-wp.logChannel:
			if !ok {
				flush() // Channel closed, drain remaining
				// Put back into pool when done (Issue #16: return pointer without copy)
				batchPool.Put(batchPtr)
				return
			}
			*batchPtr = append(*batchPtr, logData)
			if len(*batchPtr) >= wp.cfg.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush() // Timeout reached
		}
	}
}

