package main

import ("fmt"
		"net/http"
		"log"
		"encoding/json"
		"github.com/gorilla/mux"
		"strconv"
		"io/ioutil"
		"strings"
		"os/exec"
		"time"
	)

//First, a bit of architecture: Let's build the structure of a Task

type Task struct {

	//Has an ID
	ID int `json:"ID"`
	Name string `json:"Name"`
	Command string `json:"Command"`
	Status string `json:"Status"`
	Output string `json:"Output"`
	CreatedDateTime string `json:"CreatedDateTime"`
	LastRunDateTime string `json:"LastRunDateTime"`
}


func main (){
	//Set listening port 
	port := "3000"
	
	//Create router which will serve as an HTTP request router for all calls
	router := mux.NewRouter()
	//Define endpoints
	router.HandleFunc("/api/tasks", GetTasksHandler).Methods("GET")
	router.HandleFunc("/api/tasks/{id}",GetTaskByIDHandler).Methods("GET")
	router.HandleFunc("/api/tasks",AddNewTaskHandler).Methods("POST")
	router.HandleFunc("/api/tasks",ModifyTaskHandler).Methods("PATCH")
	//start HTTP endpoint with attached MUX(router)
	fmt.Printf("Starting to listen on port %s",port)
	log.Fatal(http.ListenAndServe(":" + port,router))
	

}

//Internal fuction to Read the "DB" (Create a slice of Tasks based on Json file)
func ReadAllTasks() []Task {
	var temp []Task
	
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		isjson := strings.Contains(f.Name(), ".json")
		if isjson {
			content_bytes, _ := ioutil.ReadFile(f.Name())
			var task Task
			_ = json.Unmarshal(content_bytes,&task)
			temp = append(temp,task)

		}
	}

	return temp
}

func WriteTasktoJSONFile(t Task){
	
	jsonbytes , _ := json.Marshal(t)
	err := ioutil.WriteFile(strconv.Itoa(t.ID) + ".json", jsonbytes , 0644)
	if(err != nil){
		log.Fatal(err)
	}	

}

//Get all tasks in JSON format

func GetTasksHandler(w http.ResponseWriter, r *http.Request){
	json.NewEncoder(w).Encode(ReadAllTasks())
	return
}

func SearchTaskbyID(id int) Task {
	temp := ReadAllTasks()
	var found Task
	for _, t := range temp {
		if t.ID == id {
			found = t
		}
	}
	return found
}

//Get a task using URL parameter
func GetTaskByIDHandler(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	tid, _ := strconv.Atoi(params["id"])
	json.NewEncoder(w).Encode(SearchTaskbyID(tid))
	return
}

// Helper funcion to get max ID from an array
func MinMax(array []int) (int, int) {
    var max int = array[0]
    var min int = array[0]
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

//Add the new task with an auto-generated ID using the max+1 method (or 0 if no Tasks are specified)
func AddNewTaskHandler(w http.ResponseWriter, r *http.Request){
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
		var taskids []int
	tasks := ReadAllTasks()
	if(tasks != nil){
	for _, t := range tasks{
		taskids = append(taskids,t.ID) 

	}
	var _, max = MinMax(taskids)
	tid := max + 1
	task.ID = tid
}else{
	task.ID = 0
}
	time := time.Now()
	task.CreatedDateTime = time.String()
//Write the new Task to the "DB"
	WriteTasktoJSONFile(task)
// Run the command defined in the task
	RunTaskAsync(task)
//Return the task in the HTTP response output
	json.NewEncoder(w).Encode(task)
	return



}

//Modify task (important for status update)
func ModifyTask(t Task) Task{
	for _, item := range ReadAllTasks() {
		if item.ID == t.ID {
			WriteTasktoJSONFile(t)
		}
	}
	return t
}

//External Handler for the Task modification using post
func ModifyTaskHandler(w http.ResponseWriter, r *http.Request){
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
	json.NewEncoder(w).Encode(ModifyTask(task))
	return 
}

// Run task in a goroutine
func RunTaskAsync(t Task){
	go RunTask(t)
}

//Run Task based on Task definition, and update status + output field in DB
func RunTask(t Task){
	time := time.Now()
	t.LastRunDateTime = time.String()
	t.Status = "Running"
	ModifyTask(t)	
	o, err := exec.Command("cmd", "/c", t.Command).Output()
	if(err != nil){
		t.Status = "Failed"
		t.Output = string("Error while executing command, please, check Your syntax")	
	}else{
	t.Status = "Success"
	t.Output = string(o)}
	
	ModifyTask(t)
}

