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

package ratelimit_test

import (
	"context"
	"testing"
	"time"

	smithymiddleware "github.com/aws/smithy-go/middleware"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/time/rate"

	"github.com/aws/karpenter-provider-aws/pkg/aws/ratelimit"
)

func TestRateLimit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RateLimit")
}

func newRequest() interface{} { return struct{}{} }

type noopHandler struct{}

func (noopHandler) Handle(_ context.Context, _ interface{}) (interface{}, smithymiddleware.Metadata, error) {
	return nil, smithymiddleware.Metadata{}, nil
}

var _ = Describe("Middleware", func() {
	It("should register ClientSideRateLimiter in the Initialize step", func() {
		limiter := rate.NewLimiter(rate.Limit(10), 1)
		mw := ratelimit.NewMiddleware(limiter)

		stack := smithymiddleware.NewStack("test", newRequest)
		Expect(mw(stack)).To(Succeed())

		found := false
		for _, id := range stack.Initialize.List() {
			if id == "ClientSideRateLimiter" {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	})

	It("should respect context cancellation", func() {
		limiter := rate.NewLimiter(rate.Limit(1.0/60), 1)
		limiter.Allow()

		mw := ratelimit.NewMiddleware(limiter)
		stack := smithymiddleware.NewStack("test", newRequest)
		Expect(mw(stack)).To(Succeed())

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, _, err := stack.Initialize.HandleMiddleware(ctx, struct{}{}, noopHandler{})
		Expect(err).To(HaveOccurred())
	})

	It("should consume tokens from the rate limiter", func() {
		limiter := rate.NewLimiter(rate.Limit(1), 3)
		mw := ratelimit.NewMiddleware(limiter)

		stack := smithymiddleware.NewStack("test", newRequest)
		Expect(mw(stack)).To(Succeed())

		for i := 0; i < 3; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			_, _, err := stack.Initialize.HandleMiddleware(ctx, struct{}{}, noopHandler{})
			cancel()
			Expect(err).ToNot(HaveOccurred())
		}

		Expect(limiter.Allow()).To(BeFalse())
	})

	It("should block when tokens are exhausted", func() {
		limiter := rate.NewLimiter(rate.Limit(1), 1)
		limiter.Allow()

		mw := ratelimit.NewMiddleware(limiter)
		stack := smithymiddleware.NewStack("test", newRequest)
		Expect(mw(stack)).To(Succeed())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, _, err := stack.Initialize.HandleMiddleware(ctx, struct{}{}, noopHandler{})
		Expect(err).To(HaveOccurred())
	})

	It("should pass through with high burst", func() {
		limiter := rate.NewLimiter(rate.Limit(10000), 10000)
		mw := ratelimit.NewMiddleware(limiter)

		stack := smithymiddleware.NewStack("test", newRequest)
		Expect(mw(stack)).To(Succeed())

		for i := 0; i < 100; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			_, _, err := stack.Initialize.HandleMiddleware(ctx, struct{}{}, noopHandler{})
			cancel()
			Expect(err).ToNot(HaveOccurred())
		}
	})
})
