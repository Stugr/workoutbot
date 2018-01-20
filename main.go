package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// hold config
type config struct {
	startTime time.Time
	endTime   time.Time
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

func main() {
	// TODO: change seed to use something else later?
	rand.Seed(time.Now().UTC().UnixNano())

	// define and initialise config
	conf := config{}
	setDefaultConfig(&conf)
	people := getPeople()

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

	// pick a person and an exercise
	//for i := 0; i < 10; i++ {
	chosenPerson := chooseRandomPerson(people)
	chosenExercise := chooseRandomExercise(exercises)
	chosenExerciseUnit := chosenExercise.unit
	if chosenExerciseUnit != "" {
		chosenExerciseUnit += " of "
	}

	fmt.Printf("It's time for @%s to do %d %s%s\n", chosenPerson.name, chooseRandomExerciseReps(chosenExercise), chosenExerciseUnit, chosenExercise.name)
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
		person{"max"},
		person{"john"},
	}

	return people
}

// choose random person
func chooseRandomPerson(people []person) person {
	return people[rand.Intn(len(people))]
}
