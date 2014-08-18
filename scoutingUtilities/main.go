//Package scoutingUtilities is used for easily getting and sending data related
//to an FRC competition to the Scoutmaster9000 server. Majority if not all of
//the data is pulled from the The Blue Alliance API.
package main

//Requires at least go1.1, the higher the version the better
/********************************************************************************
*TODO: Possibly Refactor all synchronization to use sync.Wait.                 *
*TODO: Use the response structs to get initial data then marshal into our own  *
       struct which would make the way we reference fields way more consistent.*
*TODO: Update README.md with instructions on how to use this.                  *
********************************************************************************/
import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/broluwo/Scoutmaster9000/structs" //<-FIXME:If ownership is changes over to blair-robot-project this obviously needs to be switched
	"github.com/codegangsta/cli"                 //<-Dependency can be removed but at cost of reworking and lack of prettiness
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
		cli.StringFlag{Name: "server, s", Value: "http://0.0.0.0:8080", Usage: "Change location of server"},
		//Ability to force may be eschewed in so that to update a resource you must use PUT or PATCH
		cli.BoolFlag{Name: "force, f",
			Usage: "Overwrite teams that already exist in the Scoutmaster 9000 database.  By default, if a team or regional already exists, it will not be changed."},
		cli.IntFlag{Name: "year, y", Value: year, Usage: "Change year of which data is being searched for."},
	}
)

func main() {
	t := time.Now()
	app := cli.NewApp()
	app.Name = "Scoutmaster Utilities"
	app.Version = "0.1"
	app.Author = "Brian Oluwo"
	app.Usage = "Contains all the functions to pull the teams and team data from the FIRST Robotics official website."
	app.Commands = []cli.Command{
		{
			Name:        "scrapeTeam",
			ShortName:   "t",
			Description: "A team (or teams) to look up and add to the DB.Employs the Blue Alliance API.",
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
	log.Println(time.Since(t))
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
		log.Fatalln("There needs to be a team num included")
	}
	args := make([]string, length)
	var e error
	for j, i := range dataSlice {
		//Converts from string to int. If it errors out, there is malformed input.
		_, e = strconv.Atoi(i)
		logErr("The provided teamNum needs to be an integer ->"+i, e)
		args[j] = teamPrefix + i
	}
	//Allows for any amount of input to a sane degree.
	scrapeTeam(args...)
}

//scrapeTeam takes a slice of team keys and sends them off to be retrieved and posted, blocking to print the results of their conquests
//Could switch to a sync.WaitGroup
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

func getData(url string, resc chan<- string, errc chan<- error) {
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
		t := structs.Team{Force: force}
		logErr("Ensure Team struct still matches up with data.", json.Unmarshal(data, &t))
		if t.Number == 0 {
			errc <- errors.New("Provided Team Number does not exist")
		} else {
			sendTeamData(t, resc, errc)
		}
	} else if strings.Contains(url, regionalURL) {
		ev := structs.EventResponse{}
		logErr("Ensure EventResponse struct still matches up with TBA API.", json.Unmarshal(data, &ev))
		r := structs.Regional{Location: ev.Name, Year: ev.Year}
		r.Matches, r.WinnerArray = getMatchAndWinnerData(ev.Key)
		go sendRegionalData(r, resc, errc)
		teamsToScrape := make([]string, 0, len(r.WinnerArray))
		for value := range r.WinnerArray {
			teamsToScrape = append(teamsToScrape, value)
		}
		scrapeTeam(teamsToScrape...)
	}
}

func getMatchAndWinnerData(eventKey string) ([]structs.Match, map[string][3]int) {
	url := regionalURL + eventKey + "/matches"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add(headerName, headerValue)
	res, err := client.Do(req)
	logErr("Request Does Not Seem to Have Gone Through. Try Again Later.", err)
	matches := []structs.Match{}
	defer res.Body.Close()
	var data []json.RawMessage
	//Do this instead of io.ReadAll so we don't need contiguous mem
	logErr("Ensure TBA API is returning an array of matches.", json.NewDecoder(res.Body).Decode(&data))
	for _, thing := range data {
		match := structs.MatchResponse{}
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

		m := structs.Match{Number: match.MatchNumber,
			Type:      match.CompLevel,
			Red:       redTeams,
			Blue:      blueTeams,
			RedScore:  int(match.Alliances.Red.Score),
			BlueScore: int(match.Alliances.Blue.Score)}
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
	//There should be a way cleaner way of doing this, could probably do it in the
	//server code.
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
		log.Panicln(err)
	}
}

//This method logs a string if there is an error.
//It's not really necessary but it helps readability
func logErr(s string, err error) {
	if err != nil {
		log.Fatalf("%v,%v\n", s, err)
	}
}

func sendTeamData(team structs.Team, resc chan<- string, errc chan<- error) {
	//this is so freaking ugly
	res, err := http.Post(serverLocation+"/team/", "application/json",
		bytes.NewReader(tossMarshalErr(json.Marshal(team)))) //If it panics here it was "Unable to encode Team struct."
	if err != nil {
		log.Println("Request didn't go through for:", team.Number, ". Please check server availability(config/location/online).")
		errc <- err
	} else if res != nil {
		resc <- res.Status
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
	val, err := structs.RegionalKeyMap[key]
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
			ev := structs.EventResponse{}
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

func sendRegionalData(r structs.Regional, resc chan<- string, errc chan<- error) {
	//If it panics here it was "Unable to encode Regional struct."
	res, err := http.Post(serverLocation+"/regional/", "application/json", bytes.NewReader(tossMarshalErr(json.Marshal(r))))

	if err != nil {
		log.Println("POST didn't go through for:", r.Location, ". Please check server config/location.")
		errc <- err
	} else if res != nil {
		resc <- res.Status
	}
}
func listRegional(c *cli.Context) {
	for k := range structs.RegionalKeyMap {
		println(k)
	}
}
