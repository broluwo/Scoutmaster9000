//Package structs holds the the structs which the various parts of Scoutmaster
//9000 will be using. It's done this way for the sake of centralization as well
//as peace of mind.
package structs

import (
	"encoding/json"
	"net/http"
)

//Team is the struct that represents a team
type Team struct {
	Force  bool   `json:"force,omitempty" bson:"force,omitempty"`
	Number int    `json:"team_number,omitempty" bson:"number,omitempty"`
	Name   string `json:"nickname,omitempty" bson:"name,omitempty"`
	//Will need to be changed to it's own struct
	Reviews []string `json:"reviews,omitempty" bson:"reviews,omitempty"`
	//Will need to be changed to it's own struct
	Matches []int `json:"matches,omitempty" bson:"matches,omitempty"`
	//Will/Should be corrected soon
	Photos []byte `json:"photos,omitempty" bson:"photos,omitempty"`
}

//TeamResponse is the Current Representation of what TBA sends back.
//Wholly unnecessary but allows me to see what I'm working with
type TeamResponse struct {
	Name     string   `json:"name,omitempty" bson:"name,omitempty"`
	Locality string   `json:"locality,omitempty" bson:"locality,omitempty"`
	Number   int      `json:"team_number,omitempty" bson:"number,omitempty"`
	Region   string   `json:"region,omitempty" bson:",omitempty"`
	Key      string   `json:"key,omitempty" bson:"key,omitempty"`
	Country  string   `json:"country_name,omitempty" bson:"country,omitempty"`
	Website  string   `json:"website,omitempty" bson:"website,omitempty"`
	Nickname string   `json:"nickname,omitempty" bson:"nickname,omitempty"`
	Events   []string `json:"events,omitempty" bson:"events,omitempty"`
}

//EventResponse is the current Representation of what TBA sends back.
type EventResponse struct { //Comments go Description <TAB> Example
	Alliances           []FinalAlliance   `json:"alliances,omitempty" bson:"alliances,omitempty"`                         //If we have alliance selection data for this event, this contains a JSON array of the alliances. The captain is the first team, followed by their picks, in order.
	Key                 string            `json:"key,omitempty" bson:"key,omitempty"`                                     //TBA event key with the format yyyy[EVENT_CODE], where yyyy is the year, and EVENT_CODE is the event code of the event.	2010sc
	Name                string            `json:"name,omitempty" bson:"name,omitempty"`                                   //Official name of event on record either provided by FIRST or organizers of offseason event.	Palmetto Regional
	ShortName           string            `json:"short_name,omitempty" bson:"short_name,omitempty"`                       //name but doesn't include event specifiers, such as 'Regional' or 'District'.	Palmetto
	EventCode           string            `json:"event_code,omitempty" bson:"event_code,omitempty"`                       //Event short code.	SC
	EventTypeString     string            `json:"event_type_string,omitempty" bson:"event_type_string,omitempty"`         //A human readable string that defines the event type.	'Regional', 'District', 'District Championships', 'District Championship','Championship Division', 'Championship Finals', 'Offseason','Preseason', '--'
	EventType           int               `json:"event_type,omitempty" bson:"event_type,omitempty"`                       //An integer that represents the event type as a constant.	List of constants to event type
	EventDistrictString string            `json:"event_district_string,omitempty" bson:"event_district_string,omitempty"` //A human readable string that defines the event's district.	'Michigan', 'Mid Atlantic', null (if regional)
	EventDistrict       int               `json:"event_district,omitempty" bson:"event_district,omitempty"`               //An integer that represents the event district as a constant.	List of constants to event district
	Year                int               `json:"year,omitempty" bson:"year,omitempty"`                                   //Year the event data is for.	2010
	Location            string            `json:"location,omitempty" bson:"location,omitempty"`                           //Long form address that includes city, and state provided by FIRST	Clemson, SC
	VenueAddress        string            `json:"venue_address,omitempty" bson:"venue_address,omitempty"`                 //Address of the event's venue, if available. Line breaks included.	Long Beach Arena\n300 East Ocean Blvd\nLong Beach, CA 90802\nUSA
	Website             string            `json:"website,omitempty" bson:"website,omitempty"`                             //The event's website, if any.	http://www.firstsv.org
	Official            bool              `json:"official,omitempty" bson:"official,omitempty"`                           //Whether this is a FIRST official event, or an offseaon event.	true
	Teams               []TeamResponse    `json:"teams,omitempty" bson:"teams,omitempty"`                                 //List of team models that attended the event
	Webcast             []json.RawMessage `json:"webcast,omitempty" bson:"webcast,omitempty"`                             //If the event has webcast data associated with it, this contains JSON data of the streams
	EndDate             string            `json:"end_date,omitempty" bson:"end_date,omitempty"`                           //Day the event ends in string format	"2014-03-29"
	StartDate           string            `json:"start_date,omitempty" bson:"start_date,omitempty"`                       //Day the event starts in string format	"2014-03-27"
	//facebook_eid null
}

//FinalAlliance is the represntation of the final alliance selection process
type FinalAlliance struct {
	Declines []string `json:"declines,omitempty" bson:"declines,omitempty"`
	Picks    []string `json:"picks,omitempty" bson:"picks,omitempty"`
}

//MatchResponse is the representation of what is sent by TBA
type MatchResponse struct {
	Key         string         `json:"key,omitempty" bson:"key,omitempty"`                   //TBA event key with the format yyyy[EVENT_CODE]_[COMP_LEVEL]m[MATCH_NUMBER], where yyyy is the year, and EVENT_CODE is the event code of the event, COMP_LEVEL is (qm, ef, qf, sf, f), and MATCH_NUMBER is the match number in the competition level. A set number may append the competition level if more than one match in required per set .	2010sc_qm10, 2011nc_qf1m2
	CompLevel   string         `json:"comp_level,omitempty" bson:"comp_level,omitempty"`     //The competition level the match was played at.	qm, ef, qf, sf, f
	SetNumber   int            `json:"set_number,omitempty" bson:"set_number,omitempty"`     //The set number in a series of matches where more than one match is required in the match series.	2010sc_qf1m2, would be match 2 in quarter finals 1.
	MatchNumber int            `json:"match_number,omitempty" bson:"match_number,omitempty"` //The match number of the match in the competition level.	2010sc_qm20
	Alliances   MatchAlliances `json:"alliances,omitempty" bson:"alliances,omitempty"`       //A list of alliances, the teams on the alliances, and their score.
	EventKey    string         `json:"event_key,omitempty" bson:"event_key,omitempty"`       //Event key of the event the match was played at.	2011sc
	Videos      []VideoLink    `json:"videos,omitempty" bson:"videos,omitempty"`             //JSON array of videos associated with this match and corresponding information	"videos": [{"key": "xswGjxzNEoY", "type": "youtube"}, {"key": "http://videos.thebluealliance.net/2010cmp/2010cmp_f1m1.mp4", "type": "tba"}]
	TimeString  string         `json:"time_string,omitempty" bson:"time_string,omitempty"`   //Time string for this match, as published on the official schedule. Of course, this may or may not be accurate, as events often run ahead or behind schedule	11:15 AM
	Time        string         `json:"time,omitempty" bson:"time,omitempty"`                 //UNIX timestamp of match time, as taken from the published schedule	1394904600
}

//VideoLink is the struct that holds the data needed to link to a YT video.
type VideoLink struct {
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Key  string `json:"key,omitempty" bson:"key,omitempty"`
}

//MatchAlliances are the two types of Alliances per match
type MatchAlliances struct {
	Red  Alliance `json:"red,omitempty" bson:"red,omitempty"`
	Blue Alliance `json:"blue,omitempty" bson:"blue,omitempty"`
}

//Alliance is a representation of the subset of teams in a match
type Alliance struct {
	Score int      `json:"score,omitempty" bson:"score,omitempty"`
	Teams []string `json:"teams,omitempty" bson:"teams,omitempty"`
}

//The Match struct is how a match is represented
type Match struct {
	Number    int    `json:"number,omitempty" bson:"number,omitempty"`
	Type      string `json:"type,omitempty" bson:"type,omitempty"`
	Red       []int  `json:"red,omitempty" bson:"red,omitempty"` //These should be strings in next v
	Blue      []int  `json:"blue,omitempty" bson:"blue,omitempty"`
	RedScore  int    `json:"rScore,omitempty" bson:"rScore,omitempty"`
	BlueScore int    `json:"bScore,omitempty" bson:"bScore,omitempty"`
	Winner    string `json:"winner,omitempty" bson:"winner,omitempty"`
}

//Regional How the python server takes regional
type Regional struct {
	Location    string            `json:"location,omitempty" bson:"location,omitempty"`
	Matches     []Match           `json:"matches,omitempty" bson:"matches,omitempty"`
	WinnerArray map[string][3]int `json:"winnerCount,omitempty" bson:"winnerCount,omitempty"`
	Year        int               `json:"year,omitempty" bson:"year,omitempty"`
}

//end marshalled structs

//Route is the struct that defines the properties we use for the routes we need
//handled by the new mux router.
//The first param of Route must have the trailing slash left off
type Route struct {
	PrefixRoute    string
	PostfixRoute   string
	PrefixHandler  func(http.ResponseWriter, *http.Request)
	PostfixHandler func(http.ResponseWriter, *http.Request)
}

//Routes ...
type Routes []Route

//end struct files
