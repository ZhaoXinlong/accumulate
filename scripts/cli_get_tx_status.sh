#!/bin/bash
#
# This script uses the accumulate cli to get the status of a transaction
# The script expects an ID and server IP:Port to be passed in
#

# if ID entered on the command line, prompt for one and exit

if [ -z $1 ]; then
	echo "Usage: cli_get_tx_status.sh ID IPAddress:Port"
	exit 0
fi

# see if $1 is really an ID

id1=$1
size=${#id1}
if [ $size -lt 48 ]; then
        echo "Expected 48 byte string"
        exit 0
fi

# see if the IP address and port were entered on the command line
# issue the faucet command for the specified ID to the specified server

if [ -z $2 ]; then
	Status="$($cli tx get $id1 2>&1 > /dev/null | /usr/bin/jq .status.code | /usr/bin/sed 's/\"/g')"
else
	Status="$($cli tx get $id1 -s http://$2/v1 2>&1 > /dev/null | /usr/bin/jq .status.code | /usr/bin/sed 's/\"/g')"
fi

# return the status information

echo $Status 


