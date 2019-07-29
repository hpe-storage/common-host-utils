#!/bin/bash

mkdir -p build/hpe-docker-plugin
cd build
# cleanup if any existing files are present
rm -rf hpe-docker-plugin/*

if [ "$#" -ne 3 ]; then
    echo "Error: Insufficient parameters provided"
    echo "Usage: ./buildDockerPlugin.sh username pluginname tag"
    exit 1
fi

if [ -z "$1" ]; then
    echo "docker user name must be supplied as ARG1"
    exit 1
fi
# docker hub user name
user=$1

if [ -z "$2" ]; then
    echo "plugin name must be supplied as ARG2"
    exit 1
fi
pluginname=$2

if [ -z "$3" ]; then
    echo "tag must be supplied as ARG3"
    exit 1
fi
tag=$3

docker plugin disable ${user}/${pluginname} -f
docker plugin rm ${user}/${pluginname} -f

for x in `ls /var/lib/docker/plugins`
do
   if [ ${x} = "storage" ] || [ ${x} = "tmp" ]; then
      echo skipping ${x}
   else
      umount /var/lib/docker/plugins/${x}/rootfs/opt/nimble
      rm -rf /var/lib/docker/plugins/${x}
   fi
done
mkdir -p hpe-docker-plugin/etc
basepath=/opt/hpe-storage

mkdir -p hpe-docker-plugin/${basepath}/etc
mkdir -p hpe-docker-plugin/${basepath}/lib
mkdir -p hpe-docker-plugin/${basepath}/log

cp ../Dockerfile hpe-docker-plugin/Dockerfile
cp ../config.json hpe-docker-plugin/config.json
cp ../config/* hpe-docker-plugin/
cp ../multipath.* hpe-docker-plugin/

# TODO pull these from artifactory based on branch
if [[ ! -f ../dockervolumed ]]; then
    echo "dockervolumed binary doesn't exit. try make utils first in common"
    exit 1
fi
cp ../dockervolumed hpe-docker-plugin/dockervolumed
if [[ ! -f ../dory ]]; then
    echo "dory binary doesn't exist. clone k8s repo and build dory binary first in k8s repo"
    exit 1
fi
cp ../dory hpe-docker-plugin/dory

cd hpe-docker-plugin/

mkdir ${pluginname} > /dev/null 2>&1

echo '#!/bin/sh
export LD_LIBRARY_PATH=/lib64
/sbin/ia "$@"' > iscsiadm
chmod 555 iscsiadm

rm -rf ${pluginname}/rootfs
docker build -t rootfsimage .
rc=$?
 if [[ $rc -ne 0 ]]; then
  echo "ERROR: failed"
  exit $rc
fi
id=$(docker create rootfsimage true)
rc=$?
 if [[ $rc -ne 0 ]]; then
  echo "ERROR: failed"
  exit $rc
fi
mkdir -p ${pluginname}/rootfs
rc=$?
 if [[ $rc -ne 0 ]]; then
  echo "ERROR: failed"
  exit $rc
fi
docker export "$id" | tar -x -C ${pluginname}/rootfs
rc=$?
 if [[ $rc -ne 0 ]]; then
  echo "ERROR: failed"
  exit $rc
fi
docker rm -vf "$id"
rc=$?
 if [[ $rc -ne 0 ]]; then
  echo "ERROR: failed"
  exit $rc
fi
docker rmi rootfsimage
rc=$?
 if [[ $rc -ne 0 ]]; then
  echo "ERROR: failed"
  exit $rc
fi
cp config.json ${pluginname}
rc=$?
 if [[ $rc -ne 0 ]]; then
  echo "ERROR: failed"
  exit $rc
fi
docker plugin create ${user}/${pluginname}:${tag} ${pluginname}
docker plugin ls
