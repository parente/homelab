## What is this?

My setup for a homelab kubernetes environment with:

- k3d for a single-host, multi-node cluster in Docker
- k3sup for a multi-host, multi-node cluster on Raspberry Pis
- Cloudflare for TLS, DNS, and proxying
- nginx ingress with TLS termination using static Cloudflare origin certs and origin pull
  verification
- cfsync for maintaining A record public IP entries for a NATed home network
- minio for object storage
- other apps of interest

## Why build it?

A chance to review things I think I know. An opportunity to learn more. An itch to build. Boredom.

## Why not use cert-manager and Let's Encrypt?

Saving on time and memory. Static origin certs are good enough for my purposes.

## What manual steps did I take?

In Cloudflare:

- Enable _Full (strict)_ encryption mode
- Generate a wildcard origin certificate for my domain and store in `secrets.yaml`
- Download the Cloudflare CA for origin pull auth and store in `values.yaml`

In GitHub:

- Create a `GHCR_TOKEN` secret with a personal access token having package write permission
- Create the `gh-pages` orphan branch

To use the minio `mc` CLI:

- Add a `homelab` alias to the `~/.mc/config.json` file
- Run `make` targets in the `minio` folder

To run on a single Raspberry Pi 3 (ARMv7):

- Install `k3d` onto the Pi
- Clone this project onto the Pi
- Run `make local-cluster`
- Copy the `~/.kube/config` back to my main machine
- Delete the default `local-path` StorageClass (probably should skip install and install custom)
- Run `make sync` from the main machine

To set up a Raspberry Pi 4 (ARMv7) cluster:

- Write empty `ssh` file boot partition
- Write `wpa_supplicant.conf` to boot partition like:

```
country=US
ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1

network={
    ssid="SSID"
    psk="PASSWORD"
}
```

- SSH to `pi@raspberrypi.local`
- Change `pi` user password
- Add SSH pubkey to `~/.ssh/authorized_hosts`
- Run `raspi-config` to set hostname, lower GPU memory, expand root partition
- Add `cgroup_enable=cpuset cgroup_memory=1 cgroup_enable=memory` to `/boot/cmdline.txt`
- Disable wifi power saving with `sudo /sbin/iw wlan0 set power_save off` and permanently in
  `/sbin/iw wlan0 set power_save off` before the exit
- Assign fixed IP
- Repeat for all nodes
- Install `k3sup` on my dev box
- Run `make cluster`

## How do I cut chart releases?

1. Bump versions in `cfsync/chart/Chart.yaml` and `helmfile.yaml`.
2. Push to main.
3. Use the GitHub web UI to create a release with matching version tag.
