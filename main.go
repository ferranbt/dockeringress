package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/miekg/dns"
)

type resolver struct {
	label   string
	lock    sync.Mutex
	records map[string][]string
}

func (rr *resolver) setContainers(cn []types.Container) {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	rr.records = map[string][]string{}
	for _, c := range cn {
		dnsName := c.Labels["dnsresolve"] + "."
		if _, ok := rr.records[dnsName]; !ok {
			rr.records[dnsName] = []string{}
		}

		// only one network expected
		networks := c.NetworkSettings.Networks
		if len(networks) != 1 {
			continue
		}
		var network *network.EndpointSettings
		for _, net := range networks {
			network = net
		}
		rr.records[dnsName] = append(rr.records[dnsName], network.IPAddress)
	}
}

func (rr *resolver) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("A Query for %s\n", q.Name)

			for _, ip := range rr.records[q.Name] {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		default:
			log.Printf("Unknown dns query: %d\n", q.Qtype)
		}
	}
}

func (rr *resolver) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		rr.parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.Usage = func() {}

	var port uint64
	var label string
	var service string

	flags.Uint64Var(&port, "port", 53, "")
	flags.StringVar(&label, "label", "dnsresolve", "")
	flags.StringVar(&service, "service", "service", "")

	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse args: %s\n", err.Error())
		os.Exit(1)
	}

	rr := &resolver{
		label: label,
	}
	if err := watchContainers(rr); err != nil {
		log.Fatalf("Failed to connect to docker: %s\n ", err.Error())
		os.Exit(1)
	}

	// attach request handler func
	dns.HandleFunc(service+".", rr.handleDnsRequest)

	// start server
	server := &dns.Server{Addr: ":" + strconv.Itoa(int(port)), Net: "udp"}
	log.Printf("Starting at %d, label %s, service %s\n", port, label, service)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}

func watchContainers(r *resolver) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	go func() {
		for {
			filters := filters.NewArgs()
			filters.Add("label", r.label)

			containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
				Filters: filters,
			})
			if err != nil {
				log.Fatalf("Failed to list containers: %s\n ", err.Error())
			}

			r.setContainers(containers)
			time.Sleep(1 * time.Second)
		}
	}()

	return nil
}
