package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
)

const (
	cfapivalidateurl  = "https://api.cloudflare.com/client/v4/user/tokens/verify"
	cfgetzoneurl      = "https://api.cloudflare.com/client/v4/zones?name=%s"
	cfgetrecordurl    = "https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s"
	cfupdaterecordurl = "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s"
)

var (
	optfile  = flag.String("optfile", "dyndns.cfg", "Path to the options file")
	quiet    = flag.Bool("quiet", false, "Suppress output")
	cfapikey = flag.String("cfapikey", "", "Cloudflare API key")
	cfemail  = flag.String("cfemail", "", "Cloudflare email address")
	domain   = flag.String("domain", "", "Domain to update")
	record   = flag.String("record", "", "Record to update")
	ipapi    = flag.String("ipapi", "https://api.ipify.org", "IP API URL")
)

func checkFlag(flag *string, key string, options map[string]string, errMsg string) {
	if *flag == "" {
		*flag = options[key]
	}

	if *flag == "" {
		panic(errMsg)
	}
}

func quietLog(msg string) {
	if !*quiet {
		println(msg)
	}
}

type Response struct {
	Result json.RawMessage `json:"result"`
}

func getJSON(url string, headers map[string]string, out interface{}) error {
	quietLog(fmt.Sprintf("Sending GET request to %s with headers %v", url, headers))
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		quietLog(fmt.Sprintf("Received non-OK status code %d with body: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return json.Unmarshal(response.Result, out)
}

func main() {
	flag.Parse()

	config, err := parseConfig(*optfile)
	if err != nil {
		panic(err)
	}

	checkFlag(cfapikey, "cfapikey", config.StringOptions, "No Cloudflare API key provided")
	checkFlag(cfemail, "cfemail", config.StringOptions, "No Cloudflare email address provided")
	checkFlag(domain, "domain", config.StringOptions, "No domain provided")
	checkFlag(record, "record", config.StringOptions, "No record provided")
	checkFlag(ipapi, "ipapi", config.StringOptions, "No IP API URL provided")

	if config.BoolOptions["quiet"] {
		*quiet = true
	}

	authHeaders := map[string]string{
		"X-Auth-Email":  *cfemail,
		"Authorization": "Bearer " + *cfapikey,
		"Content-Type":  "application/json",
	}

	var result map[string]interface{}
	err = getJSON(cfapivalidateurl, authHeaders, &result)
	if err != nil {
		panic(err)
	}

	quietLog("Cloudflare API key validated")

	var zones []map[string]interface{}
	err = getJSON(fmt.Sprintf(cfgetzoneurl, *domain), authHeaders, &zones)
	if err != nil || len(zones) == 0 {
		panic("Failed to get zone id")
	}

	zoneID := zones[0]["id"].(string)

	quietLog("Retrieved zone id: " + zoneID)

	var records []map[string]interface{}
	err = getJSON(fmt.Sprintf(cfgetrecordurl, zoneID, *record), authHeaders, &records)
	if err != nil {
		panic(err)
	}
	if len(records) == 0 {
		panic("Failed to get record id (no records found))")
	}

	recordID := records[0]["id"].(string)

	quietLog("Retrieved record id: " + recordID)

	quietLog("Sending GET request to " + *ipapi)
	resp, err := http.Get(*ipapi)
	if err != nil {
		panic("Failed to get current IP")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		quietLog(fmt.Sprintf("Received non-OK status code %d with body: %s", resp.StatusCode, string(body)))
		panic("Failed to get current IP")
	}

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Failed to read current IP")
	}

	quietLog("Current IP: " + string(ip))

	data := map[string]interface{}{
		"type":    records[0]["type"].(string),
		"name":    records[0]["name"].(string),
		"content": string(ip),
		"ttl":     records[0]["ttl"].(float64),
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		panic("Failed to marshal update data")
	}

	quietLog(fmt.Sprintf("Sending PUT request to %s with headers %v and body: %s", fmt.Sprintf(cfupdaterecordurl, zoneID, recordID), authHeaders, string(dataBytes)))
	req, err := http.NewRequest("PUT", fmt.Sprintf(cfupdaterecordurl, zoneID, recordID), bytes.NewBuffer(dataBytes))
	if err != nil {
		panic("Failed to create update request")
	}

	for k, v := range authHeaders {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		quietLog(fmt.Sprintf("Received non-OK status code %d with body: %s", resp.StatusCode, string(body)))
		panic("Failed to update record")
	}

	quietLog("Record updated successfully")
}
