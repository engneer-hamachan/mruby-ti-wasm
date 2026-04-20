#!/bin/sh
for f in ./sig/*.rbs; do
  echo "Installing $f..."
  ti-rbs2json --install "$f"
done
