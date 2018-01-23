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
	slackChannelName string // TODO: add group version
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

// hold users
type user struct {
	id     string
	active bool
}

// construct new user with defaults
func newUser(id string) user {
	return user{
		id:     id,
		active: false,
	}
}

// hold channel info
type channel struct {
	id    string
	name  string
	users []user
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
	//conf.slackChannelName = os.Getenv("SLACK_CHANNELNAME")

	// get channel id from channel name
	conf.slackChannelID = os.Getenv("SLACK_CHANNELID")
	if conf.slackChannelID == "" {
		log.Println("SLACK_CHANNELID missing, using SLACK_CHANNELNAME to lookup ID instead")
		if conf.slackChannelName == "" {
			log.Fatal("SLACK_CHANNELNAME also not set")
		}
		// TODO: store this for the user maybe
		conf.slackChannelID = getSlackChannelIDFromName(conf.slackChannelName)
	}

	// define exercises (TODO: move to json)
	exercises := []exercise{
		exercise{"pushups", 10, 20, ""},
		exercise{"planking", 30, 60, "seconds"},
		exercise{"starjumps", 30, 45, ""},
	}

	fmt.Println("Start time:", conf.startTime)
	fmt.Println("End time:", conf.endTime)

	// list exercises
	//fmt.Print(strings.Join(returnExercises(exercises), "\n\t"), "\n")
	//sendSlackMessage(strings.Join(returnExercises(exercises), "\n\t") + "\n")

	var channel channel
	channel.id = conf.slackChannelID

	// pick a person and an exercise
	//for i := 0; i < 5; i++ {

	// save users
	// TODO: don't wipe this out between runs (in case we want to store more info such as name)
	channel.users = getSlackChannelMembers(channel.id)

	// update users status
	updateSlackActiveUsers(&channel.users)

	fmt.Print(strings.Join(returnUsers(channel.users, false), "\n\t"), "\n")
	fmt.Print(strings.Join(returnUsers(channel.users, true), "\n\t"), "\n")

	// choose a user
	chosenUser := channel.getRandomActiveUser()

	// if we were able to choose a random user
	if (user{}) != chosenUser {
		chosenExercise := chooseRandomExercise(exercises)
		chosenExerciseUnit := chosenExercise.unit
		if chosenExerciseUnit != "" {
			chosenExerciseUnit += " of "
		}

		// build message and send
		message := fmt.Sprintf("It's time for <@%s> to do %d %s%s\n", chosenUser.id, chooseRandomExerciseReps(chosenExercise), chosenExerciseUnit, chosenExercise.name)
		fmt.Print(message)
		//sendSlackMessage(message)

	} else {
		fmt.Print("Nobody is active!")
	}
	//time.Sleep(time.Second * 1)
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

// return users (for printing to console/channel)
func returnUsers(users []user, onlyActiveUsers bool) []string {
	msg := "All users:"
	if onlyActiveUsers {
		msg = "Active users:"
	}
	returnSlice := []string{msg}
	for _, u := range users {
		if !onlyActiveUsers || (onlyActiveUsers && u.active) {
			returnSlice = append(returnSlice, u.id)
		}
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

// choose random user from channel struct
func (c channel) getRandomActiveUser() user {
	activeUsers := c.getActiveUsers()
	if len(activeUsers) > 0 {
		return activeUsers[rand.Intn(len(activeUsers))]
	}
	return user{}
}

// send slack message
func sendSlackMessage(message string) {
	callSlackAPI("POST", conf.slackWebHookURL, false, fmt.Sprintf("{\"text\":\"%s\"}", message))
}

// get slack channel members
// TODO: handle 0 members
// TODO: handle channel/group (different api calls)
// TODO: handle wrong channel ID
func getSlackChannelMembers(slackChannelID string) []user {
	var members []user

	htmlData := callSlackAPI("GET", "https://slack.com/api/groups.info?channel="+slackChannelID, true, "")

	type channelInfo struct {
		Group struct {
			Members []string
		}
	}
	var c channelInfo
	json.Unmarshal([]byte(htmlData), &c)

	for _, u := range c.Group.Members {
		members = append(members, newUser(u))
	}

	return members
}

// update slack active users
func updateSlackActiveUsers(users *[]user) {
	// if there are channel members
	if len(*users) > 0 {
		type getPresence struct {
			Presence string
		}

		var p getPresence

		// loop through channel members to check their presence
		for i := range *users {
			htmlData := callSlackAPI("GET", "https://slack.com/api/users.getPresence?user="+(*users)[i].id, true, "")
			json.Unmarshal([]byte(htmlData), &p)

			if p.Presence == "active" {
				(*users)[i].active = true
			} else {
				(*users)[i].active = false
			}
		}
	}
}

// get active users from channel struct
func (c channel) getActiveUsers() (activeUsers []user) {
	for _, u := range c.users {
		if u.active {
			activeUsers = append(activeUsers, u)
		}
	}

	return activeUsers
}

// get slack channel id from name
func getSlackChannelIDFromName(slackChannelName string) (slackChannelID string) {
	htmlData := callSlackAPI("GET", "https://slack.com/api/groups.list", true, "")

	type channelList struct {
		Ok     string
		Groups []struct {
			ID   string
			Name string
		}
	}
	var cl channelList

	json.Unmarshal([]byte(htmlData), &cl)

	for _, c := range cl.Groups {
		if c.Name == slackChannelName {
			slackChannelID = c.ID
		}
	}

	return slackChannelID
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
