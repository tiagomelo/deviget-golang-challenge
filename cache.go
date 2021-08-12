package sample1

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	prices             sync.Map
}

// PriceCacheEntry represents a cache entry
type PriceCacheEntry struct {
	price     float64
	timestamp time.Time
}

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	var priceCacheEntry PriceCacheEntry
	var castSuccessful bool
	now := time.Now().UTC()
	retrievedPriceCacheEntry, ok := c.prices.Load(itemCode)
	if ok {
		if priceCacheEntry, castSuccessful = retrievedPriceCacheEntry.(PriceCacheEntry); !castSuccessful {
			return 0, errors.New("error when casting")
		}
		// check if priceCacheEntry is still valid
		expDate := priceCacheEntry.timestamp.Add(c.maxAge)
		if now.Before(expDate) {
			return retrievedPriceCacheEntry.(PriceCacheEntry).price, nil
		}
		// priceCacheEntry is expired; time to evict it from cache
		c.prices.Delete(itemCode)
	}
	price, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}
	priceCacheEntry = PriceCacheEntry{
		price:     price,
		timestamp: now,
	}
	c.prices.Store(itemCode, priceCacheEntry)
	return priceCacheEntry.price, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	var g errgroup.Group
	var mu sync.Mutex
	var results []float64
	for _, itemCode := range itemCodes {
		ic := itemCode
		g.Go(func() error {
			price, err := c.GetPriceFor(ic)
			if err != nil {
				return err
			}
			mu.Lock()
			results = append(results, price)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return results, err
	}
	return results, nil
}
