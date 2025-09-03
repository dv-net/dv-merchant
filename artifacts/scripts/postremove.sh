#!/bin/sh

echo "Executing postremove script"

if ! [ -e /home/dv/merchant/dv-merchant ]
 then
   echo "Dv Backend removed. Disabling..."
   if systemctl list-unit-files | grep "dv-merchant.service"
    then
       systemctl disable dv-merchant.service
       systemctl stop dv-merchant.service
   fi

     if [ -L /usr/bin/dv-merchant ]; then
       echo "Removing symlink /usr/bin/dv-merchant"
       rm /usr/bin/dv-merchant
     fi
fi

echo "Postremove script done"
