/*
Copyright 2020 Catalogic Software

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

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Metrics contains Prometheus metrics for KubeDR.
type MetricsInfo struct {
	metrics map[string]prometheus.Collector
}

const (
	backupSizeBytesKey       = "kubedr_backup_size_bytes"
	numBackupsKey            = "kubedr_num_backups"
	numSuccessfulBackupsKey  = "kubedr_num_successful_backups"
	numFailedBackupsKey      = "kubedr_num_failed_backups"
	backupDurationSecondsKey = "kubedr_backup_duration_seconds"

	policyLabel = "policyName"
)

// NewMetricsInfo creates a new metrics structure to be used by controllers.
func NewMetricsInfo() *MetricsInfo {
	return &MetricsInfo{
		metrics: map[string]prometheus.Collector{
			backupSizeBytesKey: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: backupSizeBytesKey,
					Help: "Size of a backup in bytes",
				},
				[]string{policyLabel},
			),

			numBackupsKey: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: numBackupsKey,
					Help: "Total number of backups",
				},
				[]string{policyLabel},
			),

			numSuccessfulBackupsKey: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: numSuccessfulBackupsKey,
					Help: "Total number of successful backups",
				},
				[]string{policyLabel},
			),

			numFailedBackupsKey: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: numFailedBackupsKey,
					Help: "Total number of failed backups",
				},
				[]string{policyLabel},
			),

			backupDurationSecondsKey: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name: backupDurationSecondsKey,
					Help: "Time taken to complete backup, in seconds",
					Buckets: []float64{
						15.0,
						30.0,
						toSeconds(1 * time.Minute),
						toSeconds(5 * time.Minute),
						toSeconds(10 * time.Minute),
						toSeconds(15 * time.Minute),
						toSeconds(30 * time.Minute),
						toSeconds(1 * time.Hour),
						toSeconds(2 * time.Hour),
						toSeconds(3 * time.Hour),
						toSeconds(4 * time.Hour),
						toSeconds(5 * time.Hour),
						toSeconds(6 * time.Hour),
						toSeconds(7 * time.Hour),
						toSeconds(8 * time.Hour),
						toSeconds(9 * time.Hour),
						toSeconds(10 * time.Hour),
					},
				},
				[]string{policyLabel},
			),
		},
	}
}

// RegisterAllMetrics registers all prometheus metrics.
func (m *MetricsInfo) RegisterAllMetrics() {
	for _, pm := range m.metrics {
		crmetrics.Registry.MustRegister(pm)
	}
}

// SetBackupSizeBytes records the size of a backup.
func (m *MetricsInfo) SetBackupSizeBytes(policy string, size uint64) {
	if pm, ok := m.metrics[backupSizeBytesKey].(*prometheus.GaugeVec); ok {
		pm.WithLabelValues(policy).Set(float64(size))
	}
}

// RecordBackup updates the total number of backups.
func (m *MetricsInfo) RecordBackup(policy string) {
	if pm, ok := m.metrics[numBackupsKey].(*prometheus.CounterVec); ok {
		pm.WithLabelValues(policy).Inc()
	}
}

// RecordSuccessfulBackup updates the total number of successful backups.
func (m *MetricsInfo) RecordSuccessfulBackup(policy string) {
	if pm, ok := m.metrics[numSuccessfulBackupsKey].(*prometheus.CounterVec); ok {
		pm.WithLabelValues(policy).Inc()
	}
}

// RecordFailedBackup updates the total number of failed backups.
func (m *MetricsInfo) RecordFailedBackup(policy string) {
	if pm, ok := m.metrics[numFailedBackupsKey].(*prometheus.CounterVec); ok {
		pm.WithLabelValues(policy).Inc()
	}
}

// RecordBackupDuration records the number of seconds taken by a backup.
func (m *MetricsInfo) RecordBackupDuration(policy string, seconds float64) {
	if c, ok := m.metrics[backupDurationSecondsKey].(*prometheus.HistogramVec); ok {
		c.WithLabelValues(policy).Observe(seconds)
	}
}

func toSeconds(d time.Duration) float64 {
	return float64(d / time.Second)
}
