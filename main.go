package main

import (
	"github.com/hpcloud/tail"
    "fmt"
    "flag"
    "log"
    cb "github.com/clearblade/Go-SDK"
)


var (
	platURL      string
	messURL      string
	sysKey       string
	sysSec       string
	deviceName   string
	activeKey    string
	listenPort   string
	topicName    string
	enableTLS    bool
	tlsCertPath  string
	tlsKeyPath   string
	deviceClient *cb.DeviceClient
	filename string
)

func init() {
	flag.StringVar(&sysKey, "systemKey", "", "system key (required)")
	flag.StringVar(&sysSec, "systemSecret", "", "system secret (required)")
	flag.StringVar(&platURL, "platformURL", "", "platform url (required)")
	flag.StringVar(&messURL, "messagingURL", "", "messaging URL")
	flag.StringVar(&deviceName, "deviceName", "", "name of device (required)")
	flag.StringVar(&activeKey, "activeKey", "", "active key (password) for device (required)")
	flag.StringVar(&listenPort, "receiverPort", "", "receiver port for adapter (required)")
	flag.StringVar(&topicName, "topicName", "webhook-adapter/received", "topic name to publish received HTTP requests to (defaults to webhook-adapter/received)")
	flag.BoolVar(&enableTLS, "enableTLS", false, "enable TLS on http listener (must provide tlsCertPath and tlsKeyPath params if enabled)")
	flag.StringVar(&tlsCertPath, "tlsCertPath", "", "path to TLS .crt file (required if enableTLS flag is set)")
	flag.StringVar(&tlsKeyPath, "tlsKeyPath", "", "path to TLS .key file (required if enableTLS flag is set)")
	flag.StringVar(&filename, "file", "nohup.out", "URL Path for inbound webhook URL, ex /abcdef/endpoint1")
}

func usage() {
	log.Printf("Usage: webhook-adapter [options]\n\n")
	flag.PrintDefaults()
}

func main(){
    flag.Usage = usage
	flag.Parse()

	
    t, _ := tail.TailFile(filename, tail.Config{
        Follow: true,
        ReOpen: true})


    for line := range t.Lines {
        
        fmt.Println(line.Text)
    }

}