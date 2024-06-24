package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

// 通用的 API 响应结构
type APIResponse struct {
	Request struct {
		Operation string `json:"operation"`
		IP        string `json:"ip"`
	} `json:"request"`
	Reply Reply `json:"reply"`
}

// 通用的 Reply 结构，包含两种可能的回复类型
type Reply struct {
	Code           int         `json:"code"`
	Detail         string      `json:"detail"`
	RecordID       string      `json:"record_id,omitempty"`
	ResourceRecord []DNSRecord `json:"resource_record,omitempty"`
}

// DNSRecord 结构
type DNSRecord struct {
	RecordID string `json:"record_id"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Value    string `json:"value"`
	TTL      string `json:"ttl"`
	Distance int    `json:"distance"`
}

func fetchDNSRecords(apiKey string, domain string) ([]DNSRecord, error) {
	url := fmt.Sprintf("https://www.namesilo.com/api/dnsListRecords?version=1&type=json&key=%s&domain=%s", apiKey, domain)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if apiResponse.Reply.Code != 300 {
		return nil, fmt.Errorf("API response error: %s", apiResponse.Reply.Detail)
	}

	return apiResponse.Reply.ResourceRecord, nil
}

func getDomainPrefix(domain string) string {
	parts := strings.Split(domain, ".")
	return strings.Join(parts[:len(parts)-2], ".")
}

func getIP(ipv6 bool) (string, error) {
	url := "https://4.ipw.cn"
	if ipv6 {
		url = "https://6.ipw.cn"
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	//Convert the body to type string
	var ip = string(body)

	return ip, nil
}

func addDNSRecord(apiKey string, domain string, rrtype string, rrhost string, rrvalue string, rrttl int16) (*APIResponse, error) {
	url := fmt.Sprintf("https://www.namesilo.com/api/dnsAddRecord?version=1&type=json&key=%s&domain=%s&rrtype=%s&rrhost=%s&rrvalue=%s&rrttl=%d", apiKey, domain, rrtype, rrhost, rrvalue, rrttl)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var addRecordResponse APIResponse
	err = json.Unmarshal(body, &addRecordResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if addRecordResponse.Reply.Code != 300 {
		return nil, fmt.Errorf("API response error: %s", addRecordResponse.Reply.Detail)
	}

	return &addRecordResponse, nil
}

func updateDNSRecord(apiKey string, domain string, rrid string, rrhost string, rrvalue string, rrttl int16) (*APIResponse, error) {
	url := fmt.Sprintf("https://www.namesilo.com/api/dnsUpdateRecord?version=1&type=json&key=%s&domain=%s&rrid=%s&rrhost=%s&rrvalue=%s&rrttl=%d", apiKey, domain, rrid, rrhost, rrvalue, rrttl)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var updateRecordResponse APIResponse
	err = json.Unmarshal(body, &updateRecordResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if updateRecordResponse.Reply.Code != 300 {
		return nil, fmt.Errorf("API response error: %s", updateRecordResponse.Reply.Detail)
	}

	return &updateRecordResponse, nil
}

func loggerInit(debug bool) {
	bytesWriter := &bytes.Buffer{}
	stdoutWriter := os.Stdout
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z",
		FullTimestamp:   true})
	log.SetOutput(io.MultiWriter(bytesWriter, stdoutWriter))
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	if !debug {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	h := flag.Bool("help", false, "--help")
	flagDomain := flag.String("domain", "", "A top-level domain.")
	flagType := flag.String("type", "A", "The type of a domain name can be classified as A,AAAA,TXT,or CNAME.")
	flagName := flag.String("name", "", "Second-level domain.")
	flagRecord := flag.String("record", "", "IP value, if left blank, will be automatically obtained from the internet.")
	flagKey := flag.String("key", "", "The key for Namesilo.")
	flasDebug := flag.Bool("debug", false, "Debug model.")
	// proxyUrl := flag.String("proxy", "", "Proxy for HTTP requests.")
	flag.CommandLine.SortFlags = false
	flag.Parse()

	if *h {
		flag.Usage()
		return
	}

	if *flagDomain == "" || *flagName == "" || *flagKey == "" {
		flag.Usage()
		log.Fatalln("Please verify if the parameters you have have entered are correct.")
	}

	loggerInit(*flasDebug)

	var recordValue string = *flagRecord
	if recordValue == "" {
		recordValue, _ = getIP(*flagType == "AAAA")
		if recordValue == "" {
			log.Fatalln("Unable to obtain IP value")
		}
	}
	log.WithField("IP", recordValue).Info("Successfully obtained IP.")

	records, err := fetchDNSRecords(*flagKey, *flagDomain)
	if err != nil {
		log.Fatalf("Error fetching DNS records: %v", err)
	}

	already_exists := false
	record_id := ""

	for _, record := range records {
		if getDomainPrefix(record.Host) == *flagName && record.Type == *flagType {
			already_exists = true
			record_id = record.RecordID
		}
		log.WithFields(log.Fields{"Record ID": record.RecordID, "Record Type": record.Type, "Record Host": record.Host, "Record Value": record.Value, "Record TTL": record.TTL}).Debug("Record value")
	}

	if already_exists {
		log.WithFields(log.Fields{"Record ID": record_id, "Domain Name": *flagName, "Record Value": recordValue}).Info("Update DNS Record")
		r, err := updateDNSRecord(*flagKey, *flagDomain, record_id, *flagName, recordValue, 7207)
		if err != nil {
			log.Fatalf("Error update DNS records: %v", err)
		}
		log.WithFields(log.Fields{
			"code":            r.Reply.Code,
			"detail":          r.Reply.Detail,
			"record_id":       r.Reply.RecordID,
			"resource_record": r.Reply.ResourceRecord,
		}).Info("Update DNS record completed")
	}
	if !already_exists {
		log.WithFields(log.Fields{"Record Type": *flagType, "Domain Name": *flagName, "Record Value": recordValue}).Info("Add DNS Record")
		r, err := addDNSRecord(*flagKey, *flagDomain, *flagType, *flagName, recordValue, 7207)
		if err != nil {
			log.Fatalf("Error update DNS records: %v", err)
		}
		log.WithFields(log.Fields{
			"code":            r.Reply.Code,
			"detail":          r.Reply.Detail,
			"record_id":       r.Reply.RecordID,
			"resource_record": r.Reply.ResourceRecord,
		}).Info("Update DNS record completed")
	}
}
