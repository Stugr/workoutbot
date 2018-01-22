package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// hold config
type config struct {
	startTime        time.Time
	endTime          time.Time
	slackChannelName string
	slackChannelID   string
	slackAuthToken   string
	slackWebHookURL  string
}

// hold exercises
type exercise struct {
	name string
	min  int
	max  int
	unit string
}

// hold people
type person struct {
	name string
}

var conf = config{}

func main() {
	// TODO: change seed to use something else later?
	rand.Seed(time.Now().UTC().UnixNano())

	// load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// define and initialise config (TODO: global variable, remove pointer?)
	setDefaultConfig(&conf)

	// TODO: add error handling for missing envs
	conf.slackWebHookURL = os.Getenv("SLACK_WEBHOOKURL")
	conf.slackAuthToken = os.Getenv("SLACK_AUTHTOKEN")
	conf.slackChannelName = os.Getenv("SLACK_CHANNELNAME")

	// get channel id from channel name
	conf.slackChannelID = os.Getenv("SLACK_CHANNELID")
	if conf.slackChannelID == "" {
		// TODO: go get channel id from channel name
		fmt.Println("Don't have a channel id")
		//https://slack.com/api/channels.list
	}

	// define exercises (TODO: move to json)
	exercises := []exercise{
		exercise{"pushups", 10, 20, ""},
		exercise{"planking", 30, 60, "seconds"},
		exercise{"starjumps", 30, 45, ""},
	}

	fmt.Println("Start time:", conf.startTime)
	fmt.Println("End time:", conf.endTime)

	// print possibilities
	fmt.Print(strings.Join(returnExercises(exercises), "\n\t"), "\n")
	//fmt.Print(strings.Join(returnPeople(people), "\n\t"), "\n")

	// list exercises
	//sendSlackMessage(strings.Join(returnExercises(exercises), "\n\t") + "\n")

	// pick a person and an exercise
	//for i := 0; i < 10; i++ {
	//people := loadPeople() // load test people manually

	people := getSlackChannelActiveMembers()

	// if there is at least 1 active person
	if len(people) > 0 {
		chosenPerson := chooseRandomPerson(people)
		chosenExercise := chooseRandomExercise(exercises)
		chosenExerciseUnit := chosenExercise.unit
		if chosenExerciseUnit != "" {
			chosenExerciseUnit += " of "
		}

		// build message and send
		message := fmt.Sprintf("It's time for <@%s> to do %d %s%s\n", chosenPerson.name, chooseRandomExerciseReps(chosenExercise), chosenExerciseUnit, chosenExercise.name)
		fmt.Print(message)
		sendSlackMessage(message)
		//time.Sleep(time.Second * 1)
	} else {
		fmt.Print("Nobody is active!")
	}
	//}
}

// set config defaults
func setDefaultConfig(c *config) {
	c.startTime = time.Now()
	c.endTime = (c.startTime).Add(time.Minute * time.Duration(60))
}

// return exercises (for printing to console/channel)
func returnExercises(exercises []exercise) []string {
	returnSlice := []string{"Possible exercises:"}
	for _, e := range exercises {
		returnSlice = append(returnSlice, fmt.Sprintf("%s (%d-%d)", e.name, e.min, e.max))
	}

	return returnSlice
}

// return people (for printing to console/channel)
func returnPeople(people []person) []string {
	returnSlice := []string{"Possible people:"}
	for _, p := range people {
		returnSlice = append(returnSlice, p.name)
	}

	return returnSlice
}

// choose random exercise
func chooseRandomExercise(exercises []exercise) exercise {
	return exercises[rand.Intn(len(exercises))]
}

// choose random exercise reps from range
func chooseRandomExerciseReps(exercise exercise) int {
	return rand.Intn(exercise.max-exercise.min) + exercise.min
}

// load people (TODO: remove)
func loadPeople() []person {
	people := []person{
		person{"U09BCKHT4"},
		person{"jack"},
		person{"bob"},
		person{"fred"},
	}

	return people
}

// choose random person
func chooseRandomPerson(people []person) person {
	if len(people) > 0 {
		return people[rand.Intn(len(people))]
	}
	return person{}
}

// send slack message
func sendSlackMessage(message string) {
	callSlackAPI("POST", conf.slackWebHookURL, false, fmt.Sprintf("{\"text\":\"%s\"}", message))
}

// get slack channel members
// TODO: handle 0 members
func getSlackChannelMembers() []string {
	htmlData := callSlackAPI("GET", "https://slack.com/api/groups.info?channel="+conf.slackChannelID, true, "")

	type ChannelInfo struct {
		Group struct {
			Members []string
		}
	}
	var c ChannelInfo
	json.Unmarshal([]byte(htmlData), &c)

	return c.Group.Members
}

// get slack channel active members
func getSlackChannelActiveMembers() []person {
	// get channel members
	members := getSlackChannelMembers()
	var activeMembers []person

	// if there are channel members
	if len(members) > 0 {
		type GetPresence struct {
			Presence string
		}

		var p GetPresence

		// loop through channel members to check their presence
		for _, m := range members {
			htmlData := callSlackAPI("GET", "https://slack.com/api/users.getPresence?user="+m, true, "")
			json.Unmarshal([]byte(htmlData), &p)

			if p.Presence == "active" {
				activeMembers = append(activeMembers, person{m})
			}
		}
	}

	return activeMembers
}

// call slack api
func callSlackAPI(method string, url string, auth bool, text string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	body := strings.NewReader(text) // TODO: this needed when blank? can body just be nil?
	req, err := http.NewRequest(method, url, body)
	req = req.WithContext(ctx)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if auth {
		req.Header.Set("Authorization", "Bearer "+conf.slackAuthToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	htmlData, err := ioutil.ReadAll(resp.Body)

	return htmlData
}
