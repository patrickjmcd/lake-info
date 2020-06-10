package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gocolly/colly"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
)

//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func publishToMQTT(client MQTT.Client, name string, value string) error {
	prefix := os.Getenv("MQTT_PREFIX")
	topic := fmt.Sprintf("%s/%s", prefix, name)
	token := client.Publish(topic, 0, false, value)
	token.Wait()
	return nil
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // E.g. "ExcbsVQs"
	return str
}

func isUSACEDataRow(row []string) bool {
	if len(row) != 8 {
		return false
	}

	for i := 1; i < 8; i++ {
		_, err := strconv.ParseFloat(row[i], 32)

		if err != nil {
			return false
		}
	}
	return true
}

func getUSACEData(url string, level chan string, turbineRelease chan string, spillwayRelease chan string, totalRelease chan string) error {
	c := colly.NewCollector(
		colly.UserAgent("lake-data-collector"),
	)
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnHTML("pre", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Elevation") {
			lines := strings.Split(e.Text, "\n")

			var lastValidRow []string
			for _, l := range lines {
				splitBySpaces := strings.Fields(l)
				if isUSACEDataRow(splitBySpaces) {
					lastValidRow = splitBySpaces
				}

			}
			level <- lastValidRow[2]
			turbineRelease <- lastValidRow[5]
			spillwayRelease <- lastValidRow[6]
			totalRelease <- lastValidRow[7]
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.Visit(url)
	return nil
}

func updateMQTT(level string, turbineRelease string, spillwayRelease string, totalRelease string) error {

	server := os.Getenv("MQTT_SERVER")
	if server == "" {
		return errors.New("no MQTT_SERVER specified")
	}

	mqttPort := os.Getenv("MQTT_PORT")
	port := "1883"
	if mqttPort != "" {
		port = mqttPort
	}

	prefix := os.Getenv("MQTT_PREFIX")
	if prefix == "" {
		return errors.New("no MQTT_PREFIX specified")
	}

	mqttURI := fmt.Sprintf("tcp://%s:%s", server, port)
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	clientID := fmt.Sprintf("lake-svc-%s", randomString())

	fmt.Printf("Trying to connect to %s\n -- clientID=%s\n", mqttURI, clientID)

	opts := MQTT.NewClientOptions().AddBroker(mqttURI)
	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(f)

	mq := MQTT.NewClient(opts)
	if token := mq.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	publishToMQTT(mq, "level", level)
	publishToMQTT(mq, "turbinerelease", turbineRelease)
	publishToMQTT(mq, "spillwayrelease", spillwayRelease)
	publishToMQTT(mq, "totalrelease", totalRelease)
	mq.Disconnect(1000)
	return nil
}

func publishToInfluxdb(writeAPI api.WriteApiBlocking, prefix string, name string, value string) error {
	fullName := fmt.Sprintf("%s%s", prefix, name)
	units := "cfps"
	if name == "level" {
		units = "ft"
	}

	floatVal, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return err
	}

	p := influxdb2.NewPoint(fullName,
		map[string]string{"unit": units},
		map[string]interface{}{"value": value, "valueNum": floatVal},
		time.Now())
	err = writeAPI.WritePoint(context.Background(), p)
	return err
}

func updateInfluxdb(level string, turbineRelease string, spillwayRelease string, totalRelease string) error {
	server := os.Getenv("INFLUXDB_SERVER")
	if server == "" {
		return errors.New("no INFLUXDB_SERVER specified")
	}

	influxdbPort := os.Getenv("INFLUXDB_PORT")
	port := "8086"
	if influxdbPort != "" {
		port = influxdbPort
	}

	prefix := os.Getenv("INFLUXDB_PREFIX")
	if prefix == "" {
		return errors.New("no INFLUXDB_PREFIX specified")
	}

	influxDatabase := os.Getenv("INFLUXDB_DATABASE")
	database := "lakeinfo"
	if prefix == "" {
		database = influxDatabase
	}

	influxdbURI := fmt.Sprintf("http://%s:%s", server, port)
	client := influxdb2.NewClient(influxdbURI, "")
	writeAPI := client.WriteApiBlocking("", database)

	err := publishToInfluxdb(writeAPI, prefix, "level", level)
	err = publishToInfluxdb(writeAPI, prefix, "turbine_release", turbineRelease)
	err = publishToInfluxdb(writeAPI, prefix, "spillway_release", spillwayRelease)
	err = publishToInfluxdb(writeAPI, prefix, "total_release", totalRelease)

	return err
}

func getLatestValues(url string) string {
	usaceLevelCh := make(chan string, 1)
	usaceTurbineReleaseCh := make(chan string, 1)
	usaceSpillwayReleaseCh := make(chan string, 1)
	usaceTotalReleaseCh := make(chan string, 1)
	getUSACEData(url, usaceLevelCh, usaceTurbineReleaseCh, usaceSpillwayReleaseCh, usaceTotalReleaseCh)
	var l, turb, spill, tot string

	for i := 0; i < 4; i++ {
		select {
		case usaceLevel := <-usaceLevelCh:
			fmt.Printf("received USACE Level %s ft.\n", usaceLevel)
			l = usaceLevel
		case usaceTurbineRelease := <-usaceTurbineReleaseCh:
			fmt.Printf("received USACE Turbine Release: %s cfs\n", usaceTurbineRelease)
			turb = usaceTurbineRelease
		case usaceSpillwayRelease := <-usaceSpillwayReleaseCh:
			fmt.Printf("received USACE Spillway Release: %s cfs\n", usaceSpillwayRelease)
			spill = usaceSpillwayRelease
		case usaceTotalRelease := <-usaceTotalReleaseCh:
			fmt.Printf("received USACE Spillway Release: %s cfs\n", usaceTotalRelease)
			tot = usaceTotalRelease
		}
	}
	err := updateMQTT(l, turb, spill, tot)
	if err != nil {
		fmt.Printf("Couldn't send to MQTT: %s\n", err)
	} else {
		fmt.Println("Successfully wrote to MQTT")
	}
	err = updateInfluxdb(l, turb, spill, tot)
	if err != nil {
		fmt.Printf("Couldn't send to InfluxDB: %s\n", err)
	} else {
		fmt.Println("Successfully wrote to InfluxDB")
	}
	return fmt.Sprintf("Level: %s ft\nTurbine Release: %s cfs\nSpillway Release: %s cfs\nTotal Release: %s cfs\n", l, turb, spill, tot)
}

func main() {
	url := os.Getenv("USACE_URL")
	if url == "" {
		log.Fatal("did not provide USACE_URL")
	}
	getLatestValues(url)
}

// Handle a serverless request
func Handle(req []byte) string {
	url := os.Getenv("USACE_URL")
	if url != "" {
		log.Fatal("did not provide USACE_URL")
	}
	return getLatestValues(url)
}
