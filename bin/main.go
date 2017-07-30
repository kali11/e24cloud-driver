package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	//"github.com/kali11/e24cloud-driver"
	"github.com/kali11/e24cloud-driver"
	//"fmt"
	//"github.com/aws/aws-sdk-go/aws/client"
)

func main() {
	plugin.RegisterDriver(new(e24cloud.Driver))
	//client := e24cloud.GetClient("I1tZ5ygcYE8cU8Jk5pYGXp9S9twl9GRJ", "YdGapbDiiTK2ZqXf1WgaEU3XLQlcOSfmy6E7CGwn", "eu-poland-1poznan")
	//client.GetTemplates()
	//id := client.GetKeyIdByName("piotrkey")
	//client.CreateMachine("test_machine2", 1, 512, id)
	//machine := client.GetMachine("bb2a9826-0165-447e-8ae0-a7d06a4b89d9")
	//fmt.Println(id)
	//fmt.Println(client.DeleteMachine("bb2a9826-0165-447e-8ae0-a7d06a4b89d9"))
	//fmt.Println(details)


}