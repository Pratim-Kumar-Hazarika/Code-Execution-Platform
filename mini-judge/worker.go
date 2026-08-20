package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)


func worker(){
	fmt.Println("Worker started")

	for {
		res, err := rdb.BLPop(ctx,0,"jobs").Result()
		if err != nil{
			fmt.Println("BLop error: ",err)
			time.Sleep(time.Second)
			continue
		}

		var job Job
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil{
			fmt.Println("Bad job JSON: ",err)
			continue
		}
		fmt.Println("Running job", job.ID)

		output, runErr := runInContainer(job)
		fmt.Println("Job", job.ID, "output:", string(output))
		status:= "ok"
		errText := ""
		if runErr != nil{
			status = "error"
			errText = runErr.Error()
			fmt.Println("Job", job.ID, "error:", errText)
		}

		payload,_ := json.Marshal(map[string]string{
			"id": job.ID,
			"status":status,
			"output": string(output),
			"error": errText,
		})

		if err := rdb.Set(ctx,"result:"+job.ID,payload,time.Hour).Err(); err != nil {
			fmt.Println("Could not save result:", err)
		}

	}
}

func runInContainer(job Job)([]byte,error){
	timeout := 5 * time.Second
	if job.Language == "cpp" {
		timeout = 10 * time.Second
	}
	cmCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd

	switch job.Language {
	case "javascript":
		cmd = exec.CommandContext(cmCtx,
			"docker","run","--rm","-i",
			"--network","none",
			"--memory","128m",
			"--read-only",
			"--tmpfs","/tmp:rw,exec,nosuid,size=64m",
			"node:20-alpine",
			"node",
	)
case "cpp":
	cmd = exec.CommandContext(cmCtx,
		"docker", "run", "--rm", "-i",
		"--network", "none",
		"--memory", "128m",
		"--read-only",
		"--tmpfs", "/tmp:rw,exec,nosuid,size=64m",
		"gcc:13",
		"sh", "-c",
		"cat > /tmp/main.cpp && g++ -std=c++17 -O2 -o /tmp/a.out /tmp/main.cpp && /tmp/a.out",
	)
default:
	return nil , fmt.Errorf("Unsupported langauges: %s", job.Language)

	}
	

	cmd.Stdin = strings.NewReader(job.Code)
	return cmd.CombinedOutput()

}