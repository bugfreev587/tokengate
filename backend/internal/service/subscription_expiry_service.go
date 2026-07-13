package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// SubscriptionExpiryService periodically updates expired subscription status.
type SubscriptionExpiryService struct {
	userSubRepo       UserSubscriptionRepository
	byoAccountUpdater BYOAccountEntitlementUpdater
	interval          time.Duration
	stopCh            chan struct{}
	stopOnce          sync.Once
	wg                sync.WaitGroup
}

func NewSubscriptionExpiryService(userSubRepo UserSubscriptionRepository, interval time.Duration, byoAccountUpdater ...BYOAccountEntitlementUpdater) *SubscriptionExpiryService {
	var updater BYOAccountEntitlementUpdater
	if len(byoAccountUpdater) > 0 {
		updater = byoAccountUpdater[0]
	}
	return &SubscriptionExpiryService{
		userSubRepo:       userSubRepo,
		byoAccountUpdater: updater,
		interval:          interval,
		stopCh:            make(chan struct{}),
	}
}

func (s *SubscriptionExpiryService) Start() {
	if s == nil || s.userSubRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *SubscriptionExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *SubscriptionExpiryService) runOnce() {
	if s == nil || s.userSubRepo == nil {
		return
	}
	statusCtx, statusCancel := context.WithTimeout(context.Background(), 10*time.Second)
	updated, err := s.userSubRepo.BatchUpdateExpiredStatus(statusCtx)
	statusCancel()
	if err != nil {
		log.Printf("[SubscriptionExpiry] Update expired subscriptions failed: %v", err)
	} else if updated > 0 {
		log.Printf("[SubscriptionExpiry] Updated %d expired subscriptions", updated)
	}

	if s.byoAccountUpdater == nil {
		return
	}
	reconcileCtx, reconcileCancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = s.byoAccountUpdater.ReconcileBYOAccountEntitlements(reconcileCtx)
	reconcileCancel()
	if err != nil {
		log.Printf("[SubscriptionExpiry] Reconcile BYO entitlements failed: %v", err)
		return
	}
}
