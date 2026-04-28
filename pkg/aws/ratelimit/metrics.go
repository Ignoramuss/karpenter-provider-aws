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
	opmetrics "github.com/awslabs/operatorpkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"sigs.k8s.io/karpenter/pkg/metrics"
)

const (
	subsystem      = "aws_sdk_rate_limit"
	serviceLabel   = "service"
	operationLabel = "operation"
)

var (
	rateLimitWaitDurationSeconds = opmetrics.NewPrometheusHistogram(crmetrics.Registry, prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Subsystem: subsystem,
		Name:      "wait_duration_seconds",
		Help:      "Client-side rate limiter latency in seconds before AWS SDK calls.",
		Buckets:   metrics.DurationBuckets(),
	}, []string{serviceLabel, operationLabel})
)
