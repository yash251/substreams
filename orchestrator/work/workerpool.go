package work

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/streamingfast/substreams/reqctx"
)

type WorkerPool struct {
	workers []*WorkerStatus
}

type WorkerState int

const (
	WorkerUnset WorkerState = iota
	WorkerFree
	WorkerWorking
)

type WorkerStatus struct {
	State  WorkerState
	Worker Worker
}

func NewWorkerPool(ctx context.Context, initialWorkers, targetWorkers int, rampupPeriod time.Duration, workerFactory WorkerFactory) *WorkerPool {
	logger := reqctx.Logger(ctx)

	if initialWorkers > targetWorkers || rampupPeriod == 0 {
		initialWorkers = targetWorkers
	}

	logger.Info("initializing worker pool",
		zap.Int("target_count", targetWorkers),
		zap.Int("initial_workers", initialWorkers),
		zap.Duration("rampup_period", rampupPeriod),
	)
	workers := initWorkers(initialWorkers, targetWorkers, workerFactory, logger)

	if targetWorkers > initialWorkers {
		go rampupWorkers(targetWorkers, rampupPeriod, workers)
	}

	return &WorkerPool{
		workers: workers,
	}
}

func initWorkers(initialWorkers, targetWorkers int, workerFactory WorkerFactory, logger *zap.Logger) []*WorkerStatus {
	workers := make([]*WorkerStatus, targetWorkers)
	for i := 0; i < targetWorkers; i++ {
		workers[i] = &WorkerStatus{
			Worker: workerFactory(logger),
		}
		if i < initialWorkers {
			workers[i].State = WorkerFree
		}
	}
	return workers
}

// this particular function is thread-safe, no need for lock because nobody else ever touches workers with state=Unset, we just enable them for future use
func rampupWorkers(target int, rampup time.Duration, workers []*WorkerStatus) {
	begin := time.Now()
	rampupFloat := float32(rampup)
	targetFloat := float32(target)

	var currentlySet int
	for _, worker := range workers {
		if worker.State > WorkerUnset {
			currentlySet++
		}
	}

	for currentlySet < target {
		time.Sleep(time.Second)
		ratio := float32(time.Since(begin)) / rampupFloat
		currentTarget := int(targetFloat * ratio)
		if currentTarget > target {
			currentTarget = target
		}
		if currentTarget > currentlySet {
			for _, worker := range workers {
				if worker.State == WorkerUnset {
					worker.State = WorkerFree
					currentlySet++
					if currentTarget == currentlySet {
						break
					}
				}
			}
		}
	}
}

func (p *WorkerPool) WorkerAvailable() bool {
	for _, w := range p.workers {
		if w.State == WorkerFree {
			return true
		}
	}
	return false
}

func (p *WorkerPool) Borrow() Worker {
	for _, status := range p.workers {
		if status.State == WorkerFree {
			status.State = WorkerWorking
			return status.Worker
		}
	}
	panic("no free workers, call WorkerAvailable() first")
}

func (p *WorkerPool) Return(worker Worker) {
	for _, status := range p.workers {
		if status.Worker == worker {
			if status.State != WorkerWorking {
				panic("returned worker was already free")
			}
			status.State = WorkerFree
			return
		}
	}
}
