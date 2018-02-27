package main

import ("fmt"
		"net/http"
		"log"
		"encoding/json"
		"github.com/gorilla/mux"
		"strconv"
		"io/ioutil"
		"strings"
	)

	var Tasks []Task

	//First, a bit of architecture: Let's build the structure of a Task

type Task struct {

	//Has an ID
	ID int `json:"ID"`
	Name string `json:"Name"`
	Command string `json:"Command"`
	Output string `json:"Output"`
}

func main (){
	//Set listening port 
	port := "3000"
	
	//Create router which will serve as an HTTP request router for all calls
	router := mux.NewRouter()
	//Define endpoints
	router.HandleFunc("/api/tasks", GetTasks).Methods("GET")
	router.HandleFunc("/api/tasks/{id}",GetTaskByID).Methods("GET")
	router.HandleFunc("/api/tasks",AddNewTask).Methods("POST")
	router.HandleFunc("/api/tasks",ModifyTask).Methods("PATCH")
	//start HTTP endpoint with attached MUX(router)
	fmt.Printf("Starting to listen on port %s",port)
	log.Fatal(http.ListenAndServe(":" + port,router))
	

}

func GetTasks(w http.ResponseWriter, r *http.Request){
	
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
	json.NewEncoder(w).Encode(temp)
	return
}
func GetTaskByID(w http.ResponseWriter, r *http.Request){
	
	var temp []Task
	
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}	
	params := mux.Vars(r)

	for _, f := range files {
		isjson := strings.Contains(f.Name(), ".json")
		if isjson {
			content_bytes, _ := ioutil.ReadFile(f.Name())
			var task Task
			_ = json.Unmarshal(content_bytes,&task)
				tid, _ := strconv.Atoi(params["id"])
			if task.ID == tid {
			temp = append(temp,task)
		}

		}
	}
	json.NewEncoder(w).Encode(temp)
	return
}

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

//Add the new task with an auto-generated ID using the max+1 method
func AddNewTask(w http.ResponseWriter, r *http.Request){
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
		var taskids []int
	for _, t := range Tasks{
		taskids = append(taskids,t.ID) 
	}

	var _, max = MinMax(taskids)
	tid := max + 1
	task.ID = tid

	jsonstring , _ := json.Marshal(task)

	ioutil.WriteFile(string(tid) + ".json", []byte(jsonstring) , 0644)
	json.NewEncoder(w).Encode(task)
	return
}

func ModifyTask(w http.ResponseWriter, r *http.Request){
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
	fmt.Println(task)
	for i, item := range Tasks {
		if item.ID == task.ID {
			Tasks[i] = task
			json.NewEncoder(w).Encode(item)
			return 
		}
	}
}

