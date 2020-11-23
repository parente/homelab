## What is this?

My setup for a homelab kubernetes environment with:

- k3d for pseudo-multi-node cluster in Docker
- Cloudflare for TLS, DNS, and proxying
- nginx ingress with TLS termination using static Cloudflare origin certs and origin pull
  verification
- cfsync for maintaining A record public IP entries for a NATed home network

## Why build it?

A chance to review things I think I know. An opportunity to learn more. An itch to build. Boredom.

## Why not use cert-manager and Let's Encrypt?

Saving on time and memory. Static origin certs are good enough for my purposes.

# What manual steps did I take?

In Cloudflare:

- Enable _Full (strict)_ encryption mode
- Generate a wildcard origin certificate for my domain
- Download the Cloudflare CA for origin pull auth
