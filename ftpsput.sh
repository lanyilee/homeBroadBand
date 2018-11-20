#!/bin/sh
lftp << EOF
set ftp:ssl-auth ssl
set ftp:ssl-force yes
set ssl:verify-certificate no
set ssl:cert-file $1
set ssl:key-file $2
open $3
user $4 $5
echo "success connect to ftps">>ftpsout.txt
put $6
ls >>ftpsout.txt
bye
EOF
echo "success"
exit 0