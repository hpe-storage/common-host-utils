
# Requirements

- Docker Engine 17.09 or greater
- If using Docker Enterprise Edition 2.x, the plugin is only supported in swarmmode
- Recent Red Hat, Debian or Ubuntu-based Linux distribution
- NimbleOS 5.0.8/5.1.3 or greater on a HPE Nimble Storage array
- US regions only for HPE Cloud Volumes

### HPE Nimble Storage

| Plugin      | HPE Nimble Storage Version | Release Notes    |
|-------------|----------------------------|------------------|
| 3.0.0      | 5.0.8.x and 5.1.3.x onwards | [v3.0.0](release-notes/v3.0.0.md)|
| 3.1.0      | 5.0.8.x and 5.1.3.x onwards | [v3.1.0](release-notes/v3.1.0.md)|

### HPE Cloud Volumes

| Plugin      | Release Notes    |
|-------------|------------------|
| 3.1.0       | [v3.1.0](release-notes/v3.1.0.md)|

**Note:** Docker does not support certified and managed Docker Volume plugins with Kubernetes. If you want to use Kubernetes on Docker with HPE Nimble Storage, please use the [HPE Flexvolume Plugins](https://infosight.hpe.com/tenant/Nimble.Tenant.0013400001Ug0UxAAJ/resources/nimble/software/Integration%20Kits/HPE%20Nimble%20Storage%20Linux%20Toolkit%20(NLT)) and follow the HPE Nimble Storage Integration Guide for Docker Enterprise Edition found on [HPE InfoSight](https://infosight.hpe.com) to deploy a fully supported solution.

# Limitations
HPE Nimble Storage provides a Docker certified plugin delivered through the Docker Store. HPE Nimble Storage also provides a Docker Volume plugin for Windows Containers as part of the Nimble Windows Toolkit (NWT) which is available on [HPE InfoSight](https://infosight.hpe.com/tenant/Nimble.Tenant.0013400001Ug0UxAAJ/resources/nimble/software/Integration%20Kits/HPE%20Nimble%20Storage%20Docker%20Volume%20Plugin). Certain features and capabilities are not available through the managed plugin. Please understand these limitations before deploying either of these plugins.

The managed plugin does NOT provide:

- Support for Docker's release of Kubernetes in Docker Enterprise Edition 2.x
- Support for older versions of NimbleOS (all versions below 5.x)
- Support for Windows Containers

The managed plugin does provide a simple way to manage HPE Nimble Storage and HPE Cloud Volumes integration on your Docker hosts using Docker's interface to install and manage the plugin.

# How to Use this Plugin

## Plugin Privileges
In order to create connections, attach devices and mount file systems, the plugin requires more privileges than a standard application container. These privileges are enumerated during installation. These permissions need to be granted for the plugin to operate correctly.

```
Plugin "nimble" is requesting the following privileges:
 - network: [host]
 - mount: [/dev]
 - mount: [/run/lock]
 - mount: [/sys]
 - mount: [/etc]
 - mount: [/var/lib]
 - mount: [/var/run/docker.sock]
 - mount: [/sbin/iscsiadm]
 - mount: [/lib/modules]
 - mount: [/usr/lib64]
 - allow-all-devices: [true]
 - capabilities: [CAP_SYS_ADMIN CAP_SYS_MODULE CAP_MKNOD]
```

## Host Configuration and Installation

### HPE Nimble Storage

Setting up the plugin varies between Linux distributions. The following workflows have been tested using a Nimble iSCSI group array at **192.168.171.74** with PROVIDER_USERNAME **admin** and PROVIDER_PASSWORD **admin**:

These procedures **requires** root privileges.

Red Hat 7.5+, CentOS 7.5+:
```
yum install -y iscsi-initiator-utils device-mapper-multipath
docker plugin install --disable --grant-all-permissions --alias nimble store/nimblestorage/nimble:3.1.0
docker plugin set nimble PROVIDER_IP=192.168.171.74 PROVIDER_USERNAME=admin PROVIDER_PASSWORD=admin
docker plugin enable nimble
systemctl daemon-reload
systemctl enable iscsid multipathd
systemctl start iscsid multipathd
```

Ubuntu 16.04 LTS and Ubuntu 18.04 LTS:
```
apt-get install -y open-iscsi multipath-tools xfsprogs
modprobe xfs
sed -i"" -e "\$axfs" /etc/modules
docker plugin install --disable --grant-all-permissions --alias nimble store/nimblestorage/nimble:3.1.0
docker plugin set nimble PROVIDER_IP=192.168.171.74 PROVIDER_USERNAME=admin PROVIDER_PASSWORD=admin glibc_libs.source=/lib/x86_64-linux-gnu
docker plugin enable nimble
systemctl daemon-reload
systemctl restart open-iscsi multipath-tools
```

Debian 9.x (stable):
```
apt-get install -y open-iscsi multipath-tools xfsprogs
modprobe xfs
sed -i"" -e "\$axfs" /etc/modules
docker plugin install --disable --grant-all-permissions --alias nimble store/nimblestorage/nimble:3.1.0
docker plugin set nimble PROVIDER_IP=192.168.171.74 PROVIDER_USERNAME=admin PROVIDER_PASSWORD=admin iscsiadm.source=/usr/bin/iscsiadm glibc_libs.source=/lib/x86_64-linux-gnu
docker plugin enable nimble
systemctl daemon-reload
systemctl restart open-iscsi multipath-tools
```

**NOTE:** To use the plugin on Fibre Channel environments use `PROTOCOL=FC` environment variable.

### HPE Cloud Volumes

These procedures **requires** root privileges.

Red Hat 7.5+, CentOS 7.5+:
```
yum install -y iscsi-initiator-utils device-mapper-multipath
docker plugin install --disable --grant-all-permissions --alias cv store/cloudvolumes/cv:3.1.0
docker plugin set cv PROVIDER_IP=cloudvolumes.hpe.com PROVIDER_USERNAME=<access_key> PROVIDER_PASSWORD=<access_secret>
docker plugin enable cv
systemctl daemon-reload
systemctl enable iscsid multipathd
systemctl start iscsid multipathd
```

Ubuntu 16.04 LTS and Ubuntu 18.04 LTS:
```
apt-get install -y open-iscsi multipath-tools xfsprogs
modprobe xfs
sed -i"" -e "\$axfs" /etc/modules
docker plugin install --disable --grant-all-permissions --alias cv store/cloudvolumes/cv:3.1.0
docker plugin set cv PROVIDER_IP=cloudvolumes.hpe.com PROVIDER_USERNAME=<access_key> PROVIDER_PASSWORD=<access_secret> glibc_libs.source=/lib/x86_64-linux-gnu
docker plugin enable cv
systemctl daemon-reload
systemctl restart open-iscsi multipath-tools
```

Debian 9.x (stable):
```
apt-get install -y open-iscsi multipath-tools xfsprogs
modprobe xfs
sed -i"" -e "\$axfs" /etc/modules
docker plugin install --disable --grant-all-permissions --alias cv store/cloudvolumes/cv:3.1.0
docker plugin set cv PROVIDER_IP=cloudvolumes.hpe.com PROVIDER_USERNAME=<access_key> PROVIDER_PASSWORD=<access_secret> iscsiadm.source=/usr/bin/iscsiadm glibc_libs.source=/lib/x86_64-linux-gnu
docker plugin enable cv
systemctl daemon-reload
systemctl restart open-iscsi multipath-tools
```

### Making Changes
The `docker plugin set` command can only be used on the plugin if it is disabled. To disable the plugin, use the `docker plugin disable` command. For example:

```
docker plugin disable nimble
```

or

```
docker plugin disable cv
```

#### Settable Parameters
List of parameters which are supported to be settable by the plugin

|  Parameter              |  Description                                                |  Default |
|-------------------------|-------------------------------------------------------------|----------|
| `PROVIDER_IP`           | HPE Nimble Storage array ip                                 |`""`      |
| `PROVIDER_USERNAME`     | HPE Nimble Storage array username                           |`""`      |
| `PROVIDER_PASSWORD`     | HPE Nimble Storage array password                           |`""`      |
| `PROVIDER_REMOVE`       | Unassociate Plugin from HPE Nimble Storage array            | `false`  |
| `LOG_LEVEL`             | Log level of the plugin (`info`, `debug`, or `trace`)       | `debug`  |
| `SCOPE`                 | Scope of the plugin (`global` or `local`)                   | `global` |
| `PROTOCOL`              | Scsi protocol supported by the plugin (`iscsi` or `fc`)     | `iscsi`  |

### Security Consideration
The HPE Nimble Storage credentials are visible to any user who can execute `docker plugin inspect nimble`. To limit credential visibility, the variables should be unset after certificates have been generated. The following set of steps can be used to accomplish this:

- Add the credentials
  ```
  docker plugin set PROVIDER_IP=192.168.171.74 PROVIDER_USERNAME=admin PROVIDER_PASSWORD=admin
  ```
- Start the plugin
  ```
  docker plugin enable nimble
  ```
- Stop the plugin
  ```
  docker plugin disable nimble
  ```
- Remove the credentials
  ```
  docker plugin set nimble PROVIDER_USERNAME="" PROVIDER_PASSWORD=""
  ```
- Start the plugin
  ```
  docker plugin enable nimble
  ```
**Note:** For HPE Cloud Volumes, these steps are not applicable, as env variables needs to be set for plugin to be
functional
**Note:** Certificates are stored in `/etc/hpe-storage/` on the host and will be preserved across plugin updates.

In the event of reassociating the plugin with a different HPE Nimble Storage group, certain procedures need to be followed:

- Disable the plugin
  ```
  docker plugin disable nimble
  ```
- Set new paramters
  ```
  docker plugin set nimble PROVIDER_REMOVE=true
  ```
- Enable the plugin
  ```
  docker plugin enable nimble
  ```
- Disable the plugin
  ```
  docker plugin disable nimble
  ```
- The plugin is now ready for re-configuration
  ```
  docker plugin set nimble PROVIDER_IP=< New IP address > PROVIDER_USERNAME=admin PROVIDER_PASSWORD=admin PROVIDER_REMOVE=false
  ```

**Note:** The `PROVIDER_REMOVE=false` parameter must be set if the plugin ever has been unassociated from a HPE Nimble Storage group.

### Configuration Files and Options
The configuration directory for the plugin is `/etc/hpe-storage` on the host. Files in this directory are preserved between plugin upgrades. The `/etc/hpe-storage/volume-driver.json` file contains three sections, `global`, `defaults` and `overrides`. The global options are plugin runtime parameters and doesn't have any end-user configurable keys at this time.

The `defaults` map allows the docker host administrator to set default options during volume creation. The docker user may override these default options with their own values for a specific option.

The `overrides` map allows the docker host administrator to enforce a certain option for every volume creation. The docker user may not override the option and any attempt to do so will be silently ignored.

These maps are essential to discuss with the HPE Nimble Storage administrator. A common pattern is that a default protection template is selected for all volumes to fulfill a certain data protection policy enforced by the business it's serving. Another useful option is to override the volume placement options to allow a single HPE Nimble Storage array to provide multi-tenancy for docker environments.

**Note:** `defaults` and `overrides` are dynamically read during runtime while `global` changes require a plugin restart.

Below is an example `/etc/hpe-storage/volume-driver.json` outlining the above use cases:
```json
{
  "global": {
    "nameSuffix": ".docker"
  },
  "defaults": {
    "description": "Volume provisioned by Docker",
    "protectionTemplate": "Retain-90Daily"
  },
  "overrides": {
    "folder": "docker-prod"
  }
}
```

Example config for HPE Cloud Volumes:
```json
    {
      "global": {
                "snapPrefix": "BaseFor",
                "initiators": ["eth0"],
                "automatedConnection": true,
                "existingCloudSubnet": "10.1.0.0/24",
                "region": "us-east-1",
                "privateCloud": "vpc-data",
                "cloudComputeProvider": "Amazon AWS"
      },
      "defaults": {
                "limitIOPS": 1000,
                "fsOwner": "0:0",
                "fsMode": "600",
                "description": "Volume provisioned by the HPE Volume Driver for Kubernetes FlexVolume Plugin",
                "perfPolicy": "Other",
                "protectionTemplate": "twicedaily:4",
                "encryption": true,
                "volumeType": "PF",
                "destroyOnRm": true
      },
      "overrides": {
      }
    }
```


For an exhaustive list of options use the `help` option from the docker CLI:
```
$ docker volume create -d nimble -o help
Nimble Storage Docker Volume Driver: Create Help
Create or Clone a Nimble Storage backed Docker Volume or Import an existing
Nimble Volume or Clone of a Snapshot into Docker.

Universal options:
  -o mountConflictDelay=X X is the number of seconds to delay a mount request
                           when there is a conflict (default is 0)

Create options:
  -o sizeInGiB=X          X is the size of volume specified in GiB
  -o size=X               X is the size of volume specified in GiB (short form
                          of sizeInGiB)
  -o fsOwner=X            X is the user id and group id that should own the
                           root directory of the filesystem, in the form of
                           [userId:groupId]
  -o fsMode=X             X is 1 to 4 octal digits that represent the file
                           mode to be applied to the root directory of the
                           filesystem
  -o description=X        X is the text to be added to volume description
                          (optional)
  -o perfPolicy=X         X is the name of the performance policy (optional)
                          Performance Policies: Exchange 2003 data store,
                          Exchange log, Exchange 2007 data store,
                          SQL Server, SharePoint,
                          Exchange 2010 data store, SQL Server Logs,
                          SQL Server 2012, Oracle OLTP,
                          Windows File Server, Other Workloads,
                          DockerDefault, General, MariaDB,
                          Veeam Backup Repository,
                          Backup Repository

  -o pool=X               X is the name of pool in which to place the volume
                          Needed with -o folder (optional)
  -o folder=X             X is the name of folder in which to place the volume
                          Needed with -o pool (optional).
  -o encryption           indicates that the volume should be encrypted
                          (optional, dedupe and encryption are mutually
                          exclusive)
  -o thick                indicates that the volume should be thick provisioned
                          (optional, dedupe and thick are mutually exclusive)
  -o dedupe               indicates that the volume should be deduplicated
  -o limitIOPS=X          X is the IOPS limit of the volume. IOPS limit should
                          be in range [256, 4294967294] or -1 for unlimited.
  -o limitMBPS=X          X is the MB/s throughput limit for this volume. If
                          both limitIOPS and limitMBPS are specified, limitMBPS
                          must not be hit before limitIOPS
  -o destroyOnRm          indicates that the Nimble volume (including
                          snapshots) backing this volume should be destroyed
                          when this volume is deleted
  -o syncOnUnmount        only valid with "protectionTemplate", if the
                          protectionTemplate includes a replica destination,
                          unmount calls will snapshot and transfer the last
                          delta to the destination. (optional)
 -o protectionTemplate=X  X is the name of the protection template (optional)
                          Protection Templates: General, Retain-90Daily,
                          Retain-30Daily,
                          Retain-48Hourly-30Daily-52Weekly

Clone options:
  -o cloneOf=X            X is the name of Docker Volume to create a clone of
  -o snapshot=X           X is the name of the snapshot to base the clone on
                          (optional, if missing, a new snapshot is created)
  -o createSnapshot       indicates that a new snapshot of the volume should be
                          taken and used for the clone (optional)
  -o destroyOnRm          indicates that the Nimble volume (including
                          snapshots) backing this volume should be destroyed
                          when this volume is deleted
  -o destroyOnDetach      indicates that the Nimble volume (including
                          snapshots) backing this volume should be destroyed
                          when this volume is unmounted or detached

Import Volume options:
  -o importVol=X          X is the name of the Nimble Volume to import
  -o pool=X               X is the name of the pool in which the volume to be
                          imported resides (optional)
  -o folder=X             X is the name of the folder in which the volume to be
                          imported resides (optional)
  -o forceImport          forces the import of the volume.  Note that
                          overwrites application metadata (optional)
  -o restore              restores the volume to the last snapshot taken on the
                          volume (optional)
  -o snapshot=X           X is the name of the snapshot which the volume will
                          be restored to, only used with -o restore (optional)
  -o takeover             indicates the current group will takeover the
                          ownership of the Nimble volume and volume collection
                          (optional)
  -o reverseRepl          reverses the replication direction so that writes to
                          the Nimble volume are replicated back to the group
                          where it was replicated from (optional)

Import Clone of Snapshot options:
  -o importVolAsClone=X   X is the name of the Nimble Volume and Nimble
                          Snapshot to clone and import
  -o snapshot=X           X is the name of the Nimble snapshot to clone and
                          import (optional, if missing, will use the most
                          recent snapshot)
  -o createSnapshot       indicates that a new snapshot of the volume should be
                          taken and used for the clone (optional)
  -o pool=X               X is the name of the pool in which the volume to be
                          imported resides (optional)
  -o folder=X             X is the name of the folder in which the volume to be
                          imported resides (optional)
  -o destroyOnRm          indicates that the Nimble volume (including
                          snapshots) backing this volume should be destroyed
                          when this volume is deleted
  -o destroyOnDetach      indicates that the Nimble volume (including
                          snapshots) backing this volume should be destroyed
                          when this volume is unmounted or detached
```

## Docker Swarm and SwarmKit Considerations
If you are considering using any Docker clustering technologies for your Docker deployment, it is important to understand the fencing mechanism used to protect data. Attaching the same Docker Volume to multiple containers on the same host is fully supported. Mounting the same volume on multiple hosts is not supported.

Docker does not provide a fencing mechanism for nodes that have become disconnected from the Docker Swarm. This results in the isolated nodes continuing to run their containers. When the containers are rescheduled on a surviving node, the Docker Engine will request that the Docker Volume(s) be mounted. In order to prevent data corruption, the Docker Volume Plugin will stop serving the Docker Volume to the original node before mounting it on the newly requested node.

During a mount request, the Docker Volume Plugin inspects the ACR (Access Control Record) on the volume. If the ACR does not match the initiator requesting to mount the volume, the ACR is removed and the volume taken offline. The volume is now fenced off and other nodes are unable to access any data in the volume.

The volume then receives a new ACR matching the requesting initiator, and it is mounted for the container requesting the volume. This is done because the volumes are formatted with XFS, which is not a clustered filesystem and can be corrupted if the same volume is mounted to multiple hosts.

The side effect of a fenced node is that I/O hangs indefinitely, and the initiator is rejected during login. If the fenced node rejoins the Docker Swarm using Docker SwarmKit, the swarm tries to shut down the services that were rescheduled elsewhere to maintain the desired replica set for the service. This operation will also hang indefinitely waiting for I/O.

We recommend running a dedicated Docker host that does not host any other critical applications besides the Docker Engine. Doing this supports a safe way to reboot a node after a grace period and have it start cleanly when a hung task is detected. Otherwise, the node can remain in the hung state indefinitely.

The following kernel parameters control the system behavior when a hung task is detected:

```
# Reset after these many seconds after a panic
kernel.panic = 5

# I do consider hung tasks reason enough to panic
kernel.hung_task_panic = 1

# To not panic in vain, I'll wait these many seconds before I declare a hung task
kernel.hung_task_timeout_secs = 150
```

Add these parameters to the `/etc/sysctl.d/99-hung_task_timeout.conf` file and reboot the system.

**Note:** Docker SwarmKit declares a node as failed after five (5) seconds. Services are then rescheduled and up and running again in less than ten (10) seconds. The parameters noted above provide the system a way to manage other tasks that may appear to be hung and avoid a system panic.

## Use
The [Docker Volume Workflows](https://github.com/hpe-storage/common-host-utils/blob/master/cmd/dockervolumed/managedplugin/examples/README.md) documentation covers basic usage of the plugin. The following example demonstrates the `create` command with a set of options:

```
docker volume create -d nimble \
  -o size=300 \
  -o description="Example" \
  -o perfPolicy="Oracle OLTP" \
  -o encryption="true" \
  -o dedupe="true" \
  -o protectionTemplate="Retain-30Daily" \
  --name example
```

## Uninstall
The plugin can be removed using the `docker plugin rm` command. This command will not remove the configuration directory (`/etc/hpe-storage/`).

```
docker plugin rm nimble
```

**Note:** If this is the last plugin to reference the Nimble Group and to completely remove the configuration directory, follow the steps as below

```
docker plugin set nimble PROVIDER_REMOVE=true
docker plugin enable nimble
docker plugin rm nimble
```

**Note:** For HPE Cloud Volumes, replace `nimble` with plugin name as `cv`

## FAQ
The [FAQ](https://github.com/hpe-storage/common-host-utils/blob/master/cmd/dockervolumed/managedplugin/FAQ.md) documentation covers basic debugging. That documentation applies to this plugin as well. Plugin logs will be under `/var/log`
