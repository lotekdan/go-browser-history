#Setting location and write output for history
$currentDir = Get-Location
$path = Join-Path $currentDir -ChildPath "url_aggregator\out.json"
$output = go run .\cmd\main.go -d 7 -j
[System.IO.File]::WriteAllText($path, $output)
 
Set-Location url_aggregator

#URL Aggregation of the history into 6 record sets:
# Domains only 1, 5, 15 minute intervals
# URL Path (no args) 1, 5, 15 minute intervals
python .\url_aggregator.py -i .\out.json
Remove-Item out.json

#Output record counts for the aggregated jsons
python .\counts.py

Set-Location ..
