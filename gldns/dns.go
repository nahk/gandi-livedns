package gldns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nahk/gandi-livedns/iputil"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

const RecordTypeIpV4 = "A"
const RecordTypeIpV6 = "AAAA"

const GandiDnsApiUrl = "https://dns.api.gandi.net/api/v5"

type Domain struct {
	Domain string `json:"domain"`
	Zone   string `json:"zone"`
	Name   string `json:"dns_name"`
}

type Record struct {
	Type   string   `json:"rrset_type"`
	TTL    int      `json:"rrset_ttl"`
	Name   string   `json:"rrset_name"`
	Href   string   `json:"rrset_href"`
	Values []string `json:"rrset_values"`
}

type Records struct {
	Items []Record `json:"items"`
}

// ApiKey for Gandi Live DNS Api
var ApiKey string

// Domains holds all the domains to process
var Domains []Domain

// Client used for HTTP Calls
var Client = &http.Client{}

func UpdateDnsRecords(ip iputil.Ip) {
	for _, domain := range Domains {
		// @Todo: parallelize
		updateDnsRecords(domain, ip)
	}
}

// @Todo: Validate IPs
func updateDnsRecords(domain Domain, ip iputil.Ip) (err error) {
	records, err := fetchDnsRecords(domain)
	if err != nil {
		log.WithField("error", err).Fatal("Error while fetching DNS records")
	}

	for _, record := range records.Items {
		switch record.Type {
		case RecordTypeIpV4:
			// Update IPv4
			if ip.V4 != "" {
				record.Values[0] = ip.V4
			}
		case RecordTypeIpV6:
			// Update IPv6
			if ip.V6 != "" {
				record.Values[0] = ip.V6
			}
		}
	}

	records.update(domain)

	return
}

func (records Records) update(domain Domain) (err error) {
	log.WithFields(log.Fields{
		"records": records,
		"domain":  domain.Name,
	}).Info("Updating records")

	payload, err := json.Marshal(records)
	log.Debug(string(payload))
	if err != nil {
		return err
	}

	url := GandiDnsApiUrl + "/zones/" + domain.Zone + "/records"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", ApiKey)

	log.WithFields(log.Fields{
		"request": req,
		"payload": string(payload),
	}).Debug("Request Gandi Live DNS API")
	resp, err := Client.Do(req)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"code": resp.StatusCode,
	}).Debug("API Response: ", string(body))

	if resp.StatusCode >= 300 {
		log.WithFields(log.Fields{
			"records":  records,
			"response": resp,
		}).Error("Something went wrong while updating a records")

		return fmt.Errorf("something went wrong while updating domain %s", domain.Name)
	}

	return
}

func fetchDnsRecords(domain Domain) (r Records, err error) {
	url := GandiDnsApiUrl + "/zones/" + domain.Zone + "/records/" + domain.Name
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Add("X-Api-Key", ApiKey)

	resp, err := Client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var records []Record
	err = json.Unmarshal(body, &records)
	if err != nil {
		return
	}

	r.Items = records

	return
}
