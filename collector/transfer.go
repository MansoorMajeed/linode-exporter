package collector

import (
	"context"
	"log"
	"strconv"

	"github.com/linode/linodego"
	"github.com/prometheus/client_golang/prometheus"
)

// TransferCollector represents a Linode Instance Transfer metrics collector
type TransferCollector struct {
	client linodego.Client

	UsedBytes     *prometheus.Desc
	QuotaBytes    *prometheus.Desc
	BillableBytes *prometheus.Desc
}

// NewTransferCollector creates a TransferCollector
func NewTransferCollector(client linodego.Client) *TransferCollector {
	log.Println("[NewTransferCollector] Entered")
	subsystem := "transfer"

	return &TransferCollector{
		client: client,

		UsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "used_bytes"),
			"Total transfer used in current billing period",
			[]string{"linode_id", "label", "region"},
			nil,
		),
		QuotaBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "quota_bytes"),
			"Monthly transfer quota",
			[]string{"linode_id", "label", "region"},
			nil,
		),
		BillableBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "billable_bytes"),
			"Transfer that exceeds quota (billable)",
			[]string{"linode_id", "label", "region"},
			nil,
		),
	}
}

// Collect implements Collector interface and is called by Prometheus to collect metrics
func (c *TransferCollector) Collect(ch chan<- prometheus.Metric) {
	log.Println("[TransferCollector:Collect] Entered")
	ctx := context.Background()

	instances, err := c.client.ListInstances(ctx, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("[TransferCollector:Collect] len(instances)=%d", len(instances))

	for _, instance := range instances {
		log.Printf("[TransferCollector:Collect] Linode ID (%d)", instance.ID)

		transfer, err := c.client.GetInstanceTransfer(ctx, instance.ID)
		if err != nil {
			log.Println(err)
			continue
		}

		labelValues := []string{
			strconv.Itoa(instance.ID),
			instance.Label,
			instance.Region,
		}

		// Values are already in bytes from the API
		usedBytes := float64(transfer.Used)
		// Convert quota from GB to bytes
		quotaBytes := float64(transfer.Quota * 1024 * 1024 * 1024)
		billableBytes := float64(transfer.Billable)

		ch <- prometheus.MustNewConstMetric(
			c.UsedBytes,
			prometheus.GaugeValue,
			usedBytes,
			labelValues...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QuotaBytes,
			prometheus.GaugeValue,
			quotaBytes,
			labelValues...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BillableBytes,
			prometheus.GaugeValue,
			billableBytes,
			labelValues...,
		)
	}
	log.Println("[TransferCollector:Collect] Completes")
}

// Describe implements Collector interface and is called by Prometheus to describe metrics
func (c *TransferCollector) Describe(ch chan<- *prometheus.Desc) {
	log.Println("[TransferCollector:Describe] Entered")
	ch <- c.UsedBytes
	ch <- c.QuotaBytes
	ch <- c.BillableBytes
	log.Println("[TransferCollector:Describe] Completes")
}
