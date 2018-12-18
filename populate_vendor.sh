#!/usr/bin/env bash
#
# detect changes in Gopkg.lock and run godep to populate vendor only if necessary
# output checksum to .cache/vendor.log for future detection
#

logfile=.cache/vendor.log

old_checksum=old
new_checksum=new

if [ -r $logfile ] ; then
    old_checksum=$(cat $logfile)
fi

if [ $(which cksum | wc -l) -gt 0 ] ; then
    new_checksum=$(cksum Gopkg.lock)
else
    echo cksum command not found, cannot cache vendoring
fi

if [ "$old_checksum" == "$new_checksum" ]; then
    echo vendor is up-to-date, no vendoring needed.  To force vendoring regardless, remove $logfile file
else
    dep ensure --vendor-only -v
    dep_exit_code=$?
    if [ $dep_exit_code -eq 0 ] && [ $(which cksum | wc -l) -gt 0 ] ; then
        mkdir -p .cache
        echo $new_checksum > $logfile
    fi
    exit $dep_exit_code
fi
