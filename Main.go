package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/atix1906/PortainerGoDockerTest/portainerdata"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type jwt struct {
	Key string `json:"jwt"`
}

type stack struct {
	ID              string          `json:"Id"`
	Name            string          `json:"Name"`
	EntryPoint      string          `json:"EntryPoint"`
	SwarmID         string          `json:"SwarmId"`
	ProjectPath     string          `json:"ProjectPath"`
	Env             string          `json:"Env"`
	ResourceControl resourceControl `json:"ResourceControl"`
}

type resourceControl struct {
	ID                 string       `json:"Id"`
	ResoureceID        string       `json:"ResourceId"`
	SubResourceIDs     string       `json:"SubResourceIds"`
	Type               string       `json:"Type"`
	AdministratorsOnly string       `json:"AdministratorsOnly"`
	UserAccesses       []string     `json:"UserAccesses"`
	TeamAccesses       teamAccesses `json:"TeamAccesses"`
}

type teamAccesses struct {
	TeamID      string `json:"TeamId"`
	AccessLevel string `json:"AccessLevel"`
}

type stackUpdateRequest struct {
	StackFileContent string `json:"StackFileContent"`
	Env              env    `json:"Env[]"`
}

type env struct {
	name  string `json:"name"`
	value string `json:"value"`
}

var token jwt

func main() {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	_, err = cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello world"},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out)

	portainerGetStack()
}

func portainerAuth() {
	url := portainerdata.URL + "/api/auth"                                                                            // URL needs to be filled in
	var jsonStr = []byte(`{"username":"` + portainerdata.Username + `","password":"` + portainerdata.Password + `"}`) // Username and password need to be provided
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(body, &token)
}

func portainerGetStack() {
	portainerAuth()
	url := portainerdata.URL + "/api/endpoints/" + portainerdata.EndpointID + "/stacks" // URL and endpoint ID need to be filled in
	req, err := http.NewRequest("GET", url, nil)
	bearer := "Bearer " + token.Key
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var stacks []stack
	json.Unmarshal(body, &stacks)

	portainerUpdateStack(stacks[0].ID)
}

func portainerUpdateStack(stackID string) {
	data, err := ioutil.ReadFile(portainerdata.DockerComposeFile)
	if err != nil {
		log.Printf("Error: %s", err)
	}
	var newBody stackUpdateRequest
	newBody.StackFileContent = string(data)
	bodyBytes, err := json.Marshal(newBody)
	if err != nil{
		panic(err)
	}
	url := portainerdata.URL + "/api/endpoints/" + portainerdata.EndpointID + "/stacks/" + stackID // URL and endpoint ID need to be filled in
	req, err := http.NewRequest("PUT", url, bytes.NewReader(bodyBytes))
	bearer := "Bearer " + token.Key
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}
