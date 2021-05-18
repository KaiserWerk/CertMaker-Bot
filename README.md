# CertMaker Bot

A service app to automagically fetch fresh certificates from a *CertMaker* instance.
In order to obtain certificates for an IP address or local DNS name, you need to create a request file
with the *.yaml* extension.
The default directory for these request files is *./req*. You can change this directory with the startup 
parameter *--req*.

Example:
```bash
./certmaker-bot --req /some/other/dir
```

To configure the app, there needs to be a `config.yml` file besides the binary. If it doesn't exist,
it will be created at startup.

Eample:
```bash
./certmaker-bot --config /dir/config.yml
```

Usually, all output will be printed to Stdout. If you use the parameter `--as-service`, the log output has
less meta information, but will be written to the `./certmaker-bot.log` log file.

Eample:
```bash
./certmaker-bot --as-service
```

You can combine these parameters as required.