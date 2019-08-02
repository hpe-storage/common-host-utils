# FAQs

## Troubleshooting Docker Volume Plugin

### 1. Config Directory

The config directory is present at `/etc/hpe`storage/`

a. When a plugin is installed and enabled, the Nimble Group certificates are created in the config directory

```markdown

ll /etc/hpe-storage/
total 16
-r-------- 1 root root 1159 Aug  2 00:20 container_provider_host.cert
-r-------- 1 root root 1671 Aug  2 00:20 container_provider_host.key
-r-------- 1 root root 1521 Aug  2 00:20 container_provider_server.cert
```

b. Additionally there is a config file `volume-driver.json` present at the same location. This file can be edited
to set default parameters for create volumes for docker.

### 2. Log file

The docker plugin logs are located at `/var/log/hpe-docker-plugin.log`
