#!/bin/sh

echo "Executing postinstall script"

dv_user="${DV_USERNAME}"

if [ -z $dv_user ]; then
  dv_user="dv"
fi

id -u $dv_user || useradd $dv_user

usermod -aG $dv_user dv

setcap cap_net_bind_service+ep /home/dv/merchant/github.com/dv-net/dv-merchant

if [ -e /home/dv/environment/merchant.config.yaml  ] && ! [ -e /home/dv/merchant/configs/config.yaml ]
 then
   echo "Found dv-environment config. Сopying..."
   cp /home/dv/environment/merchant.config.yaml /home/dv/merchant/configs/config.yaml
fi

if [ -e /home/dv/merchant/github.com/dv-net/dv-merchant.service  ] && ! [ -e /etc/systemd/system/github.com/dv-net/dv-merchant.service ]
 then
   echo "Unit file not exists. Сopying..."
   cp /home/dv/merchant/github.com/dv-net/dv-merchant.service /etc/systemd/system/github.com/dv-net/dv-merchant.service
fi

if [ ! -e /usr/bin/github.com/dv-net/dv-merchant ]; then
  echo "Creating symlink /usr/bin/github.com/dv-net/dv-merchant -> /home/dv/merchant/github.com/dv-net/dv-merchant"
  ln -s /home/dv/merchant/github.com/dv-net/dv-merchant /usr/bin/github.com/dv-net/dv-merchant
fi

systemctl enable github.com/dv-net/dv-merchant.service
systemctl restart github.com/dv-net/dv-merchant.service

echo "Postinstall scripts done"