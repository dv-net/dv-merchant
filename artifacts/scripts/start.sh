#!/bin/sh

./github.com/dv-net/dv-merchant migrate up --disable-confirmations

./github.com/dv-net/dv-merchant seed

./github.com/dv-net/dv-merchant start