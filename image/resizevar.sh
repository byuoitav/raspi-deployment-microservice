#!/usr/bin/env bash

# stop r/w on var partition
init 1
service dbus stop

# unmount /var to resize it
umount /var

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

e2fsck /dev/mmcblk0p3
resize2fs /dev/mmcblk0p3

# remount partition
mount /dev/mmcblk0p3

# show mount table
echo "done. mount table:\n"
lsblk 
sleep 10

# remove file to indicate 0th boot
rm /usr/bin/games/resize

# reboot :)
init 6
