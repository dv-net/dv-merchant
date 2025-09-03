#!/bin/sh

echo "Executing preinstall script"

if systemctl list-unit-files | grep "dv-merchant.service"
 then
   systemctl stop dv-merchant.service
fi

echo "Preinstall script done"