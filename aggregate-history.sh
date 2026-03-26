#!/bin/bash
#Get history and write to url_aggregator directory
go run ./cmd/main.go -d 7 -j > url_aggregator/out.json

cd url_aggregator

#URL Aggregation of the history into 6 record sets:
# Domains only 1, 5, 15 minute intervals
# URL Path (no args) 1, 5, 15 minute intervals
python url_aggregator.py -i out.json
rm out.json

#Output record counts for the aggregated jsons
python counts.py
