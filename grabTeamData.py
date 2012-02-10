import requests as req
from bs4 import BeautifulSoup as BS

payload = {"page": "searchresults", "skip_events": "0", "skip_teams": "0", "programs": "FRC", "season_FRC": "2012", "season_FTC": "2011", "season_FLL": "2011", "season_JFLL": "2011", "season_FP": "2011", "reports": "events", "area": "All", "results_size": "500"}
pageRequest = req.post("https://my.usfirst.org/myarea/index.lasso", data=payload)
soup = BS(pageRequest.content)
links = soup.find(bgcolor="#0066b3", width="100%").find("table").find_all("a")
for i in links:
	print i.get("href")
