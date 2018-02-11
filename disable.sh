#!/usr/bin/env bash

# This script requires sshpass and parallel
# OSX: https://gist.github.com/arunoda/7790979

passwd="bwcon"
export SSHPASS="${passwd}"

export SERVICE="cgminer_exporter"

if [ -e "./workers.txt" ] ; then
	WORKERS=`cat ./workers.txt`
fi

if [ -z "${WORKERS}" ] ; then
	echo "Need some workers to disable!"
	exit 1
fi

dowork() {
	ipaddr=$1
	echo "----------- ${ipaddr} start"
	sshpass -e ssh -o StrictHostKeychecking=no root@$ipaddr "systemctl stop ${SERVICE}; systemctl disable ${SERVICE}"
	echo "----------- ${ipaddr} finish"
}

export -f dowork

parallel --no-notice dowork ::: ${WORKERS}
