#!/bin/sh

openssl req -x509 -newkey rsa:4096 -nodes -keyout ams-han-rpi.key \
    -subj '/CN=ams-han-rpi.client.svenschwermer.de' -days 3650 -out ams-han-rpi.crt
