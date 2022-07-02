#!/usr/bin/env bash

STUDIES=(
  docs/examples/star-wars.json
  docs/examples/star-trek.json
)

for study in "${STUDIES[@]}"; do
  echo "Creating study for ${study}"
  if ./prolificli study create -t "${study}" -p -s ; then
    echo " Created"
  else
    echo " Error"
  fi
done
