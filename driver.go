package e24cloud

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/aws/endpoints"
)

type Driver struct {
	*drivers.BaseDriver
	MockState state.State
	ApiKey    string
	ApiSecret    string
	InstanceId    string
	Region    string
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
	return "", nil
}

func (d *Driver) GetSSHKeyPath() string {
	return ""
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, nil
}

func (d *Driver) GetSSHUsername() string {
	return ""
}

func (d *Driver) GetState() (state.State, error) {
	return d.MockState, nil
}


// Create instance
func (d *Driver) Create() error {
	log.Info("Creating e24cloud instance...")
	// tworzenie klienta
	//endpoint := "https://eu-poland-1warszawa.api.e24cloud.com"
	log.Info(d.ApiKey)
	log.Info(d.ApiSecret)

	myCustomResolver := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return endpoints.ResolvedEndpoint{
			URL:           "https://eu-poland-1warszawa.api.e24cloud.com",
			SigningRegion: "eu-poland-1warszawa",
		}, nil
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(d.Region),
		Credentials: credentials.NewStaticCredentials(d.ApiKey, d.ApiSecret, ""),
		//Endpoint: aws.String(endpoint),
		EndpointResolver: endpoints.ResolverFunc(myCustomResolver),
		CredentialsChainVerboseErrors: aws.Bool(true),
	}))

	log.Info("debug10")
	svc := ec2.New(sess)
	log.Info("debug20")
	//result, err := svc.RunInstances(&ec2.RunInstancesInput{
	//
	//})
	var ids []*string
	ids = append(ids, aws.String("ami-5731123e"))
	//ids[0] = aws.String("ami-5731123e")

	output, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		DryRun: aws.Bool(true),
		ImageIds: ids,
	})
	log.Info("debug30")
	if err != nil {
		log.Info("Cannot describe")
		log.Info(err)
		return err
	}
	log.Info("debug40")
	log.Info(len(output.Images))
	for _, image := range output.Images {
		log.Info(image.Name)
		log.Info(image.ImageId)
	}
	log.Info("debug50")



	// call do api
	// zapisanie machineId
	// poczekanie na adres ip i stworzenie maszynki
	// info o sukcesie
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
	return nil
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Upgrade() error {
	return nil
}

