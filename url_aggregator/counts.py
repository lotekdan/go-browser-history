import json

def load_history(readfile):
    with open(readfile) as file:
        result = json.load(file)

    return(result)

def history_count(file):
    total_visits = sum(entry["count"] for entry in file.values())

    return(total_visits)

def main():
    files = ['unique_domains_15min.json', 
            'unique_domains_5min.json', 
            'unique_domains_1min.json',
            'unique_urls_15min.json',
            'unique_urls_5min.json',
            'unique_urls_1min.json']
    m = "{} visit counts {}"
    for file in files:
        result = load_history(file)
        print(m.format(file, history_count(result)))

if __name__ == '__main__':
    main()
