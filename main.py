import json
from collections import Counter

with open("podcasts.json", "r") as file:
    data = json.load(file)
    counter = Counter()
    res = []
    for p in data["Podcasts"]:
        res += [(r["Title"], r["Url"]) for r in p["Episodes"]]
    for e in res:
        counter[e[0]] += 1
    print(counter.most_common(5))
