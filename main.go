package main

import (
	"github.com/hpcloud/tail"
    "fmt"
    "flag"
    "log"
    cb "github.com/clearblade/Go-SDK"
    "regexp"
    "os"
)


var (
	platURL      string
	messURL      string
	sysKey       string
	sysSec       string
	deviceName   string
	activeKey    string
	topicName    string
	enableTLS    bool
	tlsCertPath  string
	tlsKeyPath   string
	deviceClient *cb.DeviceClient
    userClient *cb.UserClient
    filename string
    email string
    password string

)

func init() {
	flag.StringVar(&sysKey, "systemKey", "", "system key (required)")
	flag.StringVar(&sysSec, "systemSecret", "", "system secret (required)")
	flag.StringVar(&platURL, "platformURL", "", "platform url (required)")
	flag.StringVar(&messURL, "messagingURL", "", "messaging URL")
	flag.StringVar(&deviceName, "deviceName", "", "name of device (required)")
	flag.StringVar(&activeKey, "activeKey", "", "active key (password) for device (required)")
    flag.StringVar(&email, "email", "", "name of device (required)")
	flag.StringVar(&password, "password", "", "active key (password) for device (required)")
    flag.StringVar(&topicName, "topicName", "deployment-adapter/logs", "topic name to publish received HTTP requests to (defaults to webhook-adapter/received)")
	flag.BoolVar(&enableTLS, "enableTLS", false, "enable TLS on http listener (must provide tlsCertPath and tlsKeyPath params if enabled)")
	flag.StringVar(&tlsCertPath, "tlsCertPath", "", "path to TLS .crt file (required if enableTLS flag is set)")
	flag.StringVar(&tlsKeyPath, "tlsKeyPath", "", "path to TLS .key file (required if enableTLS flag is set)")
	flag.StringVar(&filename, "file", "nohup.out", "URL Path for inbound webhook URL, ex /abcdef/endpoint1")
}

func usage() {
	log.Printf("Usage: deployment-adapter [options]\n\n")
	flag.PrintDefaults()
}

func validateFlags() {
	flag.Parse()
	if sysKey == "" || sysSec == "" || platURL == "" || (deviceName == "" && email == "") || (activeKey == "" && password == "") {
		log.Printf("Missing required flags\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if enableTLS && (tlsCertPath == "" || tlsKeyPath == "") {
		log.Printf("tlsCertPath and tlsKeyPath are required if TLS is enabled\n")
		flag.Usage()
		os.Exit(1)
	}

}


func findIfMatchesStart(line string) bool {
    re := regexp.MustCompile(`controller.go:136: ADAPTOR FILE DEPLOY: Stopped adaptor`)
    return re.Match([]byte(line))
}

func findIfMatchesEnd(line string) bool {
    re := regexp.MustCompile(`controller.go:149: ADAPTOR FILE DEPLOY: Started adaptor`)
    return re.Match([]byte(line))
}



func publishLog(line string) {
    b := []byte(line)
    if err := userClient.Publish(topicName, b, 2); err != nil {
		log.Printf("Unable to publish log: %s\n", err.Error())
	}
}


func main(){
    flag.Usage = usage
    validateFlags()

	//deviceClient = cb.NewDeviceClient(sysKey, sysSec, deviceName, activeKey)
    userClient = cb.NewUserClient(sysKey, sysSec, email, password)
	if platURL != "" {
		log.Println("Setting custom platform URL to: ", platURL)
		userClient.HttpAddr = platURL
	}

	if messURL != "" {
		log.Println("Setting custom messaging URL to: ", messURL)
		userClient.MqttAddr = messURL
	}

	log.Println("Authenticating to platform with device: ", deviceName)

	if err := userClient.Authenticate(); err != nil {
		log.Fatalf("Error authenticating: %s\n", err.Error())
	}

	if err := userClient.InitializeMQTT("webhookadapter_"+deviceName, "", 30, nil, nil); err != nil {
		log.Fatalf("Unable to initialize MQTT: %s\n", err.Error())
	}
	log.Printf("MQTT connected and adapter about to tail on file: %s\n", filename)

    
    
    t, _ := tail.TailFile(filename, tail.Config{
        Follow: true,
        ReOpen: true})
    
    startFlag := false
    endFlag := false    
    
    
    for line := range t.Lines {
        if startFlag == false {
            startFlag = findIfMatchesStart(line.Text)
        }
        endFlag = findIfMatchesEnd(line.Text)
        if startFlag == true {
            fmt.Println(line.Text)
            publishLog(line.Text)
        }
        if endFlag == true {
            startFlag = false
            endFlag = false
            fmt.Printf("\n\nBreak\n\n")
        }
    }

}