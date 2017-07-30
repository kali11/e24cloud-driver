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
	"strconv"
)

type Driver struct {
	*drivers.BaseDriver
	MockState  state.State
	ApiKey     string
	ApiSecret  string
	InstanceId string
	Region     string
	SSHKeyId   int
	SSHKeyPath     string
	SSHKeyName string
	Cpus	string
	Ram	string
	Client *Client
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
		mcnflag.StringFlag{
			EnvVar: "E24CLOUD_REGION",
			Name: "e24cloud_region",
			Usage: "eu-poland-1warszawa or eu-poland-1poznan",
			Value: "eu-poland-1warszawa",
		},
		mcnflag.StringFlag{
			EnvVar: "E24CLOUD_SSHKEYNAME",
			Name: "e24cloud_sshkeyname",
			Usage: "e24cloud ssh key name",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "E24CLOUD_SSHKEYPATH",
			Name: "e24cloud_sshkeypath",
			Usage: "path to ssh private key for given key name",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "E24CLOUD_CPUS",
			Name: "e24cloud_cpus",
			Usage: "numer of cpus. Max 16",
			Value: "1",
		},
		mcnflag.StringFlag{
			EnvVar: "E24CLOUD_RAM",
			Name: "e24cloud_ram",
			Usage: "amount of RAM in MB. Max 32000",
			Value: "512",
		},
	}
}

func (d *Driver) GetClient() *Client {
	if d.Client == nil {
		client := new(Client)
		client.url = "https://" + d.Region + ".api.e24cloud.com/v2/"
		client.region = d.Region
		client.apiKey = d.ApiKey
		client.apiSecret = d.ApiSecret
		d.Client = client
	}
	return d.Client
}

func (d *Driver) PreCreateCheck() error {
	if d.Region == "" {
		return fmt.Errorf("Empty region")
	}
	if d.Region != "eu-poland-1warszawa" && d.Region != "eu-poland-1poznan" {
		return fmt.Errorf("Invalid region: %s", d.Region)
	}
	if d.ApiSecret == "" {
		return fmt.Errorf("Empty API secret")
	}
	if d.ApiKey == "" {
		return fmt.Errorf("Empty API key")
	}
	if d.SSHKeyName == "" {
		return fmt.Errorf("Empty ssh key name")
	}
	if d.SSHKeyPath == "" {
		return fmt.Errorf("Empty ssh private key path")
	}
	if _, err := os.Stat(d.SSHKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("SSH private key does not exist: %q", d.SSHKeyPath)
	}
	cpus, err := strconv.Atoi(d.Cpus)
	if err != nil {
		return err
	}
	if cpus > 16 {
		return fmt.Errorf("Max Cpus can be 16")
	}
	ram, err := strconv.Atoi(d.Ram)
	if err != nil {
		return err
	}
	if ram > 32000 {
		return fmt.Errorf("Max RAM can be 32 GB")
	}
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "e24cloud"
}

// Set driver's properties based on flags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ApiKey = flags.String("e24cloud_apikey")
	d.ApiSecret = flags.String("e24cloud_apisecret")
	d.Region = flags.String("e24cloud_region")
	d.SSHKeyPath = flags.String("e24cloud_sshkeypath")
	d.SSHKeyName = flags.String("e24cloud_sshkeyname")
	d.Cpus = flags.String("e24cloud_cpus")
	d.Ram = flags.String("e24cloud_ram")
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
	client := d.GetClient()
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

	client := d.GetClient()

	SSHKeyId, err := client.GetKeyIdByName(d.SSHKeyName)
	if err != nil {
		return err
	}
	d.SSHKeyId = SSHKeyId
	if err := copySSHKey(d.SSHKeyPath, d.GetSSHKeyPath()); err != nil {
		return err
	}

	vm_id, err := client.CreateMachine(d.MachineName, d.Cpus, d.Ram, d.SSHKeyId)
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
	d.GetClient().PowerOn(d.InstanceId)
	return nil
}

func (d *Driver) Stop() error {
	d.GetClient().ShutDown(d.InstanceId)
	return nil
}

func (d *Driver) Restart() error {
	d.GetClient().Reboot(d.InstanceId)
	return nil
}

func (d *Driver) Kill() error {
	return fmt.Errorf("Killing machine is not supported in e24cloud")
}

func (d *Driver) Remove() error {
	d.GetClient().DeleteMachine(d.InstanceId)
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
