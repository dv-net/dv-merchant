#!/bin/sh

echo "Executing postremove script"

if ! [ -e /home/dv/merchant/github.com/dv-net/dv-merchant ]
 then
   echo "Dv Backend removed. Disabling..."
   if systemctl list-unit-files | grep "github.com/dv-net/dv-merchant.service"
    then
       systemctl disable github.com/dv-net/dv-merchant.service
       systemctl stop github.com/dv-net/dv-merchant.service
   fi

     if [ -L /usr/bin/github.com/dv-net/dv-merchant ]; then
       echo "Removing symlink /usr/bin/github.com/dv-net/dv-merchant"
       rm /usr/bin/github.com/dv-net/dv-merchant
     fi
fi

echo "Postremove script done"
