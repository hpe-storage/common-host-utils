# Docker Volume Workflows

## 1. Create a Docker Volume

**Note:** The plugin applies a set of default options when you `create` new volumes unless you override them using the volume create option flags.

### Procedure

1; Create a Docker volume with the `--description` option.

```docker volume create -d nimble --name volume_name [-o description="volume_description"]```

2; (Optional) Inspect the new volume.docker volume inspect volume_name.

3; (Optional) Attach the volume to an interactive container.

```docker run -it --rm -v volume:/mountpoint --name=volume_name```

The volume is mounted inside the container specified in the `volume:/mountpoint` variable.

### Example

Create a Docker volume with the name `twiki` and the description `Hello twiki`.

`docker volume create -d nimble --name twiki -o description="Hello twiki"`

Inspect the new volume.

``` volume inspect twiki
[
    {
        "CreatedAt": "0001-01-01T00:00:00Z",
        "Driver": "nimble:latest",
        "Labels": {},
        "Mountpoint": "",
        "Name": "twiki",
        "Options": {},
        "Scope": "global",
        "Status": {
            "ApplicationCategory": "Virtual Server",
            "Blocksize": 4096,
            "CachePinned": false,
            "CachingEnabled": true,
            "Connections": 0,
            "DedupeEnabled": true,
            "Description": "Docker knows this volume as twiki.",
            "EncryptionCipher": "none",
            "Group": "group-nimble",
            "ID": "067340c04d598d01860000000000000000000000ab",
            "LimitIOPS": -1,
            "LimitMBPS": -1,
            "LimitSnapPercentOfSize": -1,
            "LimitVolPercentOfSize": 100,
            "PerfPolicy": "DockerDefault",
            "Pool": "default",
            "Serial": "0362a739df9e4b1d6c9ce900cbc1cd2a",
            "SnapUsageMiB": 0,
            "Snapshots": [],
            "ThinlyProvisioned": true,
            "VolSizeMiB": 10240,
            "VolUsageMiB": 0,
            "VolumeName": "twiki.docker",
            "delayedCreate": "true",
            "destroyOnDetach": false,
            "destroyOnRm": false,
            "filesystem": "xfs",
            "inUse": false
        }
    }
]
```

Attach the volume to an interactive container.

```docker run -it --rm -v twiki:/data alpine /bin/sh```

## 2. Clone a Docker Volume

Use the create command with the cloneOf option to clone a Docker volume to a new Docker volume.

### Procedure

Clone a Docker volume.

```docker volume create -d nimble --name=clone_name -o cloneOf=volume_name snapshot=snapshot_name```

### Example

Clone the Docker volume named twiki to a new Docker volume named twikiClone.

```docker volume create -d nimble --name=twikiClone -o cloneOf=twiki```
```wikiClone```

Select the snapshot on which to base the clone.

```docker volume create -d nimble -o cloneOf=somevol -o snapshot=mysnap --name myclone```

## 3. Provisioning Docker Volumes

There are several ways to provision a Docker volume depending on what tools are used:

- Docker Engine (CLI)
- Docker SwarmKit
- Docker Compose file with either Docker UCP or Docker Engine

The Docker Volume plugin leverages the existing Docker CLI and APIs, therefor all native Docker tools may be used to provision a volume.

**Note**: The plugin applies a set of default volume create options. Unless you override the default options using the volume option flags, the defaults are applied when you create volumes. For example, the default volume size is 10GiB.  
Config file `volume-driver.json`, which is stored at `/etc/hpe-storage/volume-driver.json:`

```
{
    "global":   {},
    "defaults": {
                 "sizeInGiB":"10",
                 "limitIOPS":"-1",
                 "limitMBPS":"-1",
                 "perfPolicy": "DockerDefault",
                },
    "overrides":{}
}
```

## 3.  Import a Volume to Docker

### Before you begin

Take the volume you want to import offline before importing it. For information about how to take a volume offline, refer to either the
`CLI Administration Guide` or the `GUI Administration Guide`. Use the create command with the importVol option to import an HPE Nimble Storage volume to Docker and name it.

### Procedure

Import an HPE Nimble Storage volume to Docker.

```docker volume create –d nimble--name=volume_name -o importVol=imported_volume_name```

### Example

Import the HPE Nimble Storage volume named importMe as a Docker volume named imported.

```docker volume create -d nimble --name=imported -o importVol=importMe```
```imported```

## 4. Import a Volume Snapshot to Docker

Use the create command with the `importVolAsClone` option to import a HPE Nimble Storage volume snapshot as a Docker volume. Optionally, specify a particular snapshot on the HPE Nimble Storage volume using the snapshot option. The new Docker volume name is in the format of volume:snapshot

### Procedure

Import a HPE Nimble Storage volume snapshot to Docker.

```docker volume create -d nimble --name=imported_snapshot -o importVolAsClone=volume -o snapshot=snapshot```

**Note:** If no snapshot is specified, the latest snapshot on the volume is imported.

### Example

Import the HPE Nimble Storage snapshot `aSnapshot` on the volume `importMe` as a Docker volume named `importedSnap`.

```docker volume create –d nimble –-name=importedSnap –o importVolAsClone=importMe –o snapshot=aSnapshot```
```importedSnap```

## 5.  Restore an Offline Docker Volume with Specified Snapshot

### Procedure

If the volume snapshot is not specified, the last volume snapshot is used.

### Example

```docker volume create -d nimble -o importVol=mydockervol -o forceImport -o restore -o snapshot=snap.for.mydockervol.1```

## 6. Create a Docker Volume using the HPE Nimble Storage Local Driver

A Docker volume created using the HPE Nimble Storage local driver is only accessible from the Docker host that created the HPE Nimble Storage volume. No other Docker hosts in the Docker Swarm will have access to that HPE Nimble Storage volume.

### Procedure

1. Create a Docker volume using the HPE Nimble Storage local driver.

```docker volume create -d nimble-local --name=local```
2. (Optional) Inspect the new volume.

```dockervolume inspectlocal```

### Example

Create a Docker volume with the HPE Nimble Storage local driver, and then inspect it.

## 7. List Volumes

### Procedure

List Docker volumes.

```docker volume ls```

### Example

``` docker volume ls
DRIVER              VOLUME NAME
nimble              twiki
nimble              twikiClone
```

## 8.  Run a Container Using a Docker Volume

### Procedure

Run a container dependent on a HPE Nimble Storage-backed Docker volume.

```docker run -it -v volume:/mountpoint --name= volume_name```

### Example

Start the container twiki:/foo using the volume twiki.

```docker run -it -v twiki:/foo alpine /bin/sh --name=twiki```


## 8. Remove a Docker Volume

When you remove volumes from Docker control they are set to the offline state on the array. Access to the volumes and related snapshots using the Docker Volume plugin can be reestablished.

**Note:** To delete volumes from the HPE Nimble Storage array using the remove command, the volume should have been created
with a `-o destroyOnRm` flag.

**Important:** Be aware that when this option is set to true, volumes and all related snapshots are deleted from the group, and can no longer be accessed by the Docker Volume plugin.

### Procedure

Remove a Docker volume.

```docker volume rm volume_name```

### Example

Remove the volume named twikiClone.

```docker volume rm twikiClone```
