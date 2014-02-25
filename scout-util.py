'''
Utilities package scout-util.py contains all the functions to pull the teams and team data from
the FIRST Robotics official website. 
'''

# Imported: Packages
# ------------------
# unicodedata used to process command line arguments
# transaction used for commiting transactions to database
# time used for timestamp
# json for teh lulz
import argparse, sys, requests
import unicodedata
import time
import json
import urllib2
import re
from collections import defaultdict

# Imported: Selected Modules
# --------------------------
# BeautifulSoup from bs4 for used webpage parsing
# DB from ZODB used for DB instantiation
from bs4 import BeautifulSoup
from datetime import datetime #Used for Timestamp
from datetime import timedelta#Used For TimeStamp

# Global script settings
# ----------------------
reload(sys)
sys.setdefaultencoding("utf-8")
MYFIRST_SITE_ROOT = "https://my.usfirst.org/myarea/index.lasso"
maxRetry = 10
mapOfTeams = {}
requestPayload = {'X-TBA-App-Id': 'frc449:Scoutmaster_Utilities:1'}

# Function: LookUpTeam
# --------------------
# Parses through HTML code provided by the official FRC website and builds the database of teams
# in a JSON file that is stored on the 'teams' page of the 449 website. It pulls the team number,
# team name, array of reviews, and array of photoes.
#
# TODO: Must also pull start and end dates for the regional as well as the match list
# TODO: Check to make sure that particular team or regional is not already there.
def lookUpTeam(team, force):
    sSize = 0
    found = False

    while not found and sSize < 2750:
        payload = {"page": "searchresults", "skip_events": "0", "skip_teams": sSize, "programs": "FRC",
                   "season_FRC": "2014", "reports": "teams", "area": "All", "results_size": "250"}
        pageRequest = requests.post(MYFIRST_SITE_ROOT, data=payload)
        soup = BeautifulSoup(pageRequest.content)
        links =  soup.find_all("a", href=re.compile("team_details"))

        for link in links: # find ALL THE HYPERLINKS!
            teamNumber = "".join(str(num) for num in link.find_all(text=True))
            if int(teamNumber) == team:
                pageRequest = requests.get(MYFIRST_SITE_ROOT + link["href"])
                soup = BeautifulSoup(pageRequest.content)
                teamNick = soup.find(text="Team Nickname").findNext("td").get_text()
                found = True
                
                teamInfo = {"number": teamNumber, "name": teamNick, "reviews": [], "matches": [], "photos": []}
                if force == True:
                    return teamInfo
                urllib2.urlopen("http://0.0.0.0:8080/teams", json.dumps(teamInfo))
                break

        if not found:
            print("Not found on page " + str(sSize / 250) + ".")
            sSize += 250
            print("Now checking page " + str(sSize / 250) + ".")

    if not found:
        print "Team # " + str(team) + " does not exist."
        print "Please ensure it was entered correctly."

# Function: LookUpRegional
# ------------------------
# Recieves a regional name and searches through the FRC website for the regional that matches.
# If found, it enters into the regional page, parses the participating team, and extracts the team
# data 
# 
# FIXME:DUPLICATES IN MATCHES.
def lookUpRegional(regional):
    sSize = 0
    found = False

    while found != True:
        payload = {"page": "searchresults", "skip_events": sSize, "skip_teams": "0", "programs": "FRC",
                   "season_FRC": "2013", "season_FTC": "2011", "season_FLL": "2011", "season_JFLL": "2011",
                   "season_FP": "2011", "reports": "events", "area": "All", "results_size": "250"}
        pageRequest = requests.post(MYFIRST_SITE_ROOT, data=payload)
        soup = BeautifulSoup(pageRequest.content)
        for link in soup.find_all("a", href=re.compile("event_details"), text=True):
            if args.regional == "".join(str(num) for num in link):
                pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
                soup = BeautifulSoup(pageRequest.content)
                
                for link in soup.find_all('a', href=re.compile("event_teamlist|matchresults")):
                    if "event_teamlist" in link.get("href"):
                        regionalTeamLink = MYFIRST_SITE_ROOT + link.get("href")
                        pageRequest = requests.get(regionalTeamLink)
                        soup = BeautifulSoup(pageRequest.content)

                        for link in soup.find_all('a', href=re.compile("team_details"), text=True):
                            teamNumber = "".join(str(num) for num in link)
                            num = int(teamNumber)
                            mapOfTeams[teamNumber] = lookUpTeam(num, True)

                    if "matchresults" in link.get("href"):
                        regionalTeamLink = link.get("href")
                        pageRequest = requests.get(regionalTeamLink)
                        soup = BeautifulSoup(pageRequest.content)
                        startDate = soup.find('table', bgcolor="black").find("tbody").findNext("tr").findNext("td").findNext("td")

                        rows = soup.find('table', style="background: black none repeat scroll 0% 50%; -moz-background-clip: initial; -moz-background-origin: initial; -moz-background-inline-policy: initial; width: 100%;").find("tbody").find_all("tr")
                        matches, num, timeM, skipThree, redScore, blueScore = [], "", "", 0, 0, 0
                         
                        for row in rows:
                            if skipThree < 3:
                                skipThree += 1
                                continue

                            cells = row.find_all("td")
                            timeM = "".join(str(cells[0].get_text())) #What is AM and PM?
                            num = "".join(str(cells[1].get_text()))
                            redTeam = [mapOfTeams[int(cells[i].get_text())] for i in [2,3,4]]
                            blueTeam = [mapOfTeams[int(cells[i].get_text())] for i in [5,6,7]]
                            redScore = int(cells[8].get_text())
                            blueScore = int(cells[9].get_text())

                            matches.append({"number": num, "type": "Qualifications", "time": "".join(timeM), "red": redTeam,
                                "blue": blueTeam, "rScore": redScore, "bScore": blueScore, "winner": "red"})

                        stuffToSend = {"location": "".join(args.regional), "matches": matches}
                        urllib2.urlopen("http://0.0.0.0:8080/regionals", json.dumps(stuffToSend))
                found = True
                break
    sSize += 250

# Function: scrapeTeam
# ------------------------
# Employs the Blue Alliance API and generates the JSON data for the input team given by the team's
# number. The teamNumber should be the official team number meaning it must be in the form of frc###. It
# will then dump everything to the Scoutmaster servers.
def scrapeTeam(matchArray, teamNumber = "", teamNumberArray = [], force = False):
    teamURL = "http://www.thebluealliance.com/api/v1/teams/show?teams="
    if teamNumber != "" and teamNumberArray != []:
        raise TypeError("Cannot define a teamNumber and a teamNumberArray")
    elif teamNumber == "":
        teamJSONArray, url, teamAdded = [], "", 0
        for teams in teamNumberArray:
            url += str(teams).strip() + ","
            teamAdded += 1
            if teamAdded > 50:
                raw = requests.get(teamURL + url[:-1], headers = requestPayload)
                teamJSONArray.append(json.loads(raw.content))
                teamAdded = 0
                url = "" # WHAT!? it was matchURL WHAT
        raw = requests.get(teamURL + url[:-1], headers = requestPayload)
        teamJSONArray.append(json.loads(raw.content))
        teamJSONArray = teamJSONArray.pop(0) #...What? Why are we throwing away the rest of the list?
        for data in teamJSONArray:
            teamNum = data["team_number"]
            teamNick = data["nickname"]
            matches = [match for match in matchArray 
                       if any(int(alliance) == teamNum in match["blue"]) or
                          any(int(alliance) == teamNum in match["red"])]
            teamInfo = {"force" : force, "number": teamNum, "name": teamNick, "reviews": [], "matches": matches, "photos": []}
            requests.post("http://0.0.0.0:8080/teams", json.dumps(teamInfo))
    else:
        raw = requests.get(teamURL + str(teamNumber).strip(), headers = requestPayload)
        data = json.loads(raw.content)
        data = data.pop(0)
        teamNum = data["team_number"]
        teamNick = data["nickname"]
        teamInfo = {"force" : force, "number": teamNum, "name": teamNick, "reviews": [], "matches": [], "photos": []}
        requests.post("http://0.0.0.0:8080/teams", json.dumps(teamInfo))

# Function: scrapeRegional
# ------------------------
# Employs the Blue Alliance API and generates the JSON data for the input regional and year. It 
# then formats the data and dumps it to the server. The force flags determines whether or no the teams
# encountered by the function should be sent to the teams page of Scoutmaster
def scrapeRegional(regionalName, regionalYear = 2014, force = False):
    eventsURL = "http://www.thebluealliance.com/api/v1/events/list?year=" + str(regionalYear).strip()
    keyData = json.loads(requests.get(eventsURL, headers = requestPayload).content)
    index = 0
    
    while keyData[index]['name'] != regionalName:
        index += 1

    regionalURL = "http://www.thebluealliance.com/api/v1/event/details?event=" + str(keyData[index]['key']).strip()
    pageData = requests.get(regionalURL, headers = requestPayload)
    teamData = json.loads(pageData.content)
    listOfMatches = teamData['matches']
    requestsArray = []
    matchURLSuffix = ""
    matchesAdded = 0

    for matches in listOfMatches:
        matchURLSuffix = matchURLSuffix + str(matches) + ","
        matchesAdded += 1
        if matchesAdded > 50:
            pageData = requests.get("http://www.thebluealliance.com/api/v1/match/details?match=" + matchURLSuffix[:-1], headers = requestPayload)
            print pageData.content
            requestsArray.append(json.loads(pageData.content))
            matchesAdded = 0
            matchURLSuffix = ""
    if matchURLSuffix[:-1] != "":
        pageData = requests.get("http://www.thebluealliance.com/api/v1/match/details?match=" + matchURLSuffix[:-1], headers = requestPayload)
        print pageData.content
        requestsArray.append(json.loads(pageData.content))
    matchArray = []
    teamsToScrape = []
    for matchData in requestsArray:
        for match in matchData:
            num = int(match["match_number"])
            level = str(match["competition_level"])
            redTeam = []
            blueTeam = []
            teamsToScrape.append(match["alliances"]["red"]["teams"])
            teamsToScrape.append(match["alliances"]["blue"]["teams"])
            for i in range(0, len(match["alliances"]["red"]["teams"])):
            	redTeam.append(match["alliances"]["red"]["teams"][i][3:])
            	blueTeam.append(match["alliances"]["blue"]["teams"][i][3:]) 
            redScore = int(match["alliances"]["red"]["score"])
            blueScore = int(match["alliances"]["blue"]["score"])
            winner = ""
            if redScore > blueScore:
                winner = "red"
            elif redScore == blueScore:
                winner = "tie"
            else: 
                winner = "blue"
            matchArray.append({"number": num, "type": level, "red": redTeam,"blue": blueTeam, "rScore":int(redScore), "bScore":int(blueScore), "winner":winner})
    winnerCount = getWinnerCount(matchArray)
    jsonData = {"location": regionalName, "matches": matchArray, "winnerCount" : winnerCount, "year" : int(regionalYear)}
    requests.post("http://0.0.0.0:8080/regionals", json.dumps(jsonData))
    for team in teamsToScrape:
        scrapeTeam(matchArray, teamNumberArray = team, force = force)

# Function: getWinnerCount
# ------------------------
# Returns the win, tie, and lose count for each team at a regional. The data will be stored in an array of the
# form [x,y,z] = [win, tie, lose]. 
def getWinnerCount(matchArray):
    blue = {"blue":0, "red":2, "tie":1}
    red = {"red":0, "blue":2, "tie":1}
    winnerArray = {}
    for match in matchArray:
        for team in match['blue']:
            winnerArray[team] = [0,0,0]
        for team in match['red']:
            winnerArray[team] = [0,0,0]
    for match in matchArray:
        for team in match["blue"]:
            winnerArray[team][blue[match["winner"]]] += 1
        for team in match["red"]:
            winnerArray[team][red[match["winner"]]] += 1
    return winnerArray

# Parser settings
# ---------------
parser = argparse.ArgumentParser(description = "Retrieves data from the Blue Alliance website and stores it in Scoutmaster 9000.", epilog="(c) 2013 James Shepherdson, Brian Oluwo, Sam Maynard, Antares Chen; Team 449")
parser.add_argument("-t", "--team", help = "A team to look up and add", metavar = "<team #>", nargs = '*', type = int)
parser.add_argument("-r", "--regional",
                    help="A regional to look up and add. Will automatically add teams participating that are not already in the database. Note that regional name must match the official name on my.usfirst.org.",
                    metavar="\"<regional name>\"")
parser.add_argument("-f", "--force",
                    help="Overwrite teams that already exist in the Scoutmaster 9000 database.  By default, if a team or regional already exists, it will not be changed.",
                    action="store_true")
parser.add_argument("-y", "--year", help = "Adds a regional year option if scraping the regional", metavar = "<year #>", type = int)
args = parser.parse_args()

if args.team:
    for i in range(0, len(args.team)):
        args.team[i] = "frc" + str(args.team[i])
if args.team and args.regional:
    print("Error:  You cannot look up a team and a regional at the same time.")
    sys.exit()
elif args.regional:
    if args.year:
        scrapeTeam(teamNumberArray=args.team, regionalYear=args.year, force=args.force)
    else:
        scrapeTeam(teamNumberArray=args.team, force=args.force)
elif args.team:
    print args.team
    scrapeTeam(teamNumberArray=args.team, force=args.force)
