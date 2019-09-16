# Docker Volume Workflows

## Create a Docker Volume
Using `docker volume create`.

**Note:** The plugin applies a set of default options when you `create` new volumes unless you override them using the `volume create -o key=value` option flags.

### Example

- Create a Docker volume with a custom description:
  ```
  docker volume create -d nimble -o description="My volume description" --name myvol1 
  ```

- (Optional) Inspect the new volume:
  ```
  docker volume inspect myvol1
  ```

- (Optional) Attach the volume to an interactive container.

  ```
  docker run -it --rm -v myvol1:/data bash
  ```

The volume is mounted inside the container on `/data`.

## Clone a Docker Volume
Use the `docker volume create` command with the `cloneOf` option to clone a Docker volume to a new Docker volume.

### Example
Clone the Docker volume named `myvol1` to a new Docker volume named `myvol1-clone`.

```
docker volume create -d nimble -o cloneOf=myvol1 --name=myvol1-clone
```

(Optional) Select a snapshot on which to base the clone.

```
docker volume create -d nimble -o snapshot=mysnap1 -o cloneOf=myvol1 --name=myvol2-clone
```

## Provisioning Docker Volumes
There are several ways to provision a Docker volume depending on what tools are used:

- Docker Engine (CLI)
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

## Import a Volume to Docker

### Before you begin
Take the volume you want to import offline before importing it. For information about how to take a volume offline, refer to either the `CLI Administration Guide` or the `GUI Administration Guide` on [HPE InfoSight](https://infosight.hpe.com). Use the `create` command with the `importVol` option to import an HPE Nimble Storage volume to Docker and name it.

### Example
Import the HPE Nimble Storage volume named `mynimblevol` as a Docker volume named `myvol3-imported`.

```
docker volume create –d nimble -o importVol=mynimblevol --name=myvol3-imported
```

## Import a Volume Snapshot to Docker

Use the create command with the `importVolAsClone` option to import a HPE Nimble Storage volume snapshot as a Docker volume. Optionally, specify a particular snapshot on the HPE Nimble Storage volume using the snapshot option.

### Procedure
Import a HPE Nimble Storage volume snapshot to Docker.

### Example
Import the HPE Nimble Storage snapshot `aSnapshot` on the volume `importMe` as a Docker volume named `importedSnap`.

```
docker volume create -d nimble -o importVolAsClone=mynimblevol -o snapshot=mysnap1 --name=myvol4-clone
```

**Note:** If no snapshot is specified, the latest snapshot on the volume is imported.

## Restore an Offline Docker Volume with Specified Snapshot
It's important that the volume to be restored is in an offline state on the array.

### Example
If the volume snapshot is not specified, the last volume snapshot is used.

```
docker volume create -d nimble -o importVol=myvol1.docker -o forceImport -o restore -o snapshot=mysnap1 --name=myvol1-restored
```

## List Volumes
List Docker volumes.

### Example

```
docker volume ls
DRIVER                     VOLUME NAME
nimble:latest              myvol1
nimble:latest              myvol1-clone
```

## Remove a Docker Volume
When you remove volumes from Docker control they are set to the offline state on the array. Access to the volumes and related snapshots using the Docker Volume plugin can be reestablished.

**Note:** To delete volumes from the HPE Nimble Storage array using the remove command, the volume should have been created
with a `-o destroyOnRm` flag.

**Important:** Be aware that when this option is set to true, volumes and all related snapshots are deleted from the group, and can no longer be accessed by the Docker Volume plugin.

### Example

Remove the volume named `myvol1`.

```
docker volume rm myvol1
```
