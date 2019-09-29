#!/usr/bin/env bash
if ! [[ -d /etc/snoopd/ ]]; then
echo "Can't find /etc/snoopd directory"
return 1
fi

case `uname -s` in
    Linux*)     sslConfig=/etc/ssl/openssl.cnf;;
    Darwin*)    sslConfig=/System/Library/OpenSSL/openssl.cnf;;
esac

openssl req \
    -newkey rsa:2048 \
    -x509 \
    -nodes \
    -keyout /etc/snoopd/snoopd.key \
    -new \
    -out /etc/snoopd/snoopd.pem \
    -subj /CN=localhost \
    -reqexts SAN \
    -extensions SAN \
    -config <(cat $sslConfig \
        <(printf '[SAN]\nsubjectAltName=DNS:localhost')) \
    -sha256 \
    -days 3650