#!/bin/bash

current="/etc/glauth/sample-simple.cfg"

files=(/etc/glauth/sample-simple.cfg*)

for f in "${files[@]}"; do
      echo "======"
      echo "= diff $current $f"    
      new=`grep -xvFf $f $current`
      old=`grep -xvFf $current $f`
      if [ "$new" ]; then
        echo "New:"
        echo "$new"
      fi
      if [ "$old" ]; then
        echo "Old:"
        echo "$old"
      fi
done
