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

### 3. Upgrade from older plugins

Upgrading from 2.5.1 or older plugins, please follow the below steps

#### Ubuntu 16.04 LTS and Ubuntu 18.04 LTS:

```markdown

1. docker plugin disable nimble:latest –f
2. docker plugin upgrade --grant-all-permissions  nimble store/hpestorage/nimble:3.0.0 --skip-remote-check
3. docker plugin set nimble PROVIDER_IP=10.18.27.197 PROVIDER_USERNAME=admin PROVIDER_PASSWORD=admin glibc_libs.source=/lib/x86_64-linux-gnu
4. docker plugin enable nimble:latest
```

#### Red Hat 7.5+, CentOS 7.5+, Oracle Enterprise Linux 7.5+ and Fedora 28+:

```markdown

1. docker plugin disable nimble:latest –f
2. docker plugin upgrade --grant-all-permissions  nimble store/hpestorage/nimble:3.0.0 --skip-remote-check
3. docker plugin enable nimble:latest
```

**Note**: In Swarm Mode Drain the existing running containers to the node where the plugin is upgraded.
