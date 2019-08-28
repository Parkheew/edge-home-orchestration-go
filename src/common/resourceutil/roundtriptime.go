/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/

// Package resourceutil provides the information of resource usage of local device
package resourceutil

import (
	"encoding/json"
	"fmt"
	"time"

	"udpinterface/helper"

	netDB "db/bolt/network"
)

const (
	pingAPI            = "/api/v1/ping"
	internalPort       = 56001
	defaultRttDuration = 5
)

var (
	udphelper     helper.UDPHelper
	netDBExecutor netDB.DBInterface
)

type packetBody struct {
	Action string
	Data   []byte
}

func init() {
	udphelper = helper.GetHelper()
	netDBExecutor = netDB.Query{}
}

func processRTT() {
	go func() {
		for {
			netInfos, err := netDBExecutor.GetList()
			if err != nil {
				return
			}

			for _, netInfo := range netInfos {
				totalCount := len(netInfo.IPv4)
				ch := make(chan float64, totalCount)
				for _, ip := range netInfo.IPv4 {
					go func(targetIP string) {
						ch <- checkRTT(targetIP)
					}(ip)
				}
				go func(info netDB.NetworkInfo) {
					result := selectMinRTT(ch, totalCount)
					info.RTT = result
					netDBExecutor.Update(info)
				}(netInfo)
			}
			time.Sleep(time.Duration(defaultRttDuration) * time.Second)
		}
	}()
}

func checkRTT(ip string) (rtt float64) {
	packet := packetBody{
		Action: "APIV1Ping",
		Data:   []byte{},
	}

	sendByte, err := json.Marshal(&packet)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	reqTime := time.Now()
	_, err = udphelper.SendRequest(fmt.Sprintf("%s:%d", ip, internalPort), sendByte)

	//targetURL := udphelper.MakeTargetURL(ip, internalPort, pingAPI)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	return time.Now().Sub(reqTime).Seconds()
}

func selectMinRTT(ch chan float64, totalCount int) (minRTT float64) {
	for i := 0; i < totalCount; i++ {
		select {
		case rtt := <-ch:
			if (rtt != 0 && rtt < minRTT) || minRTT == 0 {
				minRTT = rtt
			}
		}
	}
	return
}
