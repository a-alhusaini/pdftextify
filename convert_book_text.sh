#!/bin/bash

if ! [ -n "$1" ]
then
  echo "no pdf file provided"
  exit 1
fi

BASE_NAME=$(basename "$1" .pdf)

mkdir -p ./outputs/"${BASE_NAME}"_data
rm -f ./outputs/"${BASE_NAME}"_data/*

pdftoppm -jpeg "$1" ./outputs/"${BASE_NAME}"_data/out

for file in ./outputs/"${BASE_NAME}"_data/out*.jpg; do
  echo "Processing $file..."
  
  RETRY_DELAY=15
  
  while true; do
    response=$(llm "transcribe the following document into clean markdown with no other output. Mark unclear sections with <<UNCLEAR>>" -a "$file" 2>err.log)
    exit_code=$?
    
    if [ $exit_code -ne 0 ] || grep -qE "high demand|quota|429|ResourceExhausted" err.log; then
      echo "Rate limit or API error detected (Exit: $exit_code). Log: $(cat err.log)"
      echo "Waiting $RETRY_DELAY seconds before retrying $file..."
      
      sleep $RETRY_DELAY
      RETRY_DELAY=$(( RETRY_DELAY * 2 ))
      if [ $RETRY_DELAY -gt 120 ]; then
        RETRY_DELAY=120
      fi
    else
      echo "$response" > "$file.txt"
      break
    fi
  done

  sleep 4  
done

rm -f err.log

cat $(ls -1 ./outputs/"${BASE_NAME}"_data/out*.txt | sort -V) > "${BASE_NAME}".txt

echo "Done! Full text saved to ${BASE_NAME}.txt"
