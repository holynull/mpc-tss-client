package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/holynull/go-log"
)

var Logger = log.Logger("main")

func main() {
	var userCount int
	flag.IntVar(&userCount, "n", 1, "Number of concurrent users. Default 1")
	var domain = "localhost"
	var port = 17788
	flag.StringVar(&domain, "h", "localhost", "Domain name or IP address of the server. Default localhost")
	flag.IntVar(&port, "port", 17788, "The port of server. Default 17788")
	var protocol string
	flag.StringVar(&protocol, "p", "http", "http or https, default http")
	var name string
	flag.StringVar(&name, "name", "test_concurrency", "test_concurrency,signing,resharing. defaul test_concurrency")
	flag.Parse()
	switch name {
	case "test_concurrency":
		keyChList := make([]chan []byte, 0)
		for i := 0; i < userCount; i++ {
			ch := make(chan []byte)
			n, err := rand.Int(rand.Reader, big.NewInt(2^256))
			if err != nil {
				Logger.Error(err)
				return
			}
			userId := hex.EncodeToString(n.Bytes())
			if sk, err := rsa.GenerateKey(rand.Reader, 4096); err != nil {
				return
			} else {
				pk_hex := hex.EncodeToString(x509.MarshalPKCS1PublicKey(&sk.PublicKey))
				go runRegisterAndKeygen(protocol, domain, port, userId, pk_hex, &ch)
				keyChList = append(keyChList[:], ch)
			}

		}
		counter := 0
		userIds := make([]string, 0)
		keygenDone := make(chan []byte)
		for _, ch := range keyChList {
			go func(ch chan []byte) {
				data := <-ch
				userId := string(data)
				userIds = append(userIds[:], userId)
				counter++
				if counter == len(keyChList) {
					keygenDone <- []byte("Keygen done")
				}
			}(ch)
		}
		<-keygenDone
		Logger.Infof("Keygen done, User's number: %d", len(userIds))
		signChList := make([]chan []byte, 0)
		for _, userId := range userIds {
			ch := make(chan []byte)
			msg := hex.EncodeToString([]byte("llllllllllllllllll"))
			go runSigning(protocol, domain, port, userId, msg, &ch)
			signChList = append(signChList[:], ch)
		}
		signDone := make(chan []byte)
		counter = 0
		userIds = make([]string, 0)
		for _, ch := range signChList {
			go func(ch chan []byte) {
				data := <-ch
				userId := string(data)
				userIds = append(userIds[:], userId)
				counter++
				if counter == len(signChList) {
					signDone <- []byte("Sign done")
				}
			}(ch)
		}
		<-signDone
		Logger.Infof("Sign done, User's number: %d", len(userIds))
	case "signing":
	case "resharing":
	default:
		flag.Usage()
	}

}

func runRegisterAndKeygen(protocol string, domain string, port int, userId string, rsaPk string, rChan *chan []byte) {
	url := fmt.Sprintf("%s://%s:%d/registerAndKeygen?userId=%s&rsaPk=%s", protocol, domain, port, userId, rsaPk)
	client := http.Client{Timeout: 10 * time.Second}
	if res, err := client.Get(url); err != nil {
		Logger.Error(err)
	} else {
		defer res.Body.Close()
		if body, err := ioutil.ReadAll(res.Body); err != nil {
			Logger.Error(err)
		} else {
			Logger.Infof("Client get response body len: %d", len(body))
			client.CloseIdleConnections()
			*rChan <- []byte(userId)
		}
	}
}

func runSigning(protocol string, domain string, port int, userId string, msg string, rChan *chan []byte) {
	url := fmt.Sprintf("%s://%s:%d/startSigning?userId=%s&msg=%s", protocol, domain, port, userId, msg)
	client := http.Client{Timeout: 10 * time.Second}
	if res, err := client.Get(url); err != nil {
		Logger.Error(err)
	} else {
		defer res.Body.Close()
		if body, err := ioutil.ReadAll(res.Body); err != nil {
			Logger.Error(err)
		} else {
			Logger.Infof("Client get response body len: %d", len(body))
			Logger.Infof("Client get response body : %s", string(body))
			client.CloseIdleConnections()
			*rChan <- []byte(userId)
		}
	}
}

func runResharing(protocol string, domain string, port int, userId string, rChan *chan []byte) {
	url := fmt.Sprintf("%s://%s:%d/startResharing?userId=%s", protocol, domain, port, userId)
	client := http.Client{Timeout: 10 * time.Second}
	if res, err := client.Get(url); err != nil {
		Logger.Error(err)
	} else {
		defer res.Body.Close()
		if body, err := ioutil.ReadAll(res.Body); err != nil {
			Logger.Error(err)
		} else {
			Logger.Infof("Client get response body len: %d", len(body))
			client.CloseIdleConnections()
			*rChan <- []byte(userId)
		}
	}
}
