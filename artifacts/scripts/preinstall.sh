#!/bin/sh

echo "Executing preinstall script"

if systemctl list-unit-files | grep "github.com/dv-net/dv-merchant.service"
 then
   systemctl stop github.com/dv-net/dv-merchant.service
fi

echo "Preinstall script done"