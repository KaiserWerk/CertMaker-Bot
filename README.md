# CertMaker Bot

A service app to automagically fetch fresh certificates from a *CertMaker* instance.
In order to obtain certificates for an IP address, a DNS name or an email address, or
multiple, for that matter, you need to create request requirements files with the ``.yaml`` extension.

This project is a work in progress. That means there will be breaking changes and promises 
concerning API stability cannot be made.

## Setup

### Example requirements file

You should create a separate requirements file for every host you are trying to secure.
Post commands can be whatever you like, but make sure they work on your operating system.
Currently, post commands work on `windows` and `linux`, probably on `darwin`.
Subject fields are optional, but can provide helpful additional information.
The minimum amount of days is 1, the maximum is 182 (half a year).

```yaml
domains:
  - pihole
  - pihole.lan
ips:
  - 192.168.178.200
subject:
  organization: My Homelab
  country: DE
  province: Some province
  locality: Some city
  street_address:
  postal_code:
days: 30
cert_file: /opt/app-certs/pihole/cert.pem
key_file: /opt/app-certs/pihole/key.pem
post_commands:
  - cat /opt/app-certs/pihole/key.pem /opt/app-certs/pihole/cert.pem > /opt/app-certs/pihole/combined.pem
```

### Using docker

```powershell
docker run --rm `
  -v ${PWD}\config:/config `
  -v ${PWD}\req:/req `
  -v ${PWD}\certs:/certs `
  kaiserwerk/certmaker-bot:latest `
  --config /config/config.yaml `
  --req /req
```

### As a linux service

This is an example service unit file for ``systemd``, but ``sysvinit`` should be somewhat
similar.

```plain
[Unit]
Description=CertMaker Bot (to cover your certificate needs)
After=network.target

[Service]
Type=simple
ExecStart=/home/certmaker-bot/certmaker --logfile="/var/logs/certmaker-bot.log"
WorkingDirectory=/home/certmaker-bot
User=certmaker-bot
Group=certmaker-bot
Restart=always
RestartSec=15

[Install]
WantedBy=multi-user.target
```

* Place the above service file into ``/etc/systemd/system/certmaker-bot.service``.
* Create a local user, e.g. ``useradd -m -s /bin/bash certmaker-bot``
* Place the binary in the home directory (`/home/certmaker-bot`) of the newly created user
* Execute the commands 
  * ``sudo systemctl daemon-reload``
  * ``sudo systemctl enable certmaker-bot.service``
  * ``sudo systemctl start certmaker-bot.service``
  
  and you're ready to go.

## Command Line Parameters

### Requirement files

The default directory for the request files is ``./req``, relative to the working directory. You can change this directory with the startup 
parameter ``--req``.

Example:
```bash
./certmaker-bot --req="/some/other/dir"
```

### Configuration file

By default, the app reads configuration from ``config.yaml``.
If the file does not exist, the app uses built-in defaults:

* ``app.min_validity: 72h``
* ``app.interval: 6h``
* ``certmaker.host: http://certmaker:8880``
* ``certmaker.skip_verify: false``
* ``certmaker.api_key: ""``

Use ``--config`` to point to a custom file location.

Example:
```bash
./certmaker-bot --config="/some/dir/config.yaml"
```

### Debug mode

Logs are always emitted as JSON. By default, logs are written to standard output.

### Log file location

To additionally write logs to a file, use the ``--logfile`` parameter:

```bash
./certmaker-bot --logfile="/var/logs/certmaker-bot.log"
```

You can combine these parameters as required.
