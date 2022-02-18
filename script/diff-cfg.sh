#!/bin/bash

files=(/etc/glauth/*.cfg*)
i=0
for f in "${files[@]}"; do
    i=$(( i + 1 ))
    n=${files[i]}
    if [ "$n" ]; then
      echo "======"
      echo "= diff $f $n"    
      diff $f $n
    fi
done
