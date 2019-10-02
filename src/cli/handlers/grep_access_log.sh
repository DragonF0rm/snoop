#!/usr/bin/env bash
tail -n $1 /var/log/snoopd_access.log | sed 's/^.*]: //'