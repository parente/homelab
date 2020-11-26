package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rdegges/go-ipify"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func cfClient(cfAPIToken string) *cloudflare.API {
	cfAPI, err := cloudflare.NewWithAPIToken(cfAPIToken)
	if err != nil {
		log.Fatal(err)
	}
	return cfAPI
}

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

func publicIP() string {
	ip, err := ipify.GetIp()
	if err != nil {
		log.Fatal(err)
	}
	return ip
}

func ingressHosts(api *kubernetes.Clientset, predicate func(string) bool) *map[string]bool {
	hosts := make(map[string]bool)
	ingresses, err := api.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
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

func dnsRecords(cf *cloudflare.API, zoneID string) *map[string]*cloudflare.DNSRecord {
	records, err := cf.DNSRecords(zoneID, cloudflare.DNSRecord{
		Type: "A",
	})
	if err != nil {
		log.Fatal(err)
	}

	recordMap := make(map[string]*cloudflare.DNSRecord)
	for _, record := range records {
		recordMap[record.Name] = &record
	}

	return &recordMap
}

func sync(
	cf *cloudflare.API,
	ip string,
	zoneID string,
	hosts *map[string]bool,
	records *map[string]*cloudflare.DNSRecord,
) {
	for host, record := range *records {
		if _, ok := (*hosts)[host]; !ok {
			err := cf.DeleteDNSRecord(record.ZoneID, record.ID)
			if err != nil {
				log.Println(fmt.Sprintf("Failed to delete %s: %s", host, err))
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
			}
		}
	}

	for host := range *hosts {
		if _, ok := (*records)[host]; !ok {
			cf.CreateDNSRecord(zoneID, cloudflare.DNSRecord{
				Name:    host,
				Content: ip,
				Type:    "A",
				Proxied: true,
			})
		}
	}
}

func main() {
	cfAPIToken := os.Getenv("CF_API_TOKEN")
	cfZoneID := os.Getenv("CF_ZONE_ID")
	cfRootDomain := os.Getenv("CF_ROOT_DOMAIN")

	k8s := k8sClient()
	cf := cfClient(cfAPIToken)

	ip := ""
	iterations := 0

	for {
		log.Println("Starting sync")
		// Check public IP once every 10 cycles
		if iterations == 0 {
			ip = publicIP()
			log.Println(fmt.Sprintf("Current public IP is: %s", ip))
			iterations = 10
		}
		iterations--

		hosts := ingressHosts(k8s, func(host string) bool {
			return strings.HasSuffix(host, cfRootDomain)
		})
		log.Println(fmt.Sprintf("Ingress hosts: %v", hosts))
		records := dnsRecords(cf, cfZoneID)
		log.Println(fmt.Sprintf("A records: %v", records))
		sync(cf, ip, cfZoneID, hosts, records)
		log.Println("Completed sync")
		time.Sleep(time.Minute)
	}
}
