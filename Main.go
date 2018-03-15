package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tidwall/gjson"
)

type jwt struct {
	Key string `json:"jwt"`
}

var token jwt
var activeEndpoint = `1`

func main() {
	portainerAuth()
	deployNewStack()

}

type configuration struct {
	AppName  string
	Username string
	Password string
	Hostname string
}

var config = configuration{"Test", "user", "password", "portainerURL"} // insert correct data

// Handels the portainer authentification
func portainerAuth() {
	log.Printf("Begin authentification.")
	url := config.Hostname + "api/auth"                                                                 // Portainer URL
	var jsonStr = []byte(`{"username":"` + config.Username + `","password":"` + config.Password + `"}`) // Need to be valid for Portainer authentification
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
	log.Printf("Authentification was successful.")
}

func getSwarmClusterID() string {
	log.Println("Get SwarmID")
	url := config.Hostname + "api/endpoints/" + activeEndpoint + "/docker/info"
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

	result := gjson.GetBytes(body, "Swarm.Cluster.ID") // get nested json data
	return result.String()
}

type StackCreateRequest struct {
	Name             string `json:"Name"`
	SwarmID          string `json:"SwarmID"`
	StackFileContent string `json:"StackFileContent"`
}

func deployNewStack() {
	swarmID := getSwarmClusterID()
	stackName := config.AppName + `-app`
	url := config.Hostname + "api/endpoints/" + activeEndpoint + "/stacks"
	var newStack StackCreateRequest
	newStack.Name = stackName
	newStack.SwarmID = swarmID
	newStack.StackFileContent = string(readDockerComposeYml())
	// fmt.Println(newStack)
	bodyBytes, err := json.Marshal(newStack)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))

	if err != nil {
		panic(err)
	}

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

// readDockerComposeYml returns the docker-compose.yml as []byte
func readDockerComposeYml() []byte {
	data, err := ioutil.ReadFile("./docker-compose.yml")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil
	}
	return data
}
