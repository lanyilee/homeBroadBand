#!/bin/sh
lftp << EOF
set ftp:ssl-auth ssl
set ftp:ssl-force yes
set ssl:verify-certificate no
set ssl:cert-file "/home/tdzx/lanyi/crt/gdjtkd.crt"
set ssl:key-file "/home/tdzx/lanyi/crt/gdjtkd.key.unsecure"
open 221.176.9.229:3132
user cxftp2gdgzjtkd mANSHKJudSXOR
rm $1
ls >>deleteout.txt
bye
EOF
exit 0
