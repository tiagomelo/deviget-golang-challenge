package sample1

import (
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

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	price, ok := c.prices.Load(itemCode)
	if ok {
		// TODO: check that the price was retrieved less than "maxAge" ago!
		return price.(float64), nil
	}
	price, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}
	c.prices.Store(itemCode, price)
	return price.(float64), nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	var g errgroup.Group
	var mu sync.Mutex
	results := []float64{}
	for _, itemCode := range itemCodes {
		ic := itemCode
		g.Go(func() error {
			if price, ok := c.prices.Load(ic); ok {
				mu.Lock()
				results = append(results, price.(float64))
				mu.Unlock()
				return nil
			}
			price, err := c.GetPriceFor(ic)
			if err != nil {
				return err
			}
			c.prices.Store(ic, price)
			mu.Lock()
			results = append(results, price)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return []float64{}, err
	}
	return results, nil
}
