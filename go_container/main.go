package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	HTTPPortNumber string `json:"HTTP_Port_Number"`
	DBHostName     string `json:"DB_HostName"`
	Method         string
}

//GetConfig : Getting global config from ENV variables, or from using Config file
func GetConfig() {

	env := true
	tempcfgdbhostname := os.Getenv("APP_DB_HOSTNAME")
	fmt.Println(tempcfgdbhostname)
	tempcfgport := os.Getenv("APP_PORT")
	fmt.Println(tempcfgport)
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
		if err == nil {
			globalconfig.Method = "CofigFile"
		}
	} else {
		globalconfig.Method = "Enviroment"
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
	fmt.Println("Getting app configuration")
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
	router.HandleFunc("/api/tasks/{id}", GetTaskByIDHandler).Methods("GET")
	router.HandleFunc("/api/tasks", AddNewTaskHandler).Methods("POST")
	router.HandleFunc("/api/tasks", ModifyTaskHandler).Methods("PATCH")
	//start HTTP endpoint with attached MUX(router)
	fmt.Printf("Starting to listen on port %s", globalconfig.HTTPPortNumber)
	log.Fatal(http.ListenAndServe(":"+globalconfig.HTTPPortNumber, router))

}

//ReadAllTasks : Internal fuction to Read the "DB" (Create a slice of Tasks based on Json file)
func ReadAllTasks() []Task {
	var temp []Task
	err := globalDBcollection.Find(nil).All(&temp)
	if err != nil {
		panic(err)
	}
	return temp
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

//SearchTaskbyID : Get task by ID field and returns it in JSON format
func SearchTaskbyID(id int) Task {

	found := Task{}
	found.ID = -1

	err := globalDBcollection.Find(bson.M{"id": id}).One(&found)
	if err != nil {
		log.Fatal(err)
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
	//Write the new Task to the "DB"
	//WriteTasktoJSONFile(task)
	WriteTaskToDB(task)
	// Run the command defined in the task
	RunTaskAsync(task)
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

//RunTaskAsync : Run task in a separate goroutine
func RunTaskAsync(t Task) {
	go RunTask(t)
}

//RunTask : It runs the task based on Task definition, and update status + output field in DB
func RunTask(t Task) {
	time := time.Now()
	t.LastRunDateTime = time.String()
	t.Status = "Running"
	ModifyTask(t)
	o, err := exec.Command("cmd", "/c", t.Command).Output()
	if err != nil {
		t.Status = "Failed"
		t.Output = string("Error while executing command, please, check Your syntax. Error description: " + err.Error())
	} else {
		t.Status = "Success"
		t.Output = string(o)
	}

	ModifyTask(t)
}
