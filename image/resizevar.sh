#!/usr/bin/env bash

echo "starting resize\n"

# stop r/w on var partition
init 1
service dbus stop

echo "stopped dbus\n"

# unmount /var to resize it
umount /var

echo "unmounted var\n"

fdisk /dev/mmcblk0p3 << EOF 
p
d
3
n
p
3
6481920
p
w
EOF

echo "finished fdisk\n"

e2fsck /dev/mmcblk0p3
resize2fs /dev/mmcblk0p3

echo "fsck'ed and resized\n"

# remount partition
mount /dev/mmcblk0p3

echo "remounted\n"

# show mount table
echo "done. mount table:\n"
lsblk 
sleep 10

# remove file to indicate 0th boot
rm /usr/bin/games/resize

# reboot :)
init 6
