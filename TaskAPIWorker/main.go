package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var apiendpoint string

//Task : Default task struct
type Task struct {
	ID              int    `json:"ID"`
	Name            string `json:"Name"`
	Command         string `json:"Command"`
	Status          string `json:"Status"`
	Output          string `json:"Output"`
	CreatedDateTime string `json:"Created_DateTime"`
	LastRunDateTime string `json:"Last_Run_DateTime"`
}

func main() {
	apiptr := flag.String("apiendpoint", "", "URL of the API endpoint")
	flag.Parse()

	apihost := os.Getenv("APP_API_HOSTNAME")
	if apihost != "" {
		apiendpoint = "http://" + apihost + ":3000/api/tasks"
	} else {
		apiendpoint = *apiptr
	}
	if apiendpoint == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Ready to serve! Using API: " + apiendpoint)
	for true {
		MainProcess()
	}
}

//MainProcess : The main process of the app
func MainProcess() {
	for _, t := range GetAllPendingTasks() {
		ProcessTaskAsync(t)
	}
	time.Sleep(2 * time.Second)
}

//ModifyTask : Modifies the task using the API PATCH HTTP verb
func ModifyTask(t Task) {

	url := apiendpoint
	jsonStr, _ := json.Marshal(t)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
	}

	if resp.StatusCode == 200 {
	} else {
		log.Fatal("Error while modifying task: " + err.Error())
	}
}

//Call process task with async func
func ProcessTaskAsync(t Task) {
	idstr := strconv.Itoa(t.ID)
	fmt.Println("Running task with id: " + idstr)
	go ProcessTask(t)
}

//ProcessTask : Processes the task
func ProcessTask(t Task) {
	time := time.Now()
	t.LastRunDateTime = time.String()
	t.Status = "Running"
	ModifyTask(t)
	o, err := exec.Command(t.Command).Output()
	if err != nil {
		t.Status = "Failed"
		t.Output = string("Error while executing command, please, check Your syntax. Error description: " + err.Error())
	} else {
		t.Status = "Success"
		t.Output = string(o)
	}

	ModifyTask(t)
}

//GetAllPendingTasks : Queries all pending tasksk from the DB
func GetAllPendingTasks() []Task {
	var alltasks []Task
	url := apiendpoint + "/pending"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
	}

	if err := json.NewDecoder(resp.Body).Decode(&alltasks); err != nil {
		fmt.Println(err)
	}

	if len(alltasks) == 0 {
		//fmt.Println("No pending tasksk are available")
	}

	return alltasks

}
