package main

import (
	"fmt"
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
	people := getPeople()

	// TODO: add error handling for missing envs
	conf.slackWebHookURL = os.Getenv("SLACK_WEBHOOKURL")
	conf.slackAuthToken = os.Getenv("SLACK_AUTHTOKEN")
	conf.slackChannelName = os.Getenv("SLACK_CHANNELNAME")

	// get channel id from channel name
	conf.slackChannelID = os.Getenv("SLACK_CHANNELID")
	if conf.slackChannelID == "" {
		// TODO: go get channel id from channel name
		fmt.Println("Don't have a channel id")
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
	fmt.Print(strings.Join(returnPeople(people), "\n\t"), "\n")

	// list exercises
	sendSlackMessage(strings.Join(returnExercises(exercises), "\n\t") + "\n")

	// pick a person and an exercise
	//for i := 0; i < 10; i++ {
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

// get people (TODO: replace with dynamic channel members call)
func getPeople() []person {
	people := []person{
		person{"nick"},
		person{"dave"},
		person{"bob"},
		person{"fred"},
	}

	return people
}

// choose random person
func chooseRandomPerson(people []person) person {
	return people[rand.Intn(len(people))]
}

// send slack message
func sendSlackMessage(message string) {
	body := strings.NewReader(fmt.Sprintf("{\"text\":\"%s\"}", message))
	req, err := http.NewRequest("POST", conf.slackWebHookURL, body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
}
