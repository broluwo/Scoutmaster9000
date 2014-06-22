package main

//Requires at least go1.1, the higher the version the better
/********************************************************************************
 *TODO: Possibly Refactor all synchronization to use sync.Wait.                 *
 *TODO: Update README.md with instructions on how to use this.                  *
 ********************************************************************************/
import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli" //<-Dependency can be removed but at cost of reworking and lack of prettiness
)

//Utilities package scout-util.py contains all the functions to pull the teams and team data from the Blue Alliance
//Any structs can be found in structs.go
const (
	teamURL     = "http://www.thebluealliance.com/api/v2/team/"
	regionalURL = "http://www.thebluealliance.com/api/v2/event/"  //"http://www.thebluealliance.com/api/v1/event/details?event="
	eventsURL   = "http://www.thebluealliance.com/api/v2/events/" //"http://www.thebluealliance.com/api/v1/events/list?year="
	teamPrefix  = "frc"
	headerName  = "X-TBA-App-Id"
	headerValue = "frc449:Scoutmaster_Utilities:v0.2"
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
	app.Author = "Brian Oluwo"
	app.Usage = "Contains all the functions to pull the teams and team data from the FIRST Robotics official website."
	app.Commands = []cli.Command{
		{
			Name:        "scrapeTeam",
			ShortName:   "t",
			Description: "A team to look up and add.Employs the Blue Alliance API and generates the JSON data for the input team given by the team's number. The teamNumber should be the official team number meaning it must be in the form of frc###. It will then dump everything to the Scoutmaster servers.",
			Flags:       globalFlags,
			Usage:       "scoutingUtilities t 449 ### ### ### ... Append as many teams as you like.",
			Action:      handleTeam,
		},
		{
			Name:        "scrapeRegional",
			ShortName:   "r",
			Description: "A regional to look up and add",
			Flags:       globalFlags,
			Usage:       "scoutingUtilities r 'Washington DC Regional' ",
			Action:      handleRegional,
		},
		{
			Name:        "listRegionals",
			ShortName:   "lr",
			Description: "Lists cached Regionals.",
			Usage:       "scoutingUtilities lr",
			Action:      listRegional,
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
	if c.String("server") != serverLocation {
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

//scrapeTeam takes a slice of team keys and sends them off to be retrieved and posted, blocking to print the results of their conquests
func scrapeTeam(teamNums ...string) {
	length := len(teamNums)
	resc, errc := make(chan string), make(chan error)
	for _, team := range teamNums {
		go getData(teamURL+team, resc, errc)
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

func getData(url string, resc chan string, errc chan error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add(headerName, headerValue)
	res, err := client.Do(req)
	logErr("Request Does Not Seem to Have Gone Through. Try Again Later.", err)
	defer res.Body.Close()
	var data json.RawMessage
	//Do this instead of io.ReadAll so we don't need contiguous mem
	logErr("Malformed Input. Check TBA API for changes.", json.NewDecoder(res.Body).Decode(&data))
	if strings.Contains(url, teamURL) {
		t := Team{Force: force}
		logErr("Ensure Team struct still matches up with data.", json.Unmarshal(data, &t))
		if t.Number == 0 {
			log.Println("Provided Team Number does not exist")
		}
		//We don't do an else because we need the program to return an err so that scrapeTeam does not block infinitely
		sendTeamData(t, resc, errc)
	} else if strings.Contains(url, regionalURL) {
		ev := eventResponse{}
		logErr("Ensure EventResponse struct still matches up with TBA API.", json.Unmarshal(data, &ev))
		r := Regional{Location: ev.Name, Year: ev.Year}
		r.Matches, r.WinnerArray = getMatchAndWinnerData(ev.Key)
		go sendRegionalData(r, resc, errc)
		teamsToScrape := make([]string, 0, len(r.WinnerArray))
		for value := range r.WinnerArray {
			teamsToScrape = append(teamsToScrape, value)
		}
		scrapeTeam(teamsToScrape...)
	}
}

func getMatchAndWinnerData(eventKey string) ([]Match, map[string][3]int) {
	url := regionalURL + eventKey + "/matches"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add(headerName, headerValue)
	res, err := client.Do(req)
	logErr("Request Does Not Seem to Have Gone Through. Try Again Later.", err)
	matches := []Match{}
	defer res.Body.Close()
	var data []json.RawMessage
	//Do this instead of io.ReadAll so we don't need contiguous mem
	logErr("Ensure TBA API is returning an array of matches.", json.NewDecoder(res.Body).Decode(&data))
	for _, thing := range data {
		match := matchResponse{}
		logErr("Ensure TBA API and matchResponse struct are reconciable.", json.Unmarshal(thing, &match))
		redTeams := make([]int, 3)
		blueTeams := make([]int, 3)
		//Convert keys e.g. "frc449" to ints e.g 449
		for i, j := range match.Alliances.Red.Teams {
			redTeams[i], err = strconv.Atoi(strings.TrimPrefix(j, teamPrefix))
			if err != nil {
				logErr("Couldn't convert frc team into an int.", err)
			}
		}
		//Convert keys e.g. "frc449" to ints e.g 449
		for i, j := range match.Alliances.Blue.Teams {
			blueTeams[i], err = strconv.Atoi(strings.TrimPrefix(j, teamPrefix))
			if err != nil {
				logErr("Couldn't convert frc team into an int.", err)
			}
		}

		m := Match{Number: match.MatchNumber, Type: match.CompLevel, Red: redTeams, Blue: blueTeams, RedScore: int(match.Alliances.Red.Score), BlueScore: int(match.Alliances.Blue.Score)}
		if m.RedScore > m.BlueScore {
			m.Winner = "red"
		} else if m.RedScore < m.BlueScore {
			m.Winner = "blue"
		} else {
			m.Winner = "tie"
		}
		matches = append(matches, m)
	}
	//Winner Code Below
	blue := map[string]int{"blue": 0, "red": 2, "tie": 1}
	red := map[string]int{"red": 0, "blue": 2, "tie": 1}
	var winnerArray = make(map[string][3]int)
	for _, match := range matches {
		for _, team := range match.Blue {
			a := winnerArray[teamPrefix+strconv.Itoa(team)]
			a[blue[match.Winner]]++
			winnerArray[teamPrefix+strconv.Itoa(team)] = a
		}
		for _, team := range match.Red {
			a := winnerArray[teamPrefix+strconv.Itoa(team)]
			a[red[match.Winner]]++
			winnerArray[teamPrefix+strconv.Itoa(team)] = a
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

func sendTeamData(team Team, resc chan string, errc chan error) {
	//this is so freaking ugly
	res, err := http.Post(serverLocation+"/teams", "application/json",
		bytes.NewReader(tossMarshalErr(json.Marshal(pythonWrapTeam(team))))) //If it panics here it was "Unable to encode Team struct."

	if err != nil {
		log.Println("Request didn't go through for:", team.Number, ". Please check server availability(config/location/online).")
		errc <- err
	}
	if res != nil {
		resc <- res.Status
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

//Regional name must be surrounded by quotes because spaces are confusing
//scrapeRegional takes a slice of regionalNames and sends them off to be retrieved and posted, blocking to print the results of their conquests
func scrapeRegional(regionalNames []string) {
	length := len(regionalNames)
	resc, errc := make(chan string), make(chan error)
	for i := 0; i < length; i++ {
		getData(returnRegionalURL(regionalNames[i]), resc, errc)
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
func returnRegionalURL(key string) string {
	val, err := regionalKeyMap[key]
	if !err {
		//This means the key wasn't in our cached ones and we need to make a request
		log.Println("Regional not found locally, checking online...")
		url := eventsURL + strconv.Itoa(year)
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add(headerName, headerValue)
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
				return regionalURL + ev.Key
			}
		}
		log.Println("Regional Couldn't be found. Is this key correct? -> " + key)
		os.Exit(3)
	}
	return regionalURL + strconv.Itoa(year) + val
}

func sendRegionalData(r Regional, resc chan string, errc chan error) {
	res, err := http.Post(serverLocation+"/regionals", "application/json", bytes.NewReader(tossMarshalErr(json.Marshal(r)))) //If it panics here it was "Unable to encode Regional struct."

	if err != nil {
		log.Println("Post didn't go through for:", r.Location, ". Please check server config/location.")
		errc <- err
	} else if res != nil {
		resc <- res.Status
	}
}
func listRegional(c *cli.Context) {
	for k := range regionalKeyMap {
		println(k)
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
	Alliances           []finalAlliance   `json:"alliances"`             //If we have alliance selection data for this event, this contains a JSON array of the alliances. The captain is the first team, followed by their picks, in order.
	Key                 string            `json:"key"`                   //TBA event key with the format yyyy[EVENT_CODE], where yyyy is the year, and EVENT_CODE is the event code of the event.	2010sc
	Name                string            `json:"name"`                  //Official name of event on record either provided by FIRST or organizers of offseason event.	Palmetto Regional
	ShortName           string            `json:"short_name"`            //name but doesn't include event specifiers, such as 'Regional' or 'District'.	Palmetto
	EventCode           string            `json:"event_code"`            //Event short code.	SC
	EventTypeString     string            `json:"event_type_string"`     //A human readable string that defines the event type.	'Regional', 'District', 'District Championships', 'District Championship','Championship Division', 'Championship Finals', 'Offseason','Preseason', '--'
	EventType           int               `json:"event_type"`            //An integer that represents the event type as a constant.	List of constants to event type
	EventDistrictString string            `json:"event_district_string"` //A human readable string that defines the event's district.	'Michigan', 'Mid Atlantic', null (if regional)
	EventDistrict       int               `json:"event_district"`        //An integer that represents the event district as a constant.	List of constants to event district
	Year                int               `json:"year"`                  //Year the event data is for.	2010
	Location            string            `json:"location"`              //Long form address that includes city, and state provided by FIRST	Clemson, SC
	VenueAddress        string            `json:"venue_address"`         //Address of the event's venue, if available. Line breaks included.	Long Beach Arena\n300 East Ocean Blvd\nLong Beach, CA 90802\nUSA
	Website             string            `json:"website"`               //The event's website, if any.	http://www.firstsv.org
	Official            bool              `json:"official"`              //Whether this is a FIRST official event, or an offseaon event.	true
	Teams               []teamResponse    `json:"teams"`                 //List of team models that attended the event
	Webcast             []json.RawMessage `json:"webcast"`               //If the event has webcast data associated with it, this contains JSON data of the streams
	EndDate             string            `json:"end_date"`              //Day the event ends in string format	"2014-03-29"
	StartDate           string            `json:"start_date"`            //Day the event starts in string format	"2014-03-27"
	//facebook_eid null
}
type finalAlliance struct {
	Declines []string `json:"declines"`
	Picks    []string `json:"picks"`
}
type matchResponse struct {
	Key         string         `json:"key"`          //TBA event key with the format yyyy[EVENT_CODE]_[COMP_LEVEL]m[MATCH_NUMBER], where yyyy is the year, and EVENT_CODE is the event code of the event, COMP_LEVEL is (qm, ef, qf, sf, f), and MATCH_NUMBER is the match number in the competition level. A set number may append the competition level if more than one match in required per set .	2010sc_qm10, 2011nc_qf1m2
	CompLevel   string         `json:"comp_level"`   //The competition level the match was played at.	qm, ef, qf, sf, f
	SetNumber   int            `json:"set_number"`   //The set number in a series of matches where more than one match is required in the match series.	2010sc_qf1m2, would be match 2 in quarter finals 1.
	MatchNumber int            `json:"match_number"` //The match number of the match in the competition level.	2010sc_qm20
	Alliances   matchAlliances `json:"alliances"`    //A list of alliances, the teams on the alliances, and their score.
	EventKey    string         `json:"event_key"`    //Event key of the event the match was played at.	2011sc
	Videos      []videoLink    `json:"videos"`       //JSON array of videos associated with this match and corresponding information	"videos": [{"key": "xswGjxzNEoY", "type": "youtube"}, {"key": "http://videos.thebluealliance.net/2010cmp/2010cmp_f1m1.mp4", "type": "tba"}]
	TimeString  string         `json:"time_string"`  //Time string for this match, as published on the official schedule. Of course, this may or may not be accurate, as events often run ahead or behind schedule	11:15 AM
	Time        string         `json:"time"`         //UNIX timestamp of match time, as taken from the published schedule	1394904600
}
type videoLink struct {
	Type string `json:"type"`
	Key  string `json:"key"`
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
	Number    int    `json:"number"`
	Type      string `json:"type"`
	Red       []int  `json:"red"` //These should be strings in next v
	Blue      []int  `json:"blue"`
	RedScore  int    `json:"rScore"`
	BlueScore int    `json:"bScore"`
	Winner    string `json:"winner"`
}

//Regional How the python server takes regional
type Regional struct {
	Location    string            `json:"location"`
	Matches     []Match           `json:"matches"`
	WinnerArray map[string][3]int `json:"winnerCount"`
	Year        int               `json:"year"`
}
