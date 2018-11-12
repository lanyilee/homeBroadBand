#!/bin/sh
ADDR=$1
USER=$2
PASS=$3
echo "Starting to sftp..."
REMOTE_DIR=$4
LOCAL_PATH=$5
[ $? != 0 ] && exit 1;
lftp -u ${USER},${PASS} sftp://${ADDR} <<EOF
cd ${REMOTE_DIR}
mput ${LOCAL_PATH}
bye
EOF

#[ $? != 0 ] && exit 1;

