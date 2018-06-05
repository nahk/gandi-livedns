package main

import (
	"encoding/json"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/nahk/gandi-livedns/gldns"
	"github.com/nahk/gandi-livedns/iputil"
)

type config struct {
	Domains []gldns.Domain `json:"domains"`
}

func main() {
	ip, err := iputil.Lookup()
	log.Info("IP lookup: ", ip)
	if err != nil {
		log.Fatal("Unable to lookup the IP, make sure your connection is working")
	}

	// @Todo: Cache IP to avoid unnecessary IP lookup
	//cachedIp :=

	gldns.UpdateDnsRecords(ip)
}

func init() {
	log.SetLevel(log.DebugLevel)

	parseConfig()

	gldns.ApiKey = os.Getenv("GANDI_API_KEY")
	if gldns.ApiKey == "" {
		log.Fatal("Export your Gandi API Key as an environment variable")
	}

	gldns.Client = &http.Client{
		Timeout: 3 * time.Second,
	}
}

func parseConfig() config {
	var config config
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.WithFields(log.Fields{
			"full_message": err,
		}).Fatal("Fatal error while reading configuration file")
	}
	json.Unmarshal([]byte(configFile), &config)

	gldns.Domains = config.Domains

	log.WithFields(log.Fields{
		"value": config,
	}).Debug("Configuration loaded")

	return config
}
