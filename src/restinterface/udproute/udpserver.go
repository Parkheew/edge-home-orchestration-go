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

// Package udproute implements management functions of UDP Router
package udproute

import (
	"encoding/json"
	"log"
	"net"
)

const (
	// ConstWellknownPort is the common port for REST API
	ConstWellknownPort = 56001
)

//UDPRouter struct
type UDPRouter struct {
}

type packetBody struct {
	Action string
	data   interface{}
}

//NewUDPRouter constructs UDPRouter instance
func NewUDPRouter() *UDPRouter {
	udpRouter := new(UDPRouter)
	return udpRouter
}

//Start wraps startServer function
func (u UDPRouter) Start() {
	go u.startServer()
}

//StartServer starts a UDP server on given port
func (u UDPRouter) startServer() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: ConstWellknownPort, Zone: ""})
	if err != nil {
		log.Println("UDP server could not be started")
		return
	}
	log.Println("UDP server started on port", ConstWellknownPort)
	buf := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		log.Println("udp request from", clientAddr)

		if err != nil {
			log.Println("Read Error!", err.Error())
			continue
		}
		if n < 1 {
			log.Println("Read Length Error!")
			continue
		}

		go u.handleRequest(buf, n, conn, clientAddr)
	}
}

func (u UDPRouter) handleRequest(buffer []byte, n int, conn *net.UDPConn, clientAddr *net.UDPAddr) {
	var packet packetBody
	err := json.Unmarshal(buffer[:n], &packet)
	if err != nil {
		log.Println("Packet UnMarshal error")
		return
	}

	switch packet.Action {
	case "APIV1Ping":
		//handle ping from remote orchestration
		_, err := conn.WriteToUDP([]byte("ping received"), clientAddr)
		if err != nil {
			log.Println("cannot Respond to ping", err.Error())
		}
	case "APIV1DiscoveryFromRelay":
		//handle discovery message from relay
		_, err := conn.WriteToUDP([]byte("discovery msg received"), clientAddr)
		if err != nil {
			log.Println("cannot Respond to discovery msg", err.Error())
		}
	case "APIV1ServicemgrServicesPost":
		//handle service execution request from remote orchestration
		_, err := conn.WriteToUDP([]byte("service req received"), clientAddr)
		if err != nil {
			log.Println("cannot Respond to service req", err.Error())
		}
	case "APIV1ServicemgrServicesNotificationServiceIDPost":
		//handle service notification request from remote orchestration
		_, err := conn.WriteToUDP([]byte("service noti req received"), clientAddr)
		if err != nil {
			log.Println("cannot Respond to service notification req", err.Error())
		}
	case "APIV1ScoringmgrScoreLibnameGet":
		//handle scoring request from remote orchestration
		_, err := conn.WriteToUDP([]byte("get score req received"), clientAddr)
		if err != nil {
			log.Println("cannot Respond to score request", err.Error())
		}
	case "APIV1RequestServicePost":
		//handles service request from service application
		_, err := conn.WriteToUDP([]byte("app service req received"), clientAddr)
		if err != nil {
			log.Println("cannot Respond to request app service request", err.Error())
		}
	}
}
