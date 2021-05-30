# CertMaker Bot

A service app to automagically fetch fresh certificates from a *CertMaker* instance.
In order to obtain certificates for an IP address, a DNS name or an email address, or
multiple, for that matter, you need to create request files with the ``.yaml`` extension.

This project is a work in progress. That means there will be breaking changes and promises 
concerning API stability cannot be made.

## Setup

### As a linux service

## Usage

### Command Line Parameters

The default directory for these request files is ``./req``. You can change this directory with the startup 
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
default is 1 hour.

__This logic will be inverted with the first stable release!__

Example:
```bash
./certmaker-bot --as-service
```

You can combine these parameters as required.