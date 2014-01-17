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
import transaction
import time
import json
import urllib2

# Imported: Selected Modules
# --------------------------
# BeautifulSoup from bs4 for used webpage parsing
# DB from ZODB used for DB instantiation
from bs4 import BeautifulSoup
from datetime import datetime #Used for Timestamp
from datetime import timedelta#Used For TimeStamp
import re

# Global script settings
# ----------------------
reload(sys)
sys.setdefaultencoding("utf-8")
MYFIRST_SITE_ROOT = "https://my.usfirst.org/myarea/index.lasso"
maxRetry = 10
mapOfTeams = {}

# Parser settings
# ---------------
parser = argparse.ArgumentParser(description="Parse data from my.usfirst.org for importing into Scoutmaster 9000.", epilog="(c) 2013 James Shepherdson, Brian Oluwo, Sam Maynard; Team 449")
parser.add_argument("-t", "--team", help="A team to look up and add", metavar="<team #>", type=int)
parser.add_argument("-r", "--regional",
                    help="A regional to look up and add. Will automatically add teams participating that are not already in the database. Note that regional name must match the official name on my.usfirst.org.",
                    metavar="\"<regional name>\"")
parser.add_argument("-f", "--force",
                    help="Overwrite teams that already exist in the Scoutmaster 9000 database.  By default, if a team or regional already exists, it will not be changed.",
                    action="store_true")
args = parser.parse_args()

if args.team and args.regional:
    print("Error:  You cannot look up a team and a regional at the same time.")
    sys.exit()
if args.team:
    lookUpTeam(args.team, False)
if args.regional:
    lookUpRegional(args.regional)

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

    while found != True and sSize < 2750:
        payload = {"page": "searchresults", "skip_events": "0", "skip_teams": sSize, "programs": "FRC",
                   "season_FRC": "2014", "reports": "teams", "area": "All", "results_size": "250"}
        pageRequest = requests.post(MYFIRST_SITE_ROOT, data=payload)
        soup = BeautifulSoup(pageRequest.content)
        links =  soup.find_all("a", href=re.compile("team_details"))

        for link in links: # find ALL THE HYPERLINKS!
            teamNumber = "".join([str(num) for num in link.find_all(text=True)])
            if int(teamNumber) == team:
                pageRequest = requests.get(MYFIRST_SITE_ROOT + link["href"])
                soup = BeautifulSoup(pageRequest.content)
                teamNicknameText = soup.find(text='Team Nickname')
                teamNickTag = teamNicknameText.findNext('td')
                teamNick = teamNickTag.get_text()
                print (team)
                print(teamNick)
                found = True
                teamInfo = {"number": teamNumber, "name": teamNick, "reviews": [], "matches": [], "photos": []}
                if force == True:
                    return teamInfo
                urllib2.urlopen("http://0.0.0.0:8080/teams", json.dumps(teamInfo))
                break

        if found == False:
            print("Not found on page " + str(sSize / 250) + ".")
            sSize += 250
            print("Now checking page " + str(sSize / 250) + ".")

    if found == False:
        print "Team # " + str(team) + " does not exist."
        print "Please ensure it was entered correctly."

# Function: LookUpRegional
# ------------------------
# Recieves a regional name and searches through the FRC website for the regional that matches.
# If found, it enters into the regional page, parses the participating team, and extracts the team
# data  
def lookUpRegional(regional):
    sSize = 0
    found = False

    while found != True:
        payload = {"page": "searchresults", "skip_events": sSize, "skip_teams": "0", "programs": "FRC",
                   "season_FRC": "2013", "season_FTC": "2011", "season_FLL": "2011", "season_JFLL": "2011",
                   "season_FP": "2011", "reports": "events", "area": "All", "results_size": "250"}
        pageRequest = requests.post(MYFIRST_SITE_ROOT, data=payload)
        soup = BeautifulSoup(pageRequest.content)

        #For loop searches the FRC front page for team links
        for link in soup.find_all("a"):
            if link.get("href") and "event_details" in link.get("href"):
                regionalName = "".join([str(num) for num in link.find_all(text=True)])

                #TODO: need to give team and matches as children to the DB
                #If statement that test if the target regional is found
                if regionalName == args.regional:
                    pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
                    soup = BeautifulSoup(pageRequest.content)
                    
                    #Now inside regional specific page
                    #For loop implies target regional has been found and it looks at the FRC page for the targetted regional
                    #in order to find the
                    for link in soup.find_all('a'):
                        if link.get("href") and "event_teamlist" in link.get("href"):
                            regionalTeamLink = MYFIRST_SITE_ROOT + link.get("href")
                            # print regionalTeamLink
                            pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
                            soup = BeautifulSoup(pageRequest.content)

                            #For loop that searches the teams registered for the target regional
                            for link in soup.find_all('a'):
                                if link.get("href") and "team_details" in link.get("href"):
                                    teamNumber = "".join([str(num) for num in link.find_all(text=True)])
                                    num = int(teamNumber)
                                    print(type(num) == int)
                                    mapOfTeams[teamNumber] = lookUpTeam(num, True)
                                   # pageRequest = requests.get(MYFIRST_SITE_ROOT + link.get("href"))
                                   # soup = BeautifulSoup(pageRequest.content)
                                   # teamNicknameText = soup.find(text='Team Nickname')
                                   # teamNick = teamNicknameText.findNext('td')

                        #This if block looks for the link to the page that holds all the match results.
                        if link.get("href") and "matchresults" in link.get("href"):
                            regionalTeamLink = link.get("href")
                            pageRequest = requests.get(regionalTeamLink)
                            soup = BeautifulSoup(pageRequest.content)
                            startDate = soup.find('table', bgcolor="black").find("tbody").findNext("tr").findNext("td").findNext("td")

                            #This Gets Important Info about teams matches and Match numbers
                            rows = soup.find('table', style="background: black none repeat scroll 0% 50%; -moz-background-clip: initial; -moz-background-origin: initial; -moz-background-inline-policy: initial; width: 100%;").find("tbody").find_all("tr")
                            matches = []
                            num = ""
                            timeM = ""
                            skipThree = 0
                            redScore = 0
                            blueScore = 0

                            #For loop that searches the table of scores for the red alliance score and the blue alliance score
                            for row in rows:
                                #We need to skip the first three rows
                                if skipThree < 3:
                                    skipThree += 1
                                    print skipThree
                                    continue

                                cells = row.find_all("td")
                                redTeam = []
                                blueTeam = []

                                timeM = "".join(str(cells[0].get_text())) #What is AM and PM?
                                num = "".join(str(cells[1].get_text()))
                                redTeam.append(mapOfTeams[int(cells[2].get_text())])
                                redTeam.append(mapOfTeams[int(cells[3].get_text())])
                                redTeam.append(mapOfTeams[int(cells[4].get_text())])
                                blueTeam.append(mapOfTeams[int(cells[5].get_text())])
                                blueTeam.append(mapOfTeams[int(cells[6].get_text())])
                                blueTeam.append(mapOfTeams[int(cells[7].get_text())])
                                redScore = int(cells[8].get_text())
                                blueScore = int(cells[9].get_text())

                                matches.append({"number": num, "type": "Qualifications", "time": "".join(timeM), "red": redTeam,
                                    "blue": blueTeam, "rScore": redScore, "bScore": blueScore, "winner": "red"})

                            #FIXME:DUPLICATES IN MATCHES.
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
def scrapeTeam(teamNumber = "", teamNumberArray = {}):
    teamURL = "http://www.thebluealliance.com/api/v1/teams/show?teams="
    if teamNumber != "" and teamNumberArray != {}:
        raise TypeError("Cannot define a teamNumber and a teamNumberArray")
    elif teamNumber == "":
        teamJSONArray = []
        url = ""
        teamAdded = 0
        for teams in teamNumberArray:
            url = url + str(teams).strip() + ","
            teamAdded += 1
            if teamAdded > 50:
                raw = requests.get(teamURL + url[:-1])
                teamJSONArray.append(json.loads(raw.content))
                teamAdded = 0
                matchURLSuffix = ""
        raw = requests.get(teamURL + url[:-1])
        teamJSONArray.append(json.loads(raw.content))
        teamJSONArray = teamJSONArray.pop(0)
        for data in teamJSONArray:
            teamNum = data["team_number"]
            teamNick = data["nickname"]
            teamInfo = {"number": teamNum, "name": teamNick, "reviews": [], "matches": [], "photos": []}
            requests.post("http://0.0.0.0:8080/teams", json.dumps(teamInfo))
    else:
        raw = requests.get(teamURL + str(teamNumber).strip())
        data = json.loads(raw.content)
        data = data.pop(0)
        teamNum = data["team_number"]
        teamNick = data["nickname"]
        teamInfo = {"number": teamNum, "name": teamNick, "reviews": [], "matches": [], "photos": []}
        requests.post("http://0.0.0.0:8080/teams", json.dumps(teamInfo))

# Function: scrapeRegional
# ------------------------
# Employs the Blue Alliance API and generates the JSON data for the input regional and year. It 
# then formats the data and dumps it to the server. The setTeam flags determines whether or no the teams
# encountered by the function should be sent to the teams page of Scoutmaster
def scrapeRegional(regionalYear, regionalName, setTeam = False):
    eventsURL = "http://www.thebluealliance.com/api/v1/events/list?year=" + str(regionalYear).strip()
    keyData = json.loads(requests.get(eventsURL).content)
    index = 0
    
    while keyData[index]['name'] != regionalName:
        index += 1

    regionalURL = "http://www.thebluealliance.com/api/v1/event/details?event=" + str(keyData[index]['key']).strip()
    pageData = requests.get(regionalURL)
    teamData = json.loads(pageData.content)
    listOfMatches = teamData['matches'
    requestsArray = []
    matchURLSuffix = ""
    matchesAdded = 0

    for matches in listOfMatches:
        matchURLSuffix = matchURLSuffix + str(matches) + ","
        matchesAdded += 1
        if matchesAdded > 50:
            pageData = requests.get("http://www.thebluealliance.com/api/v1/match/details?match=" + matchURLSuffix[:-1])
            requestsArray.append(json.loads(pageData.content))
            matchesAdded = 0
            matchURLSuffix = ""
    pageData = requests.get("http://www.thebluealliance.com/api/v1/match/details?match=" + matchURLSuffix[:-1])
    requestsArray.append(json.loads(pageData.content))
    matchArray = []
    
    for matchData in requestsArray:
        print matchData
        for match in matchData:
            num = int(match["match_number"])
            level = str(match["competition_level"])
            redTeam = []
            blueTeam = []
            if setTeam:
                scrapeTeam(teamNumberArray = match["alliances"]["red"]["teams"])
                scrapeTeam(teamNumberArray = match["alliances"]["blue"]["teams"])

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

    jsonData = {"location": regionalName, "matches": matchArray, "year":int(regionalYear)}
    requests.post("http://0.0.0.0:8080/regionals", json.dumps(jsonData))