package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"gitlab.com/waringer/broadlink/broadlinkrm"
)

func main() {
	fmt.Println("Broadlink RM Toolset")

	cmdAuth := flag.Bool("auth", true, "authenticate agaist device")
	cmdConvertBroadlink := flag.String("convertbroadlink", "", "convert code provided in Broadlink format to Pronto format")
	cmdConvertPronto := flag.String("convertpronto", "", "convert code provided in Pronto format to Broadlink format")
	cmdDiscover := flag.Bool("discover", false, "search for devices")
	deviceIP := flag.String("ip", "", "ip of device")
	cmdLearn := flag.Bool("learn", false, "set device in learing mode and wait up to 30 seconds for new learned code")
	cmdGetLearned := flag.Bool("learned", false, "get the last learned code from device in Broadlink format")
	cmdSend := flag.String("send", "", "send code provided in Broadlink format")
	cmdSendPronto := flag.String("sendpronto", "", "send code provided in Pronto format")
	flag.Parse()

	broadlinkrm.DefaultTimeout = 5
	broadlinkrm.LogWarnings = false

	var (
		ip        net.IP
		dev       []broadlinkrm.Device
		irCommand []byte
		err       error
	)

	if len(*cmdSend) != 0 {
		irCommand, err = hex.DecodeString(strings.Replace(*cmdSend, " ", "", -1))

		if err != nil {
			log.Fatalln("Provided Broadlink IR code is invalid")
		}
	} else if len(*cmdSendPronto) != 0 {
		irCommand, err = hex.DecodeString(strings.Replace(*cmdSendPronto, " ", "", -1))

		if err != nil {
			log.Fatalln("Provided Pronto IR code is invalid")
		}

		irCommand = broadlinkrm.ConvertPronto2Broadlink(irCommand)
	}

	if len(*cmdConvertBroadlink) != 0 {
		broadlinkCode, errBroadlink := hex.DecodeString(strings.Replace(*cmdConvertBroadlink, " ", "", -1))

		if errBroadlink != nil {
			log.Fatalln("Provided Broadlink IR code is invalid")
		}

		fmt.Printf("Converted IR code in Pronto format: %v \n", regexp.MustCompile("(?m)(.{4})").ReplaceAllString(hex.EncodeToString(broadlinkrm.ConvertBroadlink2Pronto(broadlinkCode)), "$1 "))
	}

	if len(*cmdConvertPronto) != 0 {
		prontoCode, errPronto := hex.DecodeString(strings.Replace(*cmdConvertPronto, " ", "", -1))

		if errPronto != nil {
			log.Fatalln("Provided Pronto IR code is invalid")
		}

		fmt.Printf("Converted IR code in Broadlink format: %x \n", broadlinkrm.ConvertPronto2Broadlink(prontoCode))
	}

	if *deviceIP != "" {
		ip = net.ParseIP(*deviceIP)
	}

	if *cmdDiscover {
		if ip == nil {
			dev = broadlinkrm.Hello(5, nil)
		} else {
			dev = broadlinkrm.Hello(0, ip)
		}

		fmt.Printf("Found %v device(s)\n", len(dev))
		for id, device := range dev {
			fmt.Printf("[%02v] Device type: %X \n", id, device.DeviceType)
			fmt.Printf("[%02v] Device name: %v \n", id, device.DeviceName)
			fmt.Printf("[%02v] Device MAC: [% x] \n", id, device.DeviceMac)
			fmt.Printf("[%02v] Device IP: %v \n", id, device.DeviceAddr.IP)

			if *cmdAuth {
				broadlinkrm.Auth(&dev[id])
				fmt.Printf("[%02v] Device authenticated \n", id)
			}
		}
	}

	if *cmdLearn {
		for id, device := range dev {
			broadlinkrm.Command(3, nil, &device)
			fmt.Printf("[%02v] Wait for learned code", id)

			var learnedCode []byte
			startTime := time.Now().Add(30 * time.Second)
			for time.Now().Before(startTime) {
				learnedCode = broadlinkrm.Command(4, nil, &device)

				if len(learnedCode) != 0 {
					fmt.Printf("\n[%02v] Learned code: [%x] \n", id, learnedCode)
					break
				}
				fmt.Print(".")
				time.Sleep(1 * time.Second)
			}

			if learnedCode == nil {
				fmt.Printf("\n[%02v] No code learned! \n", id)
			}
		}
	} else if *cmdGetLearned {
		for id, device := range dev {
			learnedCode := broadlinkrm.Command(4, nil, &device)
			fmt.Printf("[%02v] Device last learned code: [%x] \n", id, learnedCode)
		}
	}

	if irCommand != nil {
		for id, device := range dev {
			response := broadlinkrm.Command(2, irCommand, &device)

			if response == nil {
				fmt.Printf("[%02v] code send failed!\n", id)
			} else {
				fmt.Printf("[%02v] code send \n", id)
			}
		}
	}
}
