package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

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

	 if request.Language != "javascript"{
		http.Error(w,"Only js is supported now",http.StatusBadRequest)
		return;
	 }

	 file,err := os.CreateTemp("","code-*.js")
	 if err != nil{
			http.Error(w,"Could not create temp file", http.StatusBadRequest)
			return
	 }
	 defer os.Remove(file.Name())

	 //write user code

	 _, err = file.WriteString(request.Code)
	 
	 if err != nil{
		http.Error(w, "Could not write code", http.StatusInternalServerError)
		return
	 }

	 file.Close()


	 //run nodejs

	 cmd := exec.Command("node",file.Name())
	 
	 output,err := cmd.CombinedOutput()

	 response := ExecuteResponse{
		Output: string(output),
	 }

	 if err != nil {
		response.Error = err.Error()
	 }

	 w.Header().Set("Content-Type","application/json")

	 json.NewEncoder(w).Encode(response)

}

func main() {
	http.HandleFunc("/hello",hello)
	http.HandleFunc("/execute", executeHandler)

	fmt.Println("Server starting at port http://localhost:3000  ")

	http.ListenAndServe(":3001",nil)
}