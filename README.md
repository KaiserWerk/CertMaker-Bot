# CertMaker Bot

A service app to automagically fetch fresh certificates from a *CertMaker* instance.
In order to obtain certificates for an IP address, a DNS name or an email address, or
multiple, for that matter, you need to create request files with the ``.yaml`` extension.

This project is a work in progress. That means there will be breaking changes and promises 
concerning API stability cannot be made.

## Setup

### Example requirements file

```yaml
domains:
  - pihole
  - pihole.lan
ips:
subject:
  organization: My Homelab
  country: DE
  province: Some province
  locality: Cologne
  street_address:
  postal_code:
days: 30
cert_file: /opt/app-certs/pihole/cert.pem
key_file: /opt/app-certs/pihole/key.pem
post_commands:
  - cat /opt/app-certs/pihole/key.pem /opt/app-certs/pihole/cert.pem > /opt/app-certs/pihole/combined.pem
```

### As a linux service

```plain
[Unit]
Description=CertMaker Bot
After=network.target

[Service]
Type=simple
ExecStart=/home/certmaker-bot/certmaker --as-service --logfile /var/logs/certmaker-bot.log
WorkingDirectory=/home/certmaker-bot
User=certmaker-bot
Group=certmaker-bot
Restart=always
RestartSec=15

[Install]
WantedBy=multi-user.target
```

* Create a local user, e.g. ``useradd -m -s /bin/bash certmaker-bot``
* Place the above service file into /etc/systemd/system/certmaker-bot.service.
* Place the binary in the home directory (`/home/certmaker-bot`) of the newly created user
* Execute the commands ``sudo systemctl daemon-reload`` and ``sudo systemctl enable certmaker-bot.service``.

## Usage



### Command Line Parameters

The default directory for the request files is ``./req``. You can change this directory with the startup 
parameter ``--req``.

Example:
```bash
./certmaker-bot --req /some/other/dir
```

To configure the app, there needs to be a ``config.yaml`` file, by default besides the binary. If it 
doesn't exist, it will be created at startup. If you want to use a different location, use the
parameter ``--config``. A good default would usually be ``/etc/certmaker-bot/config.yaml``.

Example:
```bash
./certmaker-bot --config /some/dir/config.yaml
```

Usually, all output will be printed to Stdout. If you use the parameter ``--as-service``, the log output has
less meta information, but will be written to the log file.
Also, the debug interval to check and refresh certificates is 15 seconds; the ``--as-service`` 
default is 6 hours.

__This logic will be inverted with the first stable release at the latest!__

Example:
```bash
./certmaker-bot --as-service
```

To change the location of the log file, use the ``--logfile`` parameter:

```bash
./certmaker-bot --logfile /var/logs/certmaker-bot.log
```

You can combine these parameters as required.