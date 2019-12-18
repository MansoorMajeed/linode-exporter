package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/DazWilkin/linode-exporter/collector"

	"github.com/linode/linodego"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"golang.org/x/oauth2"
)

var (
	// Injected during build using ldflags
	GitCommit string
	OSVersion string
)
var (
	token       = flag.String("linode_token", "", "Linode API Token")
	debug       = flag.Bool("debug", false, "Enable Linode REST API debugging")
	endpoint    = flag.String("endpoint", ":9388", "The endpoint of the HTTP server")
	metricsPath = flag.String("path", "/metrics", "The path on which Prometheus metrics will be served")
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprint(w, "<h2>Linode Exporter</h2>")
	fmt.Fprintf(w, "<a href=\"%s\">metrics</a>", *metricsPath)
}
func main() {
	flag.Parse()
	if *token == "" {
		log.Fatal("Provide Linode API Token")
	}
	source := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: *token,
	})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: source,
		},
	}
	client := linodego.NewClient(oauth2Client)
	client.SetDebug(*debug)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewAccountCollector(client))
	registry.MustRegister(collector.NewExporterCollector(client, OSVersion, GitCommit))
	registry.MustRegister(collector.NewInstanceCollector(client))
	registry.MustRegister(collector.NewNodeBalancerCollector(client))
	registry.MustRegister(collector.NewTicketCollector(client))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(rootHandler))
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Printf("[main] Server starting (%s)", *endpoint)
	log.Printf("[main] metrics served on: %s", *metricsPath)
	log.Fatal(http.ListenAndServe(*endpoint, mux))
}
