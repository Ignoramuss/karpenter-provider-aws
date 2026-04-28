/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/smithy-go/middleware"
	"golang.org/x/time/rate"
)

// NewMiddleware returns an APIOptions function that installs a rate-limiting
// InitializeMiddleware into the AWS SDK middleware stack. The limiter blocks
// via rate.Limiter.Wait before each API call, respecting context cancellation.
func NewMiddleware(limiter *rate.Limiter) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Initialize.Add(
			middleware.InitializeMiddlewareFunc(
				"ClientSideRateLimiter",
				func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (
					middleware.InitializeOutput, middleware.Metadata, error,
				) {
					service := middleware.GetServiceID(ctx)
					operation := middleware.GetOperationName(ctx)
					labels := map[string]string{serviceLabel: service, operationLabel: operation}

					start := time.Now()
					if err := limiter.Wait(ctx); err != nil {
						return middleware.InitializeOutput{}, middleware.Metadata{},
							fmt.Errorf("client-side rate limit wait for %s.%s: %w", service, operation, err)
					}
					waited := time.Since(start)

					rateLimitWaitDurationSeconds.Observe(waited.Seconds(), labels)

					return next.HandleInitialize(ctx, in)
				},
			),
			middleware.After,
		)
	}
}
