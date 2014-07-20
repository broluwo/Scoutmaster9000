package structs

import "encoding/json"

//Structs.go simply holds all the structs that are being used

//Start scoutingUtilities structs
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
type PythonTeamWrapper struct {
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
type TeamResponse struct {
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
type EventResponse struct { //Comments go Description <TAB> Example
	Alliances           []FinalAlliance   `json:"alliances"`             //If we have alliance selection data for this event, this contains a JSON array of the alliances. The captain is the first team, followed by their picks, in order.
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
	Teams               []TeamResponse    `json:"teams"`                 //List of team models that attended the event
	Webcast             []json.RawMessage `json:"webcast"`               //If the event has webcast data associated with it, this contains JSON data of the streams
	EndDate             string            `json:"end_date"`              //Day the event ends in string format	"2014-03-29"
	StartDate           string            `json:"start_date"`            //Day the event starts in string format	"2014-03-27"
	//facebook_eid null
}
type FinalAlliance struct {
	Declines []string `json:"declines"`
	Picks    []string `json:"picks"`
}
type MatchResponse struct {
	Key         string         `json:"key"`          //TBA event key with the format yyyy[EVENT_CODE]_[COMP_LEVEL]m[MATCH_NUMBER], where yyyy is the year, and EVENT_CODE is the event code of the event, COMP_LEVEL is (qm, ef, qf, sf, f), and MATCH_NUMBER is the match number in the competition level. A set number may append the competition level if more than one match in required per set .	2010sc_qm10, 2011nc_qf1m2
	CompLevel   string         `json:"comp_level"`   //The competition level the match was played at.	qm, ef, qf, sf, f
	SetNumber   int            `json:"set_number"`   //The set number in a series of matches where more than one match is required in the match series.	2010sc_qf1m2, would be match 2 in quarter finals 1.
	MatchNumber int            `json:"match_number"` //The match number of the match in the competition level.	2010sc_qm20
	Alliances   MatchAlliances `json:"alliances"`    //A list of alliances, the teams on the alliances, and their score.
	EventKey    string         `json:"event_key"`    //Event key of the event the match was played at.	2011sc
	Videos      []VideoLink    `json:"videos"`       //JSON array of videos associated with this match and corresponding information	"videos": [{"key": "xswGjxzNEoY", "type": "youtube"}, {"key": "http://videos.thebluealliance.net/2010cmp/2010cmp_f1m1.mp4", "type": "tba"}]
	TimeString  string         `json:"time_string"`  //Time string for this match, as published on the official schedule. Of course, this may or may not be accurate, as events often run ahead or behind schedule	11:15 AM
	Time        string         `json:"time"`         //UNIX timestamp of match time, as taken from the published schedule	1394904600
}
type VideoLink struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}
type MatchAlliances struct {
	Red  Alliance `json:"red"`
	Blue Alliance `json:"blue"`
}
type Alliance struct {
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

//end scoutingUtilities Struct
