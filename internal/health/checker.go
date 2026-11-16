package health

import (
	"context"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/Srinidhi444/Reverse-Proxy/internal/upstream"
)

type Checker struct {
	manager   *Manager
	client    *http.Client
	endpoints []*upstream.Endpoint

	interval   time.Duration
	timeout    time.Duration
	path       string
	maxWorkers int
}

func NewChecker(
	manager *Manager,
	endpoints []*upstream.Endpoint,
	interval time.Duration,
	timeout time.Duration,
	healthPath string,
	maxWorkers int,
) *Checker {
	if manager == nil {
		panic("health manager must not be nil")
	}
	if interval <= 0 {
		panic("health check interval must be positive")
	}
	if timeout <= 0 {
		panic("health check timeout must be positive")
	}
	if maxWorkers <= 0 {
		panic("maximum health check workers must be positive")
	}

	if healthPath == "" {
		healthPath = "/health"
	}

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	endpointCopy := append([]*upstream.Endpoint(nil), endpoints...)

	for _, endpoint := range endpointCopy {
		manager.Register(endpoint)
	}

	return &Checker{
		manager:    manager,
		client:     client,
		endpoints:  endpointCopy,
		interval:   interval,
		timeout:    timeout,
		path:       healthPath,
		maxWorkers: maxWorkers,
	}
}

func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Perform the initial health check immediately.
	c.checkAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			c.checkAll(ctx)
		}
	}
}

func (c *Checker) checkAll(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	jobs := make(chan *upstream.Endpoint)

	var workers sync.WaitGroup

	workerCount := c.maxWorkers
	if workerCount > len(c.endpoints) {
		workerCount = len(c.endpoints)
	}

	for i := 0; i < workerCount; i++ {
		workers.Add(1)

		go func() {
			defer workers.Done()

			for endpoint := range jobs {
				if ctx.Err() != nil {
					continue
				}

				c.checkEndpoint(ctx, endpoint)
			}
		}()
	}

sendLoop:
	for _, endpoint := range c.endpoints {
		select {
		case <-ctx.Done():
			break sendLoop
		case jobs <- endpoint:
		}
	}

	close(jobs)
	workers.Wait()
}

func (c *Checker) checkEndpoint(
	ctx context.Context,
	endpoint *upstream.Endpoint,
) {
	if endpoint == nil || endpoint.URL == nil {
		return
	}

	healthURL := *endpoint.URL

	healthURL.Path = path.Join(
		healthURL.Path,
		c.path,
	)

	healthURL.RawPath = ""

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		healthURL.String(),
		nil,
	)
	if err != nil {
		c.manager.RecordFailure(endpoint.ID)
		return
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.manager.RecordFailure(endpoint.ID)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK &&
		resp.StatusCode < http.StatusMultipleChoices {
		c.manager.RecordSuccess(endpoint.ID)
	} else {
		c.manager.RecordFailure(endpoint.ID)
	}
}
