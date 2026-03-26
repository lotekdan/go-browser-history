#!/bin/bash
go run ./cmd/main.go -d 7 -j > url_aggregator/out.json
cd url_aggregator
python url_aggregator.py -i out.json
rm out.json
python counts.py