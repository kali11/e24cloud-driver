package e24cloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/log"
	"time"
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
	SSHKeyName string
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
	d.SSHKeyName = "piotrkey"
	return nil
}

func (d *Driver) GetURL() (string, error) {
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
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	return "e24"
}

func (d *Driver) GetState() (state.State, error) {
	client := GetClient(d.ApiKey, d.ApiSecret, d.Region)
	machine, err := client.GetMachine(d.InstanceId)
	if err != nil {
		return state.None, err
	}
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

	client := GetClient(d.ApiKey, d.ApiSecret, d.Region)

	SSHKeyId, err := client.GetKeyIdByName(d.SSHKeyName)
	if err != nil {
		return err
	}
	d.SSHKeyId = SSHKeyId
	if err := copySSHKey(d.SSHKey, d.GetSSHKeyPath()); err != nil {
		return err
	}

	vm_id, err := client.CreateMachine(d.MachineName, 1, 512, d.SSHKeyId)
	if err != nil {
		log.Error("Cannot create machine by API")
		return err
	}

	d.InstanceId = vm_id

	log.Info("Waiting for the ip address...")
	var Ip string
	for {
		machine, err := client.GetMachine(vm_id)
		if err != nil {
			log.Errorf("Cannot aquire details about machine with id = %s", vm_id)
			return err
		}
		if machine.Ip.Ip != "" {
			Ip = machine.Ip.Ip
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}
	log.Infof("Machine IP address = %s", Ip)
	d.IPAddress = Ip
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
