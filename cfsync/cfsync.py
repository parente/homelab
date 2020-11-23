#!/usr/bin/env python
"""
Synchronizes Cloudflare A records with ingress hosts using the public IP from ipify.

Useful for local, homelab type k8s setups.

Requires CF_ROOT_DOMAIN, CF_ZONE_ID, and CF_API_KEY. Respects REFRESH_INTERVAL_SEC. Meant to be
run in a pod on the cluster, but can also be run external using kubeconfig.
"""
import logging
import os
import time

import json_logging
import requests

from CloudFlare import CloudFlare
from kubernetes import config, client

logger = logging.getLogger("cfsync")


def configure_logging():
    """"Configures JSON logging."""
    logging.basicConfig(level=logging.INFO)
    json_logging.init_non_web(enable_json=True)
    json_logging.config_root_logger()


def public_ip():
    """Fetches the public IP address of the host from ipify."""
    resp = requests.get("https://api.ipify.org")
    resp.raise_for_status()
    return resp.text


def ingress_hosts(k8s, predicate):
    """Fetches ingress host values across the cluster matching the predicate."""
    hosts = set()
    resp = k8s.list_ingress_for_all_namespaces(watch=False)
    for item in resp.items:
        for rule in item.spec.rules:
            if predicate(rule.host):
                hosts.add(rule.host)
    return hosts


def cloudflare_records(cf, zone_id, predicate):
    """Fetches CloudFlare records matching the predicate."""
    return {
        record["name"]: record
        for record in cf.zones.dns_records.get(zone_id)
        if predicate(record)
    }


def sync(cf, ip, zone_id, hosts, records):
    """Creates, updates, and deletes CloudFlare records to match the current public IP and ingress
    hosts defined in the cluster.
    """
    for host, record in records.items():
        if host not in hosts:
            # Remove record from Cloudflare
            cf.zones.dns_records.delete(record["zone_id"], record["id"])
        elif record["content"] != ip:
            # Update IP address for A record in Cloudflare
            cf.zones.dns_records.put(
                record["zone_id"],
                record["id"],
                data={"name": host, "type": "A", "proxied": True, "content": ip},
            )

    for host in hosts:
        if host not in records:
            # Add new A record to Cloudflare
            cf.zones.dns_records.post(
                zone_id,
                data={"name": host, "type": "A", "proxied": True, "content": ip},
            )


def main():
    root_domain = os.environ["CF_ROOT_DOMAIN"]
    zone_id = os.environ["CF_ZONE_ID"]
    interval = int(os.getenv("REFRESH_INTERVAL_SEC", "60"))

    if os.getenv("KUBERNETES_SERVICE_HOST"):
        config.load_incluster_config()
    else:
        config.load_kube_config()

    k8s = client.NetworkingV1Api()
    cf = CloudFlare()

    configure_logging()

    while 1:
        logger.info("Starting sync")
        ip = public_ip()
        logger.info("Current public IP: %s", ip)
        hosts = ingress_hosts(k8s, lambda host: host.endswith(root_domain))
        logger.info("Ingress hosts: %s", hosts)
        records = cloudflare_records(cf, zone_id, lambda record: record["type"] == "A")
        logger.info("A records: %s", list(records.keys()))
        sync(cf, ip, zone_id, hosts, records)
        logger.info("Completed sync")
        time.sleep(interval)


if __name__ == "__main__":
    main()