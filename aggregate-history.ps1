$currentDir = Get-Location
$path = Join-Path $currentDir -ChildPath "url_aggregator\out.json"
$output = go run .\cmd\main.go -d 7 -j
[System.IO.File]::WriteAllText($path, $output)
#gc .\url_aggregator\out.json  
cd url_aggregator
python .\url_aggregator.py -i .\out.json
rm out.json
python .\counts.py
cd ..
