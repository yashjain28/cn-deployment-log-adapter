package main

import (
	"github.com/hpcloud/tail"
    "fmt"
    "flag"
    "log"
    cb "github.com/clearblade/Go-SDK"
    "regexp"
	"os"
	"math/rand"
	"time"
	"bytes"
	"strconv"
	"encoding/json"
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
    flag.StringVar(&topicName, "topicName", "deployment/adapter/logs", "topic name to publish received HTTP requests to (defaults to webhook-adapter/received)")
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
    re := regexp.MustCompile(`ADAPTOR FILE DEPLOY: Stopped adaptor`)
    return re.Match([]byte(line))
}

func findIfMatchesEnd(line string) bool {
    re := regexp.MustCompile(`ADAPTOR FILE DEPLOY: Started adaptor`)
    return re.Match([]byte(line))
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func generateClientID() string {
	return randSeq(10)
}

func extractTimestamp(line string) string {
	re := regexp.MustCompile(`(\d{1,4}([/])\d{1,2}([/])\d{1,4}) (\d{1,2}:\d{1,2}:\d{1,2})`)
	return string(re.Find([]byte(line)))
}

func extractCommandOutput(deployLogs string) string {
	re := regexp.MustCompile(`.* ADAPTOR FILE DEPLOY: .* deploy script`)
	b := []byte(deployLogs)
	locStart := re.FindIndex(b)
	re = regexp.MustCompile(`.* ADAPTOR FILE DEPLOY: Started adaptor`)
	locEnd := re.FindIndex(b)
	runes := []rune(deployLogs)
	
	if(locEnd!=nil && locStart != nil){
		return string(runes[locStart[1]:locEnd[0]])
	}
	return ""
}

func extractExitStatus(deployLogs string) int {
	re := regexp.MustCompile(`Error running deploy script .*exit status (\d{1,3})`)
	//b := []byte(deployLogs)
	submatch := re.FindStringSubmatch(deployLogs)
	if( len(submatch) == 2) {
		i, err := strconv.Atoi(submatch[1])
		if err != nil {
			log.Println("Error")
		}
		return i
	}
    fmt.Printf("%#v\n", submatch)

	return 0
}

func handleLogs(deployLogs string)  {
	fmt.Println("Deploy logs: ", deployLogs)
	ts:= extractTimestamp(deployLogs)
	fmt.Println("Date:", ts)
	commandOutput := extractCommandOutput(deployLogs)
	fmt.Println("Command Output", commandOutput)
	exitStatus := extractExitStatus(deployLogs)
	errBool := false
	if exitStatus != 0 {
		errBool = true
	}
	fmt.Println("Exit Status", exitStatus)
	m := map[string]string{
		"timestamp": ts,
		"error":strconv.FormatBool(errBool),
		"command":"Sorry, I don't get that in logs",
		"commandOutput":commandOutput,
		"exitStatus":strconv.Itoa(exitStatus),
	}

	logObj, err := json.Marshal(m)   
    if err != nil {
        fmt.Println(err.Error())
        return
    }
     
    jsonStr := string(logObj)
	fmt.Println("The JSON data is:", jsonStr)
	publishLog(jsonStr)
}

func publishLog(line string) {
	log.Println("Publishing now...")
    b := []byte(line)
    if err := userClient.Publish(topicName, b, 2); err != nil {
		log.Printf("Unable to publish log: %s\n", err.Error())
	}
}


func main(){
    flag.Usage = usage
    validateFlags()

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

	clientID := generateClientID()
	if err := userClient.InitializeMQTT(clientID, "", 30, nil, nil); err != nil {
		log.Fatalf("Unable to initialize MQTT: %s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("MQTT connected and adapter about to tail on file: %s\n", filename)

    
    
    t, _ := tail.TailFile(filename, tail.Config{
		Location: &tail.SeekInfo{
			Whence:os.SEEK_END,
		},
        Follow: true,
        ReOpen: true})
    
    startFlag := false
    endFlag := false    
	
	var buffer bytes.Buffer
	
    
    for line := range t.Lines {
        if startFlag == false {
            startFlag = findIfMatchesStart(line.Text)
        }
        endFlag = findIfMatchesEnd(line.Text)
        if startFlag == true {
			//fmt.Println(line.Text)
			buffer.WriteString(line.Text + "\n")  
		}

		// if buffer.Len() > 1024 {
		// 	publishLog(buffer.String())
		// 	buffer.Reset()
		// }

        if endFlag == true {
            startFlag = false
			endFlag = false
			handleLogs(buffer.String())
			buffer.Reset()
            fmt.Printf("\n\nBreak\n\n")
        }
    }

}