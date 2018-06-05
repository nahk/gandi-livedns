package iputil

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type Ip struct {
	V4 string `json:"v4"`
	V6 string `json:"v6"`
}

// Lookup gets the external IP of the machine.
// IPv6 is a bonus and IPv4 should be enough.
// @Todo: Parallelize
// @Todo: Validate IPs
func Lookup() (ip Ip, err error) {
	ip.V4, err = Get("https://v4.ifconfig.co/", 4)
	if err != nil {
		return Ip{}, err
	}

	var err6 error
	ip.V6, err6 = Get("https://v6.ifconfig.co/", 6)
	if err6 != nil {
		log.Warn("Unable to lookup the IP v6")
	}

	return
}

// get an URL body
func Get(url string, protocolVersion uint) (string, error) {
	protocol := "tcp4"
	if protocolVersion == 6 {
		protocol = "tcp6"
	}

	dialer := &net.Dialer{
		DualStack: false,
	}
	tr := &http.Transport{
		DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, protocol, addr)
		},
		Dial: dialer.Dial,
	}
	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}
