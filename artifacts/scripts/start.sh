#!/bin/sh

./dv-merchant migrate up --disable-confirmations

./dv-merchant seed

./dv-merchant start