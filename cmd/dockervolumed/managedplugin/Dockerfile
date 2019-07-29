FROM alpine:3.9
RUN apk update && apk upgrade && apk add device-mapper multipath-tools xfsprogs e2fsprogs btrfs-progs iproute2 util-linux

# create install paths
RUN mkdir -p /opt/hpe-storage/etc/ && mkdir /opt/hpe-storage/log/
RUN mkdir -p /var/lib/kubelet/plugins/hpe.com/mounts/

# add config files for each plugin type
ADD ["nimble-volume-driver.json", "/opt/hpe-storage/etc/nimble-volume-driver.json"]
ADD ["simplivity-volume-driver.json", "/opt/hpe-storage/etc/simplivity-volume-driver.json"]
ADD ["cv-volume-driver.json", "/opt/hpe-storage/etc/cv-volume-driver.json"]

ADD ["multipath.conf.generic", "/opt/hpe-storage/nimbletune/multipath.conf.generic"]
ADD ["multipath.conf.upstream", "/opt/hpe-storage/nimbletune/multipath.conf.upstream"]

# add binaries
ADD [ "iscsiadm", "/sbin/iscsiadm"]
ADD [ "dockervolumed", "/opt/hpe-storage/dockervolumed" ]
ENTRYPOINT [ "/opt/hpe-storage/dockervolumed" ]