#!/bin/bash

cat <&0 >&2 &

stuff="1"
while true ; do
	echo hi
	sleep .25
	# >&2 echo lo
	# sleep .25
done
