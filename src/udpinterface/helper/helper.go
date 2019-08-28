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

// Package helper implements rest helper functions
package helper

import (
	"log"
	"net"
)

const (
	StatusOK      = 0
	NOTFoundError = 404
)

// UDPHelper is the interface implemented by rest helper functions
type UDPHelper interface {
	requestHelper
	responseHelper
}

type requestHelper interface {
	SendRequest(endpoint string, data []byte) ([]byte, error)
}

type responseHelper interface {
	ResponseJSON(conn *net.UDPConn, bytes []byte, addr *net.UDPAddr)
}

type helperImpl struct{}

var helper helperImpl

// GetHelper returns helperImpl instance
func GetHelper() UDPHelper {
	return helper
}

func (helperImpl) ResponseJSON(conn *net.UDPConn, bytes []byte, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP(bytes, addr)
	if err != nil {
		log.Println("failed sending response, err : ", err)
	}
}

func (helperImpl) SendRequest(endpoint string, data []byte) ([]byte, error) {
	endpointResolvedAddr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		log.Println("Cant send UDP request", err.Error())
		return []byte{}, err
	}

	conn, err := net.DialUDP("udp", nil, endpointResolvedAddr)
	if err != nil {
		log.Println("Cant make UDP connection to", endpointResolvedAddr.String(), err.Error())
		return []byte{}, err
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Println("Cant write to", endpointResolvedAddr.String(), err.Error())
		return []byte{}, err
	}

	buffer := make([]byte, 1024)

	_, err = conn.Read(buffer)
	if err != nil {
		log.Println("Error reading from", endpointResolvedAddr.String(), err.Error())
		return []byte{}, err
	}

	return buffer, nil
}
