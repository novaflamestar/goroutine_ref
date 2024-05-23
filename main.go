package main

import (
	"fmt"
	"math/rand"
	"time"
)

type JobErrorInterface interface {
	Error() string
	GetId() int
}

type JobError struct {
	Id int
}

func (e JobError) Error() string {
	return fmt.Sprintf("Failed job %d", e.Id)
}

func (e JobError) GetId() int {
	return e.Id
}

func worker(id int) JobErrorInterface {
	fmt.Printf("working on job %d...\n", id)
	time.Sleep(time.Second * time.Duration(rand.Intn(5)))
	if id%2 == 0 {
		fmt.Printf("throwing error for job %d\n", id)
		return JobError{Id: id}
	}
	fmt.Printf("Finished job %d!\n", id)
	return nil
}

func validateJobs(completedJobs map[int]bool, numberOfJobs int) bool {
	for i := 0; i < numberOfJobs; i++ {
		if _, ok := completedJobs[i]; !ok {
			return false
		}
	}
	return true
}

func main() {
	// Initial value setup.
	completedJobs := make(map[int]bool)        // completed jobs holds a map of every completed job id associated with a true value if it ran
	numberOfJobs := 100                        // controls the number of jobs to run in total.
	doneJobs := make(chan int, 1)              // channel to pass finished jobs ids
	errChan := make(chan JobErrorInterface, 1) // channel to pass a custom job error if a job fails. This custom error exists to also pass the id
	maxConcurrent := make(chan bool, 5)        // The size of this buffer dictates how many jobs can be ran in parrallel at one time

	// Start all go routines. Each go routine pool is managed by the size of the maxConcurrent buffered channel
	for i := 0; i < numberOfJobs; i++ {
		go func(id int, errChan chan<- JobErrorInterface, idChan chan<- int, maxConcurrent chan bool) {
			maxConcurrent <- true
			defer func() { <-maxConcurrent }()
			err := worker(id)
			if err != nil {
				errChan <- err
				return
			}
			idChan <- id
		}(i, errChan, doneJobs, maxConcurrent)
	}

	// for the number of jobs run, check for an output or an error
	for i := 0; i < numberOfJobs; i++ {
		select {
		case id := <-doneJobs:
			fmt.Printf("finished job %d, incrementing coutner\n", id)
			completedJobs[id] = true
		case err := <-errChan:
			fmt.Printf("Error with a job: %v\n", err)
			completedJobs[err.GetId()] = true
		}
	}

	// validate that all jobs ran successfully
	if !validateJobs(completedJobs, numberOfJobs) {
		panic("some jobs did not finish")
	}
	fmt.Println("All jobs passed!")
}
