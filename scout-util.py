import argparse, sys, requests
from bs4 import BeautifulSoup

# Global script settings
MYFIRST_SITE_ROOT = "https://my.usfirst.org/myarea/index.lasso"
# ----------------------

parser = argparse.ArgumentParser(description="Parse data from my.usfirst.org for importing into Scoutmaster 9000.", epilog="(c) 2012 James Shepherdson, Team 449")

# Command line arguments
parser.add_argument("-s", "--server", help="The Scoutmaster 9000 server address with which to communicate.  Without this option, the results will just be printed to stdout.", metavar="localhost:9000")
parser.add_argument("-t", "--team", help="A team to look up and add", metavar="<team #>", type=int)
parser.add_argument("-r", "--regional", help="A regional to look up and add.  Will automatically add teams participating  that are not already in the database.  Note that regional name must match the official name on my.usfirst.org.", metavar="\"<regional name>\"")
parser.add_argument("-f", "--force", help="Overwrite teams that already exist in the Scoutmaster 9000 database.  By default, if a team or regional already exists, it will not be changed.", action="store_true")

args = parser.parse_args()

if args.team and args.regional:
	print("Error:  You cannot look up a team and a regional at the same time.")
	sys.exit()

def lookUpTeam(team):
	payload = {"page": "searchresults", "skip_events": "0", "skip_teams": "0", "programs": "FRC", "season_FRC": "2012", "season_FTC": "2011", "season_FLL": "2011", "season_JFLL": "2011", "season_FP": "2011", "reports": "teams", "area": "All", "results_size": "250"}
	pageRequest = requests.post(MYFIRST_SITE_ROOT, data=payload)
	soup = BeautifulSoup(pageRequest.content)
	for link in soup.find_all("a"): # find ALL THE HYPERLINKS!
		if link.get("href") and "team_details" in link.get("href"):
			teamNumber = "".join([str(num) for num in link.findAll(text=True)])
			if int(teamNumber) == args.team:
				print MYFIRST_SITE_ROOT + link.get("href")
				pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
				soup = BeautifulSoup(pageRequest.content)
				for i in soup.findAll("tr"):
					if "Team Nickname" in i.get_text():
						print i.get_text()

if args.team:
	lookUpTeam(args.team)
