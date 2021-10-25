#!/bin/bash
#
# This script uses the accumulate cli to create a transfer transaction
# The script expects two IDs, number of tokesn and server IP:Port to be passed in
#

# if IDs not entered on the command line, prompt for one and exit

if [ -z $1 ]; then
	echo "Usage: cli_xfer_tokens.sh fromID toID numTokens IPAddress:Port"
	exit 0
fi

# see if $1 is really an ID

id1=$1
size=${#id1}
if [ $size -lt 59 ]; then
	echo "Expected acc://<48 byte string>/ACME"
	exit 0
fi

if [ -z $2 ]; then
	echo "Usage: cli_xfer_tokens.sh fromID toID numTokens IPAddress:Port"
	exit 0
fi

# see if $2 is really an ID

id2=$2
size=${#id2}
if [ $size -lt 59 ]; then
	echo "Expected acc://<48 byte string>/ACME"
	exit 0
fi

if [ -z $3 ]; then
	echo "Usage: cli_xfer_tokens.sh fromID toID numTokens IPAddress:Port"
	exit 0
fi

# see if the IP address and port were entered on the command line

if [ -z $4 ]; then
	echo "You must enter an IPAddress:Port for a server to transfer tokens"
	echo "Usage: cli_xfer_tokens.sh fromID toID numTokens IPAddress:Port"
	exit 0
fi

# is there enough of a balance in $1 to xfer to $2?

bal=`./cli_get_balance.sh $id1 $4`

# remove the "s

bal=`echo $bal | /usr/bin/sed 's/"//g'`

if [[ bal -lt $3 ]]; then
	echo "Not enough funds in $1 for requested xfer"
	exit 0
fi

# issue the account get command for the specified ID to the specified server

txid=`cli tx create $id1 $id2 $3 -s "http://$4/v1"`

# did we get a valid txid?

if [ -z "$txid" ]; then
	echo "Transaction failed"
	exit 0
fi

bal=`./cli_get_balance.sh $id2 $4`

# return the balance in id2

echo $bal 

