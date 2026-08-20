package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx  = context.Background()
var rdb  *redis.Client 


func initRedis(){
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	_,err := rdb.Ping(ctx).Result()

	if err != nil {
		fmt.Println("Failed to connect to Redis:", err)
		return;
	}
	fmt.Println("Connected to Redis", addr)
}

func hello ( w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w,"Hello from mini judge")

}

type ExecuteRequest struct {
	Language string `json:"language"`
	Code string `json:"code"`
}

type ExecuteResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}


type Job struct {
	ID string `json:"id"`
	Language string `json:"language"`
	Code string `json:"code"`
}
func executeHandler(w http.ResponseWriter,r  *http.Request){

	 //only post

	 if r.Method != http.MethodPost {
		http.Error(w, "Only Post is allowed",http.StatusMethodNotAllowed)
		return;
	 }

	 //read request body
	
	 var request ExecuteRequest
	 
	 err :=  json.NewDecoder(r.Body).Decode(&request)

	 if err != nil {
		http.Error(w,"Invalid Json", http.StatusBadRequest)
		return;
	 }

	 //only js

	 if request.Language != "javascript" &&request.Language != "cpp"{
		http.Error(w,"Only javascript and cpp is supported now",http.StatusBadRequest)
		return;
	 }
	 
	 id, err := rdb.Incr(ctx,"job:counter").Result()
	 if err != nil{
		http.Error(w,"Could not create job id", http.StatusInternalServerError)
		return;
	 }
	 	
	 job := Job{
		ID:       fmt.Sprintf("%d", id),
		Language: request.Language,
		Code:     request.Code,
	}

	jobJson, err := json.Marshal(job)

	if err != nil{
		http.Error(w,"Could not encode job", http.StatusInternalServerError)
		return
	}

	err = rdb.RPush(ctx,"jobs",jobJson).Err()
	if err != nil {
		http.Error(w,"Could not add job to queue",http.StatusInternalServerError)
		return;
	}

	 w.Header().Set("Content-Type","application/json")

	 json.NewEncoder(w).Encode(map[string]string{
		"id": job.ID,
		"status":"queued",
	 })
	
	//  file,err := os.CreateTemp("","code-*.js")
	//  if err != nil{
	// 		http.Error(w,"Could not create temp file", http.StatusBadRequest)
	// 		return
	//  }
	//  defer os.Remove(file.Name())

	//  //write user code

	//  _, err = file.WriteString(request.Code)
	 
	//  if err != nil{
	// 	http.Error(w, "Could not write code", http.StatusInternalServerError)
	// 	return
	//  }

	//  file.Close()


	//  //run nodejs

	//  cmd := exec.Command("node",file.Name())
	 
	//  output,err := cmd.CombinedOutput()

	//  response := ExecuteResponse{
	// 	Output: string(output),
	//  }

	//  if err != nil {
	// 	response.Error = err.Error()
	//  }

	

}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	val, err := rdb.Get(ctx, "result:"+id).Result()
	if err == redis.Nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"id":     id,
			"status": "pending",
		})
		return
	}
	if err != nil {
		http.Error(w, "Could not read result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(val))
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func main() {

	initRedis()

	role := os.Getenv("ROLE")
	if role == "" {
		role = "all"
	}

	// Kubernetes:
	//   API Deployment    ROLE=api
	//   Worker Deployment ROLE=worker  (scale this with KEDA)
	// Local:
	//   ROLE=all (default) — one process, HTTP + one worker

	if role == "worker" {
		fmt.Println("Starting as worker")
		worker()
		return
	}

	if role == "all" {
		go worker()
	}

	http.HandleFunc("/hello", withCORS(hello))
	http.HandleFunc("/execute", withCORS(executeHandler))
	http.HandleFunc("/result", withCORS(resultHandler))

	fmt.Println("Server starting at port http://localhost:3001  ")

	http.ListenAndServe(":3001",nil)

}