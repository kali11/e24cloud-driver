package e24cloud

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/log"
	"time"
	"github.com/docker/machine/libmachine/ssh"
	"io/ioutil"
	"github.com/docker/machine/libmachine/mcnutils"
	"os"
)

type Driver struct {
	*drivers.BaseDriver
	MockState  state.State
	ApiKey     string
	ApiSecret  string
	InstanceId string
	Region     string
	SSHKeyId   int
	SSHKey     string
}

// Flags - driver params passed from command line
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "E24CLOUD_APIKEY",
			Name: "e24cloud_apikey",
			Usage: "e24cloud api key",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "E24CLOUD_APISECRET",
			Name: "e24cloud_apisecret",
			Usage: "e24cloud api secret",
			Value: "",
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "e24cloud"
}

// Set driver's properties based on flags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ApiKey = flags.String("e24cloud_apikey")
	d.ApiSecret = flags.String("e24cloud_apisecret")
	d.Region = "eu-poland-1warszawa"
	d.SSHKey = "/home/piotr/Pulpit/id_rsa"
	return nil
}

func (d *Driver) GetURL() (string, error) {
	log.Info("GetURL")
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	log.Info("GetSSHHostname")
	log.Info(d.GetSSHKeyPath())
	log.Info(d.GetIP())
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	log.Info("GetSSHUSername")
	return "e24"
}

func (d *Driver) publicSSHKeyPath() string {
	log.Info("publicSSHKeyPath")
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) createSSHKey() (string, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

func (d *Driver) GetState() (state.State, error) {
	client := GetClient(d.ApiKey, d.ApiSecret, d.Region)
	machine := client.GetMachine(d.InstanceId)
	switch machine.State {
	case "online":
		return state.Running, nil
	case "offline":
		return state.Stopped, nil
	case "installing":
		return state.Starting, nil
	case "deleting":
		return state.Stopping, nil
	}
	return state.None, nil
}


// Create instance
func (d *Driver) Create() error {
	log.SetDebug(true)
	log.Info("Creating e24cloud instance...")

	//log.Info("Creating SSH key...")
	//key, err := d.createSSHKey()
	//if err != nil {
	//	return err
	//}


	// tworzenie klienta
	client := GetClient(d.ApiKey, d.ApiSecret, d.Region)

	d.SSHKeyId = client.GetKeyIdByName("piotrkey")
	if err := copySSHKey(d.SSHKey, d.GetSSHKeyPath()); err != nil {
		return err
	}
	// call do api
	vm_id := client.CreateMachine(d.MachineName, 1, 512, d.SSHKeyId)

	// zapisanie machineId
	d.InstanceId = vm_id

	// poczekanie na adres ip
	log.Info("Waiting for the ip address...")
	var Ip string
	for {
		machine := client.GetMachine(vm_id)
		if machine.State == "online" {
			Ip = machine.Ip.Ip
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}
	log.Infof("Machine is online! machine_id = %s, IP address = %s", vm_id, Ip)
	d.IPAddress = Ip
	d.GetSSHKeyPath()
	return nil
}

func (d *Driver) Start() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) Stop() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) Restart() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) Kill() error {
	d.MockState = state.Stopped
	log.Info("kill...")
	return nil
}

func (d *Driver) Remove() error {
	client := GetClient(d.ApiKey, d.ApiSecret, d.Region)
	client.DeleteMachine(d.InstanceId)
	return nil
}

func (d *Driver) Upgrade() error {
	return nil
}

// copied from digitalocean driver
func copySSHKey(src, dst string) error {
	if err := mcnutils.CopyFile(src, dst); err != nil {
		return fmt.Errorf("unable to copy ssh key: %s", err)
	}

	if err := os.Chmod(dst, 0600); err != nil {
		return fmt.Errorf("unable to set permissions on the ssh key: %s", err)
	}

	return nil
}
