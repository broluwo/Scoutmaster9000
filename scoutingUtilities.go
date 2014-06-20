package main

/******************************
*TODO: FIX API MISNOMERS   *
******************************/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli" //<-Dependency can be removed but at cost of a little reworking and lack of prettiness
)

//Utilities package scout-util.py contains all the functions to pull the teams and team data from the Blue Alliance
//Any structs can be found in structs.go
const (
	//TODO: Update this to v2 of API
	teamURL        = "http://www.thebluealliance.com/api/v1/teams/show?teams="
	regionalURL    = "http://www.thebluealliance.com/api/v1/event/details?event="
	eventsURL      = "www.thebluealliance.com/api/v2/event/" //"http://www.thebluealliance.com/api/v1/events/list?year="
	matchURL       = "http://www.thebluealliance.com/api/v1/match/details?match="
	teamPrefix     = "frc"
	teamPacketSize = 15
)

//Similar to constants but are changeable by flags.
var (
	force          = false
	serverLocation = "http://0.0.0.0:8080"
	year           = time.Now().Year()

	globalFlags = []cli.Flag{
		cli.StringFlag{"server, s", "http://0.0.0.0:8080", "Change location of server"},
		cli.BoolFlag{"force, f", "Overwrite teams that already exist in the Scoutmaster 9000 database.  By default, if a team or regional already exists, it will not be changed."},
		cli.IntFlag{"year, y", year, "Change location of server"},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "Scoutmaster Utilities"
	app.Version = "0.1"
	app.Usage = "Contains all the functions to pull the teams and team data from the FIRST Robotics official website."
	app.Commands = []cli.Command{
		{
			Name:        "scrapeTeam",
			ShortName:   "t",
			Description: "A team to look up and add",
			Flags:       globalFlags,
			Usage:       "Employs the Blue Alliance API and generates the JSON data for the input team given by the team's number. The teamNumber should be the official team number meaning it must be in the form of frc###. It will then dump everything to the Scoutmaster servers.",
			Action:      handleTeam,
		},
		{
			Name:        "scrapeRegional",
			ShortName:   "r",
			Description: "A team to look up and add",
			Flags:       globalFlags,
			Usage:       "Employs the Blue Alliance API and generates the JSON data for the input team given by the team's number. The teamNumber should be the official team number meaning it must be in the form of frc###. It will then dump everything to the Scoutmaster servers.",
			Action:      handleRegional,
		},
	}
	app.Action = func(c *cli.Context) {
		log.Println("You didn't specify a command. Type <command> help to see options.")
	}
	err := app.Run(os.Args)
	if err != nil {
		os.Exit(-1)
	}
}

//I feel as if there should be a way to loop through these, instead of hard coding. It's an easy fix either way
//The int being returned is telling me the number of args i need to truncate before processing the input
func checkGlobalFlags(c *cli.Context) int {
	var index int
	if c.Bool("force") { // If the force flag has been set
		force = true
	}
	if c.String("server") != "http://0.0.0.0:8080" {
		u, err := url.Parse(c.Args()[index])
		logErr("URL invalid. Ensure you attatched a scheme, i.e. http:// or https://", err)
		serverLocation = u.String()
		index++
	}
	if c.Int("year") != year {
		var err error
		year, err = strconv.Atoi(c.Args()[index])
		logErr("The provided year needs to be an convertible integer like 2015.", err)
		index++
	}
	return index
}

// All of the arguments should be ints...
//handleTeams ensures that everything is in order before going ahead to the request section
func handleTeam(c *cli.Context) {
	//Get rid of flag arguments before parsing
	dataSlice := c.Args()[checkGlobalFlags(c):]
	length := len(dataSlice)
	//Check if there is enough params passed through before allocating mem
	if length < 1 {
		log.Println("There needs to be a team num included")
		os.Exit(3)
	}
	args := make([]string, length)
	var e error
	for j, i := range dataSlice {
		//Converts from string to int. If it errors out, there is malformed input.
		//FIXME: Waste of a call? Could do a interface type assertion...
		_, e = strconv.Atoi(i)
		logErr("The provided teamNum needs to be an integer ->"+i, e)
		args[j] = teamPrefix + i
	}
	//Allows for slice variadic
	scrapeTeam(args...)
}

// Function: scrapeTeam
// ------------------------
// Employs the Blue Alliance API and generates the JSON data for the input team given by the team's
// number. The teamNumber should be the official team number meaning it must be in the form of frc###. It
// will then dump everything to the Scoutmaster servers.
func scrapeTeam(teamNums ...string) {
	length := len(teamNums)
	resc, errc := make(chan string), make(chan error)
	for i := 0; i <= length; i += teamPacketSize {
		//Again using a variadic because it's nice
		go getData(returnTeamURL(teamNums[i:int(math.Min(float64((i+teamPacketSize)), float64(length)))]...), resc, errc)
	}
	for i := 0; i < length; i++ { //Loop length amount of times to make sure i have gotten a response from each goroutine before proceeding
		select { // Force this for loop to wait for a response from either errc(error channel) or resc (the response channel)
		case res := <-resc: // If i get a response status print it out (this also tells me that this particular goroutine is done
			log.Println(res)
		case err := <-errc: // If i get an error print it out along with time and date
			log.Println(err)
		}
	}
}

//Accepts a slice so that changing that packetSize cascades down, with minimal work
func returnTeamURL(vals ...string) string {
	result := teamURL
	for _, i := range vals {
		result += i + ","
	}
	//Get rid of last comma
	return result[:len(result)-1]
}

func getData(url string, resc chan string, errc chan error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-TBA-App-Id", "frc449:Scoutmaster_Utilities:v2")
	res, err := client.Do(req)
	logErr("Request Does Not Seem to Have Gone Through. Try Again Later.", err)
	if strings.Contains(url, "teams") {
		teams := []Team{}
		defer res.Body.Close()
		var data []json.RawMessage
		//Do this instead of io.ReadAll so we don't need contiguous mem
		panicErr(json.NewDecoder(res.Body).Decode(&data))
		for _, thing := range data {
			t := Team{}
			t.Force = force
			panicErr(json.Unmarshal(thing, &t))
			teams = append(teams, t)
		}
		sendTeamData(teams, resc, errc)
	} else if strings.Contains(url, "event") {
		defer res.Body.Close()
		var data json.RawMessage
		//Do this instead of io.ReadAll so we don't need contiguous mem
		panicErr(json.NewDecoder(res.Body).Decode(&data))
		ev := eventResponse{}
		panicErr(json.Unmarshal(data, &ev))
		sendRegionalData(packageRegionalData(&ev), resc, errc)
	}
}

//Returns the struct to be sent to the server
func packageRegionalData(ev *eventResponse) Regional {
	r := Regional{Location: ev.Name, Year: ev.Year}
	r.Matches, r.WinnerArray = getMatchAndWinnerData(ev.Key)
	//Now take every key from winner's array send it off to scrape team for processing
	// teamsURL := regionalURL + ev.Key + "/teams"
	// for team in teamsToScrape:
	// scrapeTeam(matchArray, teamNumberArray = team, force = force)*/
	fmt.Println(ev.Name)
	return r
}

func getMatchAndWinnerData(eventKey string) ([]Match, map[string][3]int) {
	url := regionalURL + eventKey + "/matches"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-TBA-App-Id", "frc449:Scoutmaster_Utilities:v2")
	res, err := client.Do(req)
	logErr("Request Does Not Seem to Have Gone Through. Try Again Later.", err)
	matches := []Match{}
	defer res.Body.Close()
	var data []json.RawMessage
	//Do this instead of io.ReadAll so we don't need contiguous mem
	panicErr(json.NewDecoder(res.Body).Decode(&data))
	for _, thing := range data {
		match := matchResponse{}
		panicErr(json.Unmarshal(thing, &match))
		//MatchNum may not be suitable for conversion
		a, e := strconv.Atoi(match.MatchNumber[:])
		logErr("MatchNum can't be converted", e)

		m := Match{Number: a, Type: match.CompLevel, Red: match.Alliances.Red.Teams, Blue: match.Alliances.Blue.Teams}
		if int(match.Alliances.Red.Score) > int(match.Alliances.Blue.Score) {
			m.Winner = "red"
		} else if int(match.Alliances.Red.Score) < int(match.Alliances.Blue.Score) {
			m.Winner = "blue"
		} else {
			m.Winner = "tie"
		}
		matches = append(matches, m)
	}

	blue := map[string]int{"blue": 0, "red": 2, "tie": 1}
	red := map[string]int{"red": 0, "blue": 2, "tie": 1}
	var winnerArray = make(map[string][3]int)
	for _, match := range matches {
		for _, team := range match.Blue {
			a := winnerArray[team]
			a[blue[match.Winner]]++
			winnerArray[team] = a
		}
		for _, team := range match.Red {
			a := winnerArray[team]
			a[red[match.Winner]]++
			winnerArray[team] = a
		}
	}
	return matches, winnerArray
}

//This method panics out if there is a non-nil error
//It's not really necessary but it helps readability
func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

//This method logs a string if there is an error.
//It's not really necessary but it helps readability
func logErr(s string, err error) {
	if err != nil {
		log.Println(s)
		os.Exit(3)
	}
}

func sendTeamData(teams []Team, resc chan string, errc chan error) {
	for _, this := range teams {
		//this is so freaking ugly
		res, err := http.Post(serverLocation+"/teams", "application/json",
			bytes.NewReader(tossMarshalErr(json.Marshal(pythonWrapTeam(this)))))

		if err != nil {
			log.Println("Request didn't go through for:", this.Number, ". Please check server config/location.")
			errc <- err
		}
		if res != nil {
			resc <- res.Status
		}
	}
}

func pythonWrapTeam(t Team) pythonTeamWrapper {
	return pythonTeamWrapper{
		Name:    &t.Name,
		Number:  &t.Number,
		Force:   &t.Force,
		Reviews: &t.Reviews,
		Matches: &t.Matches,
		Photos:  &t.Photos,
	}
}

func tossMarshalErr(a []byte, err error) []byte {
	panicErr(err)
	return a
}

// All of the arguments should be strings...
//handleRegional ensures that everything is in order before going ahead to the request section
func handleRegional(c *cli.Context) {
	//Get rid of flag arguments before parsing
	dataSlice := c.Args()[checkGlobalFlags(c):]
	//Check if there is enough params passed through before allocating mem
	if len(dataSlice) < 1 {
		log.Println("There needs to be a regional name included")
		os.Exit(3)
	}
	scrapeRegional(dataSlice)
}

//REgional name must be surrounded by quotes because spaces are confusing
func scrapeRegional(regionalNames []string) {
	length := len(regionalNames)
	resc, errc := make(chan string), make(chan error)
	for i := 0; i < length; i++ {
		key, _ := returnRegionalURL(regionalNames[i])
		fmt.Println(key)
		//Will be used later for alternative path for getting data
		// if sig != 0 {
		// 	continue
		// }
		go getData(key, resc, errc)
	}
	for i := 0; i < length; i++ { //Loop length amount of times to make sure i have gotten a response from each goroutine before proceeding
		select { // Force this for loop to wait for a response from either errc(error channel) or resc (the response channel)
		case res := <-resc: // If i get a response status print it out (this also tells me that this particular goroutine is done)
			log.Println(res)
		case err := <-errc: // If i get an error print it out along with time and date
			log.Println(err)
		}
	}
}

//Accepts a slice so that changing that packetSize cascades down, with minimal work
func returnRegionalURL(key string) (string, int) {
	val, err := regionalKeyMap[key]
	if !err {
		// log.Println("Key Not Found", err)
		//This means the key wasn't in our cached ones and we need to make a request
		log.Println("Regional not found locally, checking online...")
		url := eventsURL + strconv.Itoa(year)
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("X-TBA-App-Id", "frc449:Scoutmaster_Utilities:v2")
		res, err := client.Do(req)
		defer res.Body.Close()
		logErr("Regional Key not found as Request Does Not Seem to Have Gone Through. Try Again Later.", err)
		var data []json.RawMessage
		//Do this instead of io.ReadAll so we don't need contiguous mem
		logErr("JSON couldn't marshall into struct. Check expected output versus struct format", json.NewDecoder(res.Body).Decode(&data))
		for _, thing := range data {
			ev := eventResponse{}
			logErr("Malformed Input, check TBA API return to ensure nothing has changed.", json.Unmarshal(thing, &ev))
			if ev.Name == key {
				//Stupid as at this point we have the object already, but for compatability lets go with it
				//Launch go func and skip the get data part
				regURL := regionalURL + ev.Key
				// go sendRegionalData(regURL)
				// return "",1
				return regURL, 0
			}
		}
	}
	return regionalURL + strconv.Itoa(year) + val, 0
}

func sendRegionalData(r Regional, resc chan string, errc chan error) {
	fmt.Printf("%v", r)

	res, err := http.Post(serverLocation+"/regionals", "application/json", bytes.NewReader(tossMarshalErr(json.Marshal(r))))

	if err != nil {
		log.Println("Post didn't go through for:", r.Location, ". Please check server config/location.")
		errc <- err
	}
	if res != nil {
		resc <- res.Status
	}
}

//This takes regional names and maps them to the keys TBA uses
//Since this is hard coded it is a faster but non safe way of checking values.
//If a new regional is ever added and isn't present or the keys change this has to be updated.
//All names and keys were retrieved from https://docs.google.com/spreadsheet/ccc?key=0ApRO2Yzh2z01dExFZEdieV9WdTJsZ25HSWI3VUxsWGc#gid=0 .
//An alternative way of doing this is to query "http://www.thebluealliance.com/api/v1/events/list?year=<yearGoesHere>"
//to get a list of all keys and check each returned json array's name param to see if it matches the one provided. Slower but safe(r).
var regionalKeyMap = map[string]string{
	"Alamo Regional sponsored by Rackspace Hosting":                  "txsa",
	"Autodesk Oregon Regional":                                       "orpo",
	"BAE Systems Granite State Regional":                             "nhma",
	"Bayou Regional":                                                 "lake",
	"Bedford FIRST Robotics District Competition":                    "mibed",
	"Boilermaker Regional":                                           "inwl",
	"Boston Regional":                                                "mabo",
	"Bridgewater-Raritan FIRST Robotics District Competition":        "njbrg",
	"Buckeye Regional":                                               "ohcl",
	"Central Valley Regional":                                        "cama",
	"Central Washington Regional":                                    "wase",
	"Chesapeake Regional":                                            "mdba",
	"Colorado Regional":                                              "code",
	"Connecticut Regional sponsored by UTC":                          "ctha",
	"Crossroads Regional":                                            "inth",
	"Dallas Regional":                                                "txda",
	"Detroit FIRST Robotics District Competition":                    "midet",
	"Festival de Robotique FRC a Montreal Regional":                  "qcmo",
	"Finger Lakes Regional":                                          "nyro",
	"Grand Blanc FIRST Robotics District Competition":                "migbl",
	"Greater Kansas City Regional":                                   "mokc",
	"Greater Toronto East Regional":                                  "onto",
	"Greater Toronto West Regional":                                  "onto",
	"Gull Lake FIRST Robotics District Competition":                  "migul",
	"Hatboro-Horsham FIRST Robotics District Competition":            "pahat",
	"Hawaii Regional sponsored by BAE Systems":                       "hiho",
	"Hub City Regional":                                              "txlu",
	"Inland Empire Regional":                                         "casb",
	"Israel Regional":                                                "ista",
	"Kettering University FIRST Robotics District Competition":       "miket",
	"Lake Superior Regional":                                         "mndu",
	"Las Vegas Regional":                                             "nvlv",
	"Lenape Seneca FIRST Robotics District Competition":              "njlen",
	"Livonia FIRST Robotics District Competition":                    "miliv",
	"Lone Star Regional":                                             "txho",
	"Los Angeles Regional":                                           "calb",
	"Michigan FRC State Championship":                                "micmp",
	"Mid-Atlantic Robotics FRC Region Championship":                  "mrcmp",
	"Midwest Regional":                                               "ilch",
	"Minnesota 10000 Lakes Regional":                                 "mnmi",
	"Minnesota North Star Regional":                                  "mnmi2",
	"Mount Olive FIRST Robotics District Competition":                "njfla",
	"New York City Regional":                                         "nyny",
	"North Carolina Regional":                                        "ncre",
	"Northern Lights Regional":                                       "mndu2",
	"Oklahoma Regional":                                              "okok",
	"Orlando Regional":                                               "flor",
	"Palmetto Regional":                                              "scmb",
	"Peachtree Regional":                                             "gadu",
	"Phoenix Regional":                                               "azch",
	"Pine Tree Regional":                                             "mele",
	"Pittsburgh Regional":                                            "papi",
	"Queen City Regional":                                            "ohic",
	"Razorback Regional":                                             "arfa",
	"Sacramento Regional":                                            "casa",
	"San Diego Regional":                                             "casd",
	"SBPLI Long Island Regional":                                     "nyli",
	"Seattle Regional":                                               "wase",
	"Silicon Valley Regional":                                        "casj",
	"Smoky Mountains Regional":                                       "tnkn",
	"South Florida Regional":                                         "flbr",
	"Spokane Regional":                                               "wach",
	"Springside - Chestnut Hill FIRST Robotics District Competition": "paphi",
	"St Joseph FIRST Robotics District Competition":                  "misjo",
	"St. Louis Regional":                                             "mosl",
	"TCnj FIRST Robotics District Competition":                       "njewn",
	"Traverse City FIRST Robotics District Competition":              "mitvc",
	"Troy FIRST Robotics District Competition":                       "mitry",
	"Utah Regional sponsored by NASA":                                "utwv",
	"Virginia Regional":                                              "vari",
	"Washington DC Regional":                                         "dcwa",
	"Waterford FIRST Robotics District Competition":                  "miwfd",
	"Waterloo Regional":                                              "onwa",
	"West Michigan FIRST Robotics District Competition":              "miwmi",
	"Western Canadian FRC Regional":                                  "abca",
	"Wisconsin Regional":                                             "wimi",
	"WPI Regional":                                                   "mawo",
}

//Inserted Lovingly with B@$# . If this prints, there should be no errors for the map insertion.

//Team is the struct that represents a team
type Team struct {
	Force  bool   `json:"force"`
	Number int    `json:"team_number"`
	Name   string `json:"nickname"`
	//Will need to be changed to it's own struct
	Reviews []string `json:"reviews"`
	//Will need to be changed to it's own struct
	Matches []int `json:"matches"`
	//Will/Should be corrected soon
	Photos []byte `json:"photos"`
}

//While still using a python server wrappers are needed
type pythonTeamWrapper struct {
	Force  *bool   `json:"force"`
	Number *int    `json:"number"`
	Name   *string `json:"name"`
	//Will need to be changed to it's own struct
	Reviews *[]string `json:"reviews"`
	//Will need to be changed to it's own struct
	Matches *[]int `json:"matches"`
	//Will/Should be corrected soon
	Photos *[]byte `json:"photos"`
}

//Current Representation of what TBA sends back.
//Wholly unnecessary but allows me to see what I'm working with
type teamResponse struct {
	Name     string   `json:"name"`
	Locality string   `json:"locality"`
	Number   int      `json:"team_number"`
	Region   string   `json:"region"`
	Key      string   `json:"key"`
	Country  string   `json:"country_name"`
	Website  string   `json:"website"`
	Nickname string   `json:"nickname"`
	Events   []string `json:"events"`
}

//Current Representation of what TBA sends back.
type eventResponse struct { //Comments go Description <TAB> Example
	Key                 string         `json:"key"`                   //TBA event key with the format yyyy[EVENT_CODE], where yyyy is the year, and EVENT_CODE is the event code of the event.	2010sc
	Name                string         `json:"name"`                  //Official name of event on record either provided by FIRST or organizers of offseason event.	Palmetto Regional
	ShortName           string         `json:"short_name"`            //name but doesn't include event specifiers, such as 'Regional' or 'District'.	Palmetto
	EventCode           string         `json:"event_code"`            //Event short code.	SC
	EventTypeString     string         `json:"event_type_string"`     //A human readable string that defines the event type.	'Regional', 'District', 'District Championships', 'District Championship','Championship Division', 'Championship Finals', 'Offseason','Preseason', '--'
	EventType           int            `json:"event_type"`            //An integer that represents the event type as a constant.	List of constants to event type
	EventDistrictString string         `json:"event_district_string"` //A human readable string that defines the event's district.	'Michigan', 'Mid Atlantic', null (if regional)
	EventDistrict       int            `json:"event_district"`        //An integer that represents the event district as a constant.	List of constants to event district
	Year                int            `json:"year"`                  //Year the event data is for.	2010
	Location            string         `json:"location"`              //Long form address that includes city, and state provided by FIRST	Clemson, SC
	VenueAddress        string         `json:"venue_address"`         //Address of the event's venue, if available. Line breaks included.	Long Beach Arena\n300 East Ocean Blvd\nLong Beach, CA 90802\nUSA
	Website             string         `json:"website"`               //The event's website, if any.	http://www.firstsv.org
	Official            bool           `json:"official"`              //Whether this is a FIRST official event, or an offseaon event.	true
	Teams               []teamResponse `json:"teams"`                 //List of team models that attended the event
	// matches	//List of match models for the event.
	// awards	//List of award models for the event.
	// webcast	If the event has webcast data associated with it, this contains JSON data of the streams
	// alliances []string	If we have alliance selection data for this event, this contains a JSON array of the alliances. The captain is the first team, followed by their picks, in order.
}
type matchResponse struct {
	Key         string            `json:"key"`          //TBA event key with the format yyyy[EVENT_CODE]_[COMP_LEVEL]m[MATCH_NUMBER], where yyyy is the year, and EVENT_CODE is the event code of the event, COMP_LEVEL is (qm, ef, qf, sf, f), and MATCH_NUMBER is the match number in the competition level. A set number may append the competition level if more than one match in required per set .	2010sc_qm10, 2011nc_qf1m2
	CompLevel   string            `json:"comp_level"`   //The competition level the match was played at.	qm, ef, qf, sf, f
	SetNumber   string            `json:"set_number"`   //The set number in a series of matches where more than one match is required in the match series.	2010sc_qf1m2, would be match 2 in quarter finals 1.
	MatchNumber string            `json:"match_number"` //The match number of the match in the competition level.	2010sc_qm20
	Alliances   matchAlliances    `json:"alliances"`    //A list of alliances, the teams on the alliances, and their score.
	EventKey    string            `json:"event_key"`    //Event key of the event the match was played at.	2011sc
	Videos      []json.RawMessage `json:"videos"`       //JSON array of videos associated with this match and corresponding information	"videos": [{"key": "xswGjxzNEoY", "type": "youtube"}, {"key": "http://videos.thebluealliance.net/2010cmp/2010cmp_f1m1.mp4", "type": "tba"}]
	TimeString  string            `json:"time_string"`  //Time string for this match, as published on the official schedule. Of course, this may or may not be accurate, as events often run ahead or behind schedule	11:15 AM
	Time        int               `json:"time"`         //UNIX timestamp of match time, as taken from the published schedule	1394904600
}
type matchAlliances struct {
	Red  alliance `json:"red"`
	Blue alliance `json:"blue"`
}
type alliance struct {
	Score int      `json:"score"`
	Teams []string `json:"teams"`
}

//The Match struct is how a match is represented
type Match struct {
	Number    int      `json:"number"`
	Type      string   `json:"type"`
	Red       []string `json:"red"`  //These might be ints
	Blue      []string `json:"blue"` //These might be ints
	RedScore  int      `json:"rScore"`
	BlueScore int      `json:"bScore"`
	Winner    string   `json:"winner"`
}

//Regional How the python server takes regional
type Regional struct {
	Location    string            `json:"location"`
	Matches     []Match           `json:"matches"`
	WinnerArray map[string][3]int `json:"winnerCount"`
	Year        int               `json:"year"`
}