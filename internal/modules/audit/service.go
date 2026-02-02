package audit

import (
	"log"
	"sync"
	"time"

	"github.com/maksroxx/flowkeeper/internal/config"
	"gorm.io/gorm"
)

type Service interface {
	Log(userID uint, action, entity string, entityID uint, details, ip string)
	GetLogs(limit, offset int) ([]AuditLog, int64, error)
	StartWorker()
	StopWorker()
}

type asyncService struct {
	db        *gorm.DB
	logChan   chan AuditLog
	batchSize int
	flushTime time.Duration
	wg        sync.WaitGroup
}

func NewAsyncService(db *gorm.DB, cfg config.AuditConfig) Service {
	size := cfg.BatchSize
	if size <= 0 {
		size = 50
	}

	seconds := cfg.FlushIntervalSeconds
	if seconds <= 0 {
		seconds = 2
	}

	return &asyncService{
		db:        db,
		logChan:   make(chan AuditLog, 1000),
		batchSize: size,
		flushTime: time.Duration(seconds) * time.Second,
	}
}

func (s *asyncService) Log(userID uint, action, entity string, entityID uint, details, ip string) {
	entry := AuditLog{
		UserID:    userID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Details:   details,
		IPAddress: ip,
		CreatedAt: time.Now(),
	}

	select {
	case s.logChan <- entry:
	default:
		log.Println("⚠️ Audit log channel full, dropping message")
	}
}

func (s *asyncService) GetLogs(limit, offset int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var count int64
	s.db.Model(&AuditLog{}).Count(&count)

	err := s.db.Preload("User").Preload("User.Role").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, count, err
}

func (s *asyncService) StartWorker() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		var batch []AuditLog
		ticker := time.NewTicker(s.flushTime)
		defer ticker.Stop()

		for {
			select {
			case entry, ok := <-s.logChan:
				if !ok {
					if len(batch) > 0 {
						s.saveBatch(batch)
					}
					return
				}

				batch = append(batch, entry)

				if len(batch) >= s.batchSize {
					s.saveBatch(batch)
					batch = nil
				}

			case <-ticker.C:
				if len(batch) > 0 {
					s.saveBatch(batch)
					batch = nil
				}
			}
		}
	}()
}

func (s *asyncService) StopWorker() {
	close(s.logChan)
	s.wg.Wait()
}

func (s *asyncService) saveBatch(logs []AuditLog) {
	if len(logs) == 0 {
		return
	}
	if err := s.db.CreateInBatches(logs, len(logs)).Error; err != nil {
		log.Printf("❌ Failed to save audit logs batch: %v", err)
	}
}
