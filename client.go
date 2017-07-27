package e24cloud

import (
	"fmt"
	"net/http"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"bytes"
)

const poznan_zone_id = "24e12e20-0851-5354-e2c3-04b16c4c9c45"
const warsaw_zone_id  = "0daeacdb-7b1b-f510-d44e-ec8fd457d7aa"

type Client struct {
	url string
	region string
	apiKey string
	apiSecret string
}

type CreateMachine struct {
	Cpus int `json:"cpus"`
	Ram int `json:"ram"`
	Zone_id string `json:"zone_id"`
	Name string `json:"name"`
	Boot_type string `json:"boot_type,omitempty"`
	Cdrom string `json:"cdrom,omitempty"`
	Os string `json:"os"`
	Password string `json:"password,omitempty"`
	Key_id string `json:"key_id,omitempty"`
	User_data string `json:"user_data,omitempty"`
}

type CreateMachineWrapper struct {
	Create_vm CreateMachine `json:"create_vm"`
}

type MachineDetails struct {
	Id string `json:"id"`
	Cpu int `json:"cores"`
	Ram int `json:"ram"`
	State string `json:"state"`
	Ip MachineDetailsIP `json:"public_interface"`
}

type MachineDetailsIP struct {
	Ip string `json:"primary_ip_ipv4address"`
}

type MachineDetailsWrapper struct {
	Success bool `json:"success"`
	Machine MachineDetails `json:"virtual_machine"`
}

type CreateMachineResponse struct {
	Success bool `json:"success"`
	MachineId CreateMachineResponseId `json:"virtual_machine"`
}

type CreateMachineResponseId struct {
	Id string `json:"id"`
}

type Success struct {
	Success bool `json:"success"`
}
func GetClient(apiKey, apiSecret, region string) *Client {
	client := new(Client)
	client.url = "https://" + region + ".api.e24cloud.com/v2/"
	client.region = region
	client.apiKey = apiKey
	client.apiSecret = apiSecret
	return client
}

func (c* Client) GetMachine(vm_id string) *MachineDetails {
	response := c.SendRequest("virtual-machines/" + vm_id, "GET", []byte(""))
	var machine MachineDetailsWrapper
	json.Unmarshal([]byte(response), &machine)
	return &machine.Machine
}

// returns true if machine deleted successfully
func (c* Client) DeleteMachine(vm_id string) bool {
	response := c.SendRequest("virtual-machines/" + vm_id, "DELETE", []byte(""))
	var success Success
	json.Unmarshal([]byte(response), &success)
	return success.Success
}

// Create machine and return machine ID
func (c *Client) CreateMachine(name string, cpus, ram int) string {
	zone_id := ""
	if c.region == "eu-poland-1warszawa" {
		zone_id = warsaw_zone_id
	} else {
		zone_id = poznan_zone_id
	}
	machine := CreateMachine{
		Cpus: cpus,
		Ram: ram,
		Zone_id: zone_id,
		Name: name,
		Boot_type: "image",
		Os: "2528", //ubuntu
		Password: "kali1",
	}

	str, err := json.Marshal(CreateMachineWrapper{machine})
	if err != nil {
		fmt.Println("Cannot marshal CreateVM struct")
	}

	response := c.SendRequest("virtual-machines", "PUT", str)

	var r CreateMachineResponse
	json.Unmarshal([]byte(response), &r)
	fmt.Println(r)
	return r.MachineId.Id
}

func (c *Client) GetRegions() string {
	response := string(c.SendRequest("regions", "GET", []byte("")))
	fmt.Println(response)
	return response
}

func (c *Client) GetTemplates() string {
	response := c.SendRequest("templates", "GET", []byte(""))
	var x map[string]interface{}
	json.Unmarshal([]byte(response), &x)

	fmt.Println(x["templates"])
	return string(response)
}

func (c *Client) SendRequest(path, method string, reqBody []byte) []byte {
	req := GetRequest(method, c.url + path, c.apiKey, c.apiSecret, reqBody)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(resp)
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body
}

func GetRequest(method, url, apiKey, apiSecret string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	x_date := time.Now().Format(time.RFC1123)
	req.Header.Set("X-Date", x_date)
	requestString := GetRequestString(method, url, x_date, string(body))
	sign := ComputeHmac256(requestString, apiSecret)
	req.Header.Set("Authorization", apiKey + ":" + sign)

	return req
}

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func GetRequestString(method, uri, x_date, body string) string {
	parsedUrl, err := url.Parse(uri)
	if err != nil {
		fmt.Println("Cannot parse url: " + uri)
		fmt.Println(err)
		return ""
	}
	result := ""
	result += method
	result += "\n"
	result += parsedUrl.Host
	result += "\n"
	result += x_date
	result += "\n"
	result += parsedUrl.Path
	result += ""
	result += "\n"
	result += body
	return result
}