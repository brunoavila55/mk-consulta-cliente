package mk

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// doWithRetry executa req com retentativas em falhas transitórias (erro de
// rede ou HTTP 5xx). Erros 4xx não são retentados, pois indicam um problema
// na própria requisição, não uma falha passageira do sistema MK.
func doWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxAttempts int) (*http.Response, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		res, err := client.Do(req.Clone(ctx))
		if err == nil && res.StatusCode < http.StatusInternalServerError {
			return res, nil
		}

		if err == nil {
			lastErr = fmt.Errorf("resposta HTTP %d", res.StatusCode)
			res.Body.Close()
		} else {
			lastErr = err
		}

		if attempt == maxAttempts {
			break
		}

		backoff := time.Duration(attempt) * 150 * time.Millisecond
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}

	return nil, lastErr
}
