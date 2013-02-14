import argparse, sys, requests
import unicodedata
from bs4 import BeautifulSoup
from ZODB import DB
from ZODB.FileStorage import FileStorage
from ZODB.PersistentMapping import PersistentMapping
import transaction
from datetime import datetime
from datetime import timedelta
import time
# Global script settings
MYFIRST_SITE_ROOT = "https://my.usfirst.org/myarea/index.lasso"
# ----------------------

parser = argparse.ArgumentParser(description="Parse data from my.usfirst.org for importing into Scoutmaster 9000.", epilog="(c) 2012 James Shepherdson, Brian Oluwo, Sam Maynard; Team 449")

# Command line arguments
parser.add_argument("-s", "--server", help="The Scoutmaster 9000 server address with which to communicate.  Without this option, the results will just be printed to stdout.", metavar="localhost:9000")
parser.add_argument("-t", "--team", help="A team to look up and add", metavar="<team #>", type=int)
parser.add_argument("-r", "--regional", help="A regional to look up and add.  Will automatically add teams participating  that are not already in the database. Default is -f. YOU HAZ NO CHOICE.  Note that regional name must match the official name on my.usfirst.org.", metavar="\"<regional name>\"")
parser.add_argument("-f", "--force", help="Overwrite teams that already exist in the Scoutmaster 9000 database.  By default, if a team or regional already exists, it will not be changed.", action="store_true")

args = parser.parse_args()

if args.team and args.regional:
	print("Error:  You cannot look up a team and a regional at the same time.")
	sys.exit()

def lookUpTeam(team,specificTeam,regionalLink):
	sSize = 250
	found = False
	while found != True:
	    payload = {"page": "searchresults", "skip_events": "0", "skip_teams": sSize, "programs": "FRC", "season_FRC": "2013", "season_FTC": "2011", "season_FLL": "2011", "season_JFLL": "2011", "season_FP": "2011", "reports": "teams", "area": "All", "results_size":"250"}
	    pageRequest = requests.post(MYFIRST_SITE_ROOT, data=payload)
	    soup = BeautifulSoup(pageRequest.content)
	    print("soup has da stuffs")
	    for link in soup.find_all("a"): # find ALL THE HYPERLINKS!
		    print("Href="+link.get("href"))
		    if link.get("href") and "team_details" in link.get("href"):
			    teamNumber = "".join([str(num) for num in link.findAll(text=True)])
			    print(teamNumber) 
			    if int(teamNumber) == args.team:
#print MYFIRST_SITE_ROOT + link.get("href")
				    pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
				    soup = BeautifulSoup(pageRequest.content)
				    teamNicknameText = soup.find(text='Team Nickname')
				    teamNick = teamNicknameText.findNext('td')
				    print (args.team)
				    print(teamNick)
				    found = True
				    break
				   # storeInDB(teamNumber,teamNick,False)
	    sSize= sSize+250
						#for i in soup.findAll("td"):
					#	if "Team Nickname" in i.get_text():
					#		print i.get_text()

def lookUpRegional(regional):
	sSize = 0
	found = False
	while found != True:
	    payload = {"page": "searchresults", "skip_events": sSize, "skip_teams":"0", "programs": "FRC", "season_FRC": "2013", "season_FTC": "2011", "season_FLL": "2011", "season_JFLL": "2011", "season_FP": "2011", "reports": "events", "area": "All", "results_size":"250"}
	    pageRequest = requests.post(MYFIRST_SITE_ROOT, data=payload)
	    soup = BeautifulSoup(pageRequest.content)
	    for link in soup.find_all("a"): # find ALL THE HYPERLINKS!
		    if link.get("href") and "event_details" in link.get("href"):
			    regionalName = unicodeNormalize("".join([str(num) for num in link.findAll(text=True)]))
			    #print(regionalName)#TODO: Need to give team and matches as children to DB
			    if regionalName == args.regional:
		      		    print(args.regional)
				    pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
				    soup = BeautifulSoup(pageRequest.content)
				    for link in soup.find_all("a"):
				       if link.get("href") and "event_teamlist" in link.get("href"):
					 regionalTeamLink = MYFIRST_SITE_ROOT + link.get("href")
					 print regionalTeamLink#It gets Here
					 pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
					 soup = BeautifulSoup(pageRequest.content)
					 #print soup.prettify()
					 #print("soup has da stuffs")
					 for link in soup.find_all("a"): # find ALL THE HYPERLINKS!
						# print("Href="+link.get("href"))
						 #print("HAI= " + link.contents[0])
						 if link.get("href") and "team_details" in link.get("href"):
							 teamNumber = "".join([str(num) for num in link.findAll(text=True)])
							 print(teamNumber)
							 pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
							 soup = BeautifulSoup(pageRequest.content)
							 teamNicknameText = soup.find(text='Team Nickname')
							 teamNick = teamNicknameText.findNext('td')
							 print(teamNick)
							 #storeInDB(teamNumber,teamNick,True)
				    found = True
				    break
	
				   # storeInDB(teamNumber,teamNick,False)
	sSize= sSize+250
					 #BADDDDDDDD:(lookUpTeam("10000",False,regionalTeamLink)#needs a  way to iterate through the link
					 #break
					 # pageRequest = requests.get(regional)
					 # soup = BeautifulSoup(pageRequest.content)			     
	
#TODO: Add found team to scout.db . Also we need to go to next page Do that by incrementsing skip teams.
def unicodeNormalize(text):
    text = unicode(str(text),"utf-8")
    text = unicodedata.normalize("NFKD", text)
    text = text.encode("ascii", "ignore")
    return text

def storeInDB(data,sData,strOrInt):
	storage = FileStorage("../Scoutmaster-9000-server/scout.db")#TODO: Get Rid of hard code. Make it variablicious.
	db = DB(storage)
	connection = db.open()
	root = connection.root()

if args.team:
	lookUpTeam(args.team,True,None)
#TODO: We need to add a regional def. Runs def Regional same as def team except 'reports = "events"' in payload 
if args.regional:
	lookUpRegional(args.regional)
