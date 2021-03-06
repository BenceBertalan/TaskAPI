package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

//Define global variables to be used throughout the App
var globalconfig Config
var globalDBcollection *mgo.Collection

//Task : The structure of a Task
type Task struct {
	ID              int    `json:"ID"`
	Name            string `json:"Name"`
	Command         string `json:"Command"`
	Status          string `json:"Status"`
	Output          string `json:"Output"`
	CreatedDateTime string `json:"Created_DateTime"`
	LastRunDateTime string `json:"Last_Run_DateTime"`
}

//Config : Structure to contain Config data
type Config struct {
	HTTPPortNumber string `json:"APP_PORT"`
	DBHostName     string `json:"APP_DB_HOSTNAME"`
	Method         string
}

//GetConfig : Getting global config from ENV variables, or from using Config file
func GetConfig() {

	env := true
	tempcfgdbhostname := os.Getenv("APP_DB_HOSTNAME")
	tempcfgport := os.Getenv("APP_PORT")
	if tempcfgdbhostname != "" {
		globalconfig.DBHostName = tempcfgdbhostname
	} else {
		env = false
	}
	if tempcfgport != "" {
		globalconfig.HTTPPortNumber = tempcfgport
	} else {
		env = false
	}

	if env == false {
		configbytes, err := ioutil.ReadFile("config.json")
		if err != nil {
		}
		err = json.Unmarshal(configbytes, &globalconfig)
		fmt.Println(globalconfig)
		if err == nil {
			globalconfig.Method = "CofigFile"
		}
	} else {
		globalconfig.Method = "EnvironmentVars"
	}
}

//CreateDBConn : Universal function to create a DB connection and return a pointer to the collection containing the Task data
func CreateDBConn(checkconn bool) {
	if checkconn {
		fmt.Println("Connecting to DBHost: " + globalconfig.DBHostName)
	}
	session, err := mgo.Dial(globalconfig.DBHostName)
	if err == nil {
		if checkconn {
			fmt.Println("DB is available!")
		}
	} else {
		panic("DB can not be reached")
	}

	globalDBcollection = session.DB("TaskAPI").C("Tasks")

}

func main() {
	//Load Global Config
	fmt.Println("Getting app configuration...")
	GetConfig()
	if globalconfig.Method == "" {
		panic("Unable to get config using any method")
	} else {
		fmt.Println("AppConfig successfully loaded using " + globalconfig.Method)
	}

	fmt.Println("Checking MongoDB Connection...")

	//create DB connection and set global DB collection to work with
	CreateDBConn(true)

	//Create router which will serve as an HTTP request router for all calls
	router := mux.NewRouter()
	//Define endpoints
	router.HandleFunc("/api/tasks", GetTasksHandler).Methods("GET")
	router.HandleFunc("/api/tasks/pending", GetPendingTasksHandler).Methods("GET")
	router.HandleFunc("/api/tasks/{id}", GetTaskByIDHandler).Methods("GET")
	router.HandleFunc("/api/tasks", AddNewTaskHandler).Methods("POST")
	router.HandleFunc("/api/tasks", ModifyTaskHandler).Methods("PATCH")
	//start HTTP endpoint with attached MUX(router)
	fmt.Println("Starting to listen on port " + globalconfig.HTTPPortNumber)
	log.Fatal(http.ListenAndServe(":"+globalconfig.HTTPPortNumber, router))

}

//ReadAllTasks : Internal fuction to Read the "DB" (Create a slice of Tasks based on Json file)
func ReadAllTasks() []Task {
	temp := make([]Task, 0)
	err := globalDBcollection.Find(nil).All(&temp)
	if err != nil {
	}
	return temp
}

//GetAllPendingTasks : Queries pending tasksk from the db
func GetAllPendingTasks() []Task {
	pending := make([]Task, 0)
	for _, t := range ReadAllTasks() {
		if t.Status == "Pending" {
			pending = append(pending, t)
		}
	}
	return pending
}

//WriteTaskToDB : Inserts the Task to the DB
func WriteTaskToDB(t Task) {
	err := globalDBcollection.Insert(&t)
	if err != nil {
		panic(err)
	}
}

//GetTasksHandler : Get all tasks in JSON format
func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ReadAllTasks())
	return
}

//GetPendingTasksHandler : Writes Pending task response
func GetPendingTasksHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(GetAllPendingTasks())
	return
}

//SearchTaskbyID : Get task by ID field and returns it in JSON format
func SearchTaskbyID(id int) Task {

	found := Task{}
	found.ID = -1

	err := globalDBcollection.Find(bson.M{"id": id}).One(&found)
	if err != nil {

	}

	return found
}

//GetTaskByIDHandler : Get a task using URL parameter
func GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tid, _ := strconv.Atoi(params["id"])
	json.NewEncoder(w).Encode(SearchTaskbyID(tid))
	return
}

//MinMax : Helper funcion to get max ID from an array
func MinMax(array []int) (int, int) {
	max := array[0]
	min := array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

//AddNewTaskHandler Add the new task with an auto-generated ID using the max+1 method (or 0 if no Tasks are specified)
func AddNewTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
	var taskids []int
	tasks := ReadAllTasks()
	if tasks != nil {
		for _, t := range tasks {
			taskids = append(taskids, t.ID)

		}
		var _, max = MinMax(taskids)
		tid := max + 1
		task.ID = tid
	} else {
		task.ID = 0
	}
	time := time.Now()
	task.CreatedDateTime = time.String()
	task.Status = "Pending"
	//Write the new Task to the "DB"
	WriteTaskToDB(task)
	//Return the task in the HTTP response output
	json.NewEncoder(w).Encode(task)
	return

}

//ModifyTask : Queries task by TaskID from DB and modifies it (important for status update)
func ModifyTask(t Task) Task {
	query := bson.M{"id": t.ID}
	change := bson.M{"$set": t}
	err := globalDBcollection.Update(query, change)
	if err != nil {
		panic(err)
	}

	/*
		for _, item := range ReadAllTasks() {
			if item.ID == t.ID {
				WriteTasktoJSONFile(t)
			}
		}*/
	return t
}

//ModifyTaskHandler : External Handler for the Task modification using post
func ModifyTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
	json.NewEncoder(w).Encode(ModifyTask(task))
	return
}
