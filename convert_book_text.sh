#!/bin/bash

if ! [ -n "$1" ]
then
  echo "no pdf file provided"
  exit 1
fi

mkdir ./outputs
mkdir ./outputs/$1_data
rm ./outputs/$1_data/*

F=0

pdftoppm -jpeg "$1" ./outputs/$1_data/out

while read -r file; do
  go run f.go "$file" > "$file.txt" || F=1
done < <(ls ./outputs/$1_data/out*jpg | sort)

if [ $F -eq 1 ]
then
  exit 1
fi

cat ./outputs/$1_data/out*txt > $1.txt
