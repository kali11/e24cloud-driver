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
const os_id = "2268" //debian TODO make it customable

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
	Key_id int `json:"key_id,omitempty"`
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

type SshKey struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

type Account struct {
	SshKeys []SshKey `json:"sshkeys"`
}

type AccountWrapper struct {
	Success bool `json:"success"`
	Account Account `json:"account"`
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

func (c* Client) GetMachine(vm_id string) (*MachineDetails, error) {
	response, err := c.SendRequest("virtual-machines/" + vm_id, "GET", []byte(""))
	if err != nil {
		return nil, err
	}
	var machine MachineDetailsWrapper
	json.Unmarshal([]byte(response), &machine)
	return &machine.Machine, nil
}

// returns true if machine deleted successfully
func (c* Client) DeleteMachine(vm_id string) (bool, error) {
	response, err := c.SendRequest("virtual-machines/" + vm_id, "DELETE", []byte(""))
	if err != nil {
		return false, err
	}
	var success Success
	json.Unmarshal([]byte(response), &success)
	return success.Success, nil
}

// Create machine and return machine ID
func (c *Client) CreateMachine(name string, cpus, ram int, key_id int) (string, error) {
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
		Os: os_id,
		Key_id: key_id,
	}

	str, err := json.Marshal(CreateMachineWrapper{machine})
	if err != nil {
		return "", err
	}

	response, err := c.SendRequest("virtual-machines", "PUT", str)
	if err != nil {
		return "", err
	}

	var r CreateMachineResponse
	json.Unmarshal([]byte(response), &r)
	return r.MachineId.Id, nil
}

func (c *Client) GetRegions() (string, error) {
	response, err := c.SendRequest("regions", "GET", []byte(""))
	if err != nil {
		return "", err
	}
	return string(response), nil
}

func (c *Client) GetTemplates() (string, error) {
	response, err := c.SendRequest("templates", "GET", []byte(""))
	if err != nil {
		return "", err
	}
	var x map[string]interface{}
	json.Unmarshal([]byte(response), &x)

	return string(response), nil
}

func (c *Client) GetAccount() (*Account, error) {
	response, err := c.SendRequest("account", "GET", []byte(""))
	if err != nil {
		return nil, err
	}
	var account AccountWrapper
	json.Unmarshal([]byte(response), &account)
	return &account.Account, nil
}

func (c *Client) GetKeyIdByName(keyName string) (int, error) {
	account, err := c.GetAccount()
	if err != nil {
		return 0, err
	}
	for {
		sshkey := account.SshKeys[0]
		if sshkey.Name == keyName {
			return sshkey.Id, nil
		}
	}
	return 0, fmt.Errorf("Cannot find SshKey with name = %s", keyName)
}

func (c *Client) SendRequest(path, method string, reqBody []byte) ([]byte, error) {
	req, err := CreateRequest(method, c.url + path, c.apiKey, c.apiSecret, reqBody)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error while making API call. Request: %s, Status code: %s, response: %s", req, resp.StatusCode, body)
	}
	return body, nil
}

func CreateRequest(method, url, apiKey, apiSecret string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	x_date := time.Now().Format(time.RFC1123)
	req.Header.Set("X-Date", x_date)
	requestString := GetRequestString(method, url, x_date, string(body))
	sign := ComputeHmac256(requestString, apiSecret)
	req.Header.Set("Authorization", apiKey + ":" + sign)

	return req, nil
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