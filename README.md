[![Build Status](https://travis-ci.org/blockassets/cgminer_exporter.svg?branch=master)](https://travis-ci.org/blockassets/cgminer_exporter)

# cgminer Prometheus Exporter

[Prometheus.io](https://prometheus.io/) exporter for the [cgminer API](https://github.com/ckolivas/cgminer/blob/master/API-README).

Thanks to [HyperBitShop.io](https://hyperbitshop.io) for sponsoring this project.

### Usage (defaults):

Start cgminer with `--api-listen`

Start cgminer_exporter like this:

``
./cgminer_exporter -cghost MINER_IP -cgport 4028 -port 4030 -cgtimeout 5s
``

### Setup

Install [dep](https://github.com/golang/dep) and the dependencies...

`make dep`

### Build binary for amd64

`make amd64`

### Install onto miner

The [releases tab](https://github.com/blockassets/cgminer_exporter/releases) has `master` binaries cross compiled for AMD64. These are built automatically on [Travis](https://travis-ci.org/blockassets/cgminer_exporter).

There is install.sh / update.sh / disable.sh helper scripts which automate the process described below. Just create a workers.txt file and add the IP addresses one per line.

Download the latest release and copy the `cgminer_exporter` binary to `/usr/bin`

```
chmod ugo+x cgminer_exporter
scp cgminer_exporter root@SERVER_IP:/usr/bin
```

Create `/etc/systemd/system/cgminer_exporter.service`

```
ssh root@SERVER_IP "echo '
[Unit]
Description=cgminer_exporter
After=init.service

[Service]
Type=simple
ExecStart=/usr/bin/cgminer_exporter
Restart=always
RestartSec=4s
StandardOutput=journal+console

[Install]
WantedBy=multi-user.target
' > /etc/systemd/system/cgminer_exporter.service"
```

Enable the service:

```
ssh root@MINER_IP "systemctl enable cgminer_exporter; systemctl start cgminer_exporter"
```

### Test install

Open your browser to `http://SERVER_IP:4030/metrics`

### Prometheus configuration

`prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'cgminer_exporter'
    static_configs:
      - targets: ['localhost:4030']
```
