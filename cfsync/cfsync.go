// Command cfsync synchronizes Cloudflare A records with ingress hosts using the public IP from
// ipify. It is useful for local, homelab type k8s setups. Can run as a pod in-cluster or using a
// a kubeconfig out-of-cluster.
//
// The command requires the following environment variables to function properly:
//
// CF_ROOT_DOMAIN: Cloudflare domain name to sync
// CF_ZONE_ID: Cloudflare zone ID of the root domain
// CF_API_TOKEN: Cloudflare API token granting Zone read and DNS edit
//
// The command also respects the following env vars:
//
// SYNC_INTERVAL: Interval between Kubernetes-Cloudflare syncs (default: 1m)
// IP_INTERVAL: Interval between public IP checks (default: 10m)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rdegges/go-ipify"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// cfClient builds a new Cloudflare client
func cfClient(cfAPIToken string) *cloudflare.API {
	cfAPI, err := cloudflare.NewWithAPIToken(cfAPIToken)
	if err != nil {
		log.Fatal(err)
	}
	return cfAPI
}

// k8sClient builds a new Kubernetes client from in-cluster (preferred) or out-of-cluster config
func k8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal("Could not find in- or out-of-cluster configuration")
		}
		log.Println("Using out-of-cluster configuration")
	} else {
		log.Println("Using in-cluster configuration")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return clientset
}

// ingressHosts gets the set of host names satisfying the predicate used by ingresses across the
// entire k8s cluster.
func ingressHosts(api *kubernetes.Clientset, predicate func(string) bool) *map[string]bool {
	hosts := make(map[string]bool)
	ingresses, err := api.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err)
	}
	for _, ingress := range ingresses.Items {
		for _, rule := range ingress.Spec.Rules {
			if predicate(rule.Host) {
				hosts[rule.Host] = true
			}
		}
	}

	return &hosts
}

// dnsRecords gets the set of Cloudflare A records in the zone.
func dnsRecords(cf *cloudflare.API, zoneID string) *map[string]cloudflare.DNSRecord {
	records, err := cf.DNSRecords(zoneID, cloudflare.DNSRecord{
		Type: "A",
	})
	if err != nil {
		log.Println(err)
	}

	recordMap := make(map[string]cloudflare.DNSRecord)
	for _, record := range records {
		recordMap[record.Name] = record
	}

	return &recordMap
}

// syncRecords creates, updates, and deletes Cloudflare records to match the current public IP and
// ingress hosts.
func syncRecords(
	cf *cloudflare.API,
	ip string,
	zoneID string,
	hosts *map[string]bool,
	records *map[string]cloudflare.DNSRecord,
) {
	for host, record := range *records {
		if _, ok := (*hosts)[host]; !ok {
			err := cf.DeleteDNSRecord(record.ZoneID, record.ID)
			if err != nil {
				log.Println(fmt.Sprintf("Failed to delete %s: %s", host, err))
			} else {
				delete(*hosts, host)
				log.Println(fmt.Sprintf("Deleted record %s", host))
			}

		} else if record.Content != ip {
			err := cf.UpdateDNSRecord(record.ZoneID, record.ID, cloudflare.DNSRecord{
				Name:    host,
				Content: ip,
				Type:    "A",
				Proxied: true,
			})
			if err != nil {
				log.Println(fmt.Sprintf("Failed to update %s: %s", host, err))
			} else {
				log.Println(fmt.Sprintf("Updated record: %s", host))
			}
		}
	}

	for host := range *hosts {
		if _, ok := (*records)[host]; !ok {
			_, err := cf.CreateDNSRecord(zoneID, cloudflare.DNSRecord{
				Name:    host,
				Content: ip,
				Type:    "A",
				Proxied: true,
			})
			if err != nil {
				log.Println(fmt.Sprintf("Failed to create %s: %s", host, err))
			} else {
				log.Println(fmt.Sprintf("Created record: %s", host))
			}
		}
	}
}

// ipifyIP gets the public IPv4 address of the requester according to ipify.org
func ipifyIP() string {
	ip, err := ipify.GetIp()
	if err != nil {
		log.Println(err)
	}
	return ip
}

type publicIP struct {
	ip    string
	mutex sync.RWMutex
}

func (p *publicIP) Get() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.ip
}

func (p *publicIP) Set(ip string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.ip = ip
}

func main() {
	cfAPIToken := os.Getenv("CF_API_TOKEN")
	cfZoneID := os.Getenv("CF_ZONE_ID")
	cfRootDomain := os.Getenv("CF_ROOT_DOMAIN")

	var err error

	// Configure the interval for syncing between Kubernetes and Cloudflare
	var syncInterval time.Duration
	syncIntervalStr := os.Getenv("SYNC_INTERVAL")
	if syncIntervalStr == "" {
		syncInterval = time.Minute
	} else {
		syncInterval, err = time.ParseDuration(syncIntervalStr)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Interval for checking the public IP
	var ipInterval time.Duration
	ipIntervalStr := os.Getenv("IP_INTERVAL")
	if ipIntervalStr == "" {
		ipInterval = 10 * time.Minute
	} else {
		ipInterval, err = time.ParseDuration(ipIntervalStr)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Build clients
	k8s := k8sClient()
	cf := cfClient(cfAPIToken)

	// Fetch initial public IP
	pip := publicIP{}
	pip.Set(ipifyIP())

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Goroutine for DNS record sync
	go func() {
		defer wg.Done()
		for {
			log.Println("Starting sync")
			hosts := ingressHosts(k8s, func(host string) bool {
				return strings.HasSuffix(host, cfRootDomain)
			})
			log.Println(fmt.Sprintf("Ingress hosts: %v", hosts))
			records := dnsRecords(cf, cfZoneID)
			log.Println(fmt.Sprintf("A records: %d", len(*records)))
			syncRecords(cf, pip.Get(), cfZoneID, hosts, records)
			log.Println("Completed sync")
			time.Sleep(syncInterval)
		}
	}()

	// Goroutine for public IP checks
	go func() {
		defer wg.Done()
		for {
			log.Println("Checking public IP")
			ip := ipifyIP()
			pip.Set(ip)
			log.Println(fmt.Sprintf("Completed IP check: %s", ip))
			time.Sleep(ipInterval)
		}
	}()

	wg.Wait()
}
