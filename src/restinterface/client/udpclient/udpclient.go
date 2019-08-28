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

// Package udpclient implements client functions to send UDP reqeust to remote orchestration
package udpclient

import (
	"encoding/json"
	"errors"
	"log"

	"restinterface/cipher"
	"restinterface/client"
	"restinterface/resthelper"
)

type udpClientImpl struct {
	port int
	cipher.HasCipher
	helper resthelper.RestHelper
}

type packetBody struct {
	action string
	data   []byte
}

const (
	constWellknownPort = 56001
	logPrefix          = "[udpclient]"
)

var udpClient *udpClientImpl

func init() {
	udpClient = new(udpClientImpl)
	udpClient.port = constWellknownPort
	udpClient.helper = resthelper.GetHelper()
}

// GetUDPClient returns the singleton restClientImpl instance
func GetUDPClient() client.Clienter {
	return udpClient
}

// DoExecuteRemoteDevice sends request to remote orchestration (APIV1ServicemgrServicesPost) to execute service
func (c udpClientImpl) DoExecuteRemoteDevice(appInfo map[string]interface{}, target string) (err error) {
	if c.IsSetKey == false {
		return errors.New("[" + logPrefix + "] does not set key")
	}

	encryptBytes, err := c.Key.EncryptJSONToByte(appInfo)
	if err != nil {
		return errors.New("[" + logPrefix + "] can not encrypt " + err.Error())
	}

	packet := packetBody{
		action: "APIV1ServicemgrServicesPost",
		data:   encryptBytes,
	}
	sendByte, err := json.Marshal(&packet)
	if err != nil {
		return errors.New("[" + logPrefix + "] can not Marshal " + err.Error())
	}

	respBytes, err := c.helper.SendUDPRequest(target, sendByte)
	if err != nil {
		return errors.New("[" + logPrefix + "] request return error" + err.Error())
	}

	respMsg, err := c.Key.DecryptByteToJSON(respBytes)
	if err != nil {
		return errors.New("[" + logPrefix + "] can not decrytion " + err.Error())
	}

	log.Println("[JSON] : ", respMsg)

	str := respMsg["Status"].(string)
	if str == "Failed" {
		err = errors.New("failed")
		return err
	}
	return nil
}

// DoNotifyAppStatusRemoteDevice sends request to remote orchestration (APIV1ServicemgrServicesNotificationServiceIDPost) to notify status
func (c udpClientImpl) DoNotifyAppStatusRemoteDevice(statusNotificationInfo map[string]interface{}, appID uint64, target string) error {
	if c.IsSetKey == false {
		return errors.New("[" + logPrefix + "] does not set key")
	}
	encryptBytes, err := c.Key.EncryptJSONToByte(statusNotificationInfo)
	if err != nil {
		return errors.New("[" + logPrefix + "] can not encryption " + err.Error())
	}

	packet := packetBody{
		action: "APIV1ServicemgrServicesNotificationServiceIDPost",
		data:   encryptBytes,
	}
	sendByte, err := json.Marshal(&packet)
	if err != nil {
		return errors.New("[" + logPrefix + "] can not Marshal " + err.Error())
	}

	_, err = c.helper.SendUDPRequest(target, sendByte)
	if err != nil {
		return errors.New("[" + logPrefix + "] request return error" + err.Error())
	}
	return nil
}

// DoGetScoreRemoteDevice  sends request to remote orchestration (APIV1ScoringmgrScoreLibnameGet) to get score
func (c udpClientImpl) DoGetScoreRemoteDevice(devID string, endpoint string) (scoreValue float64, err error) {
	if c.IsSetKey == false {
		return 0, errors.New("[" + logPrefix + "] does not set key")
	}

	info := make(map[string]interface{})
	info["devID"] = devID
	encryptBytes, err := c.Key.EncryptJSONToByte(info)
	if err != nil {
		return scoreValue, errors.New("[" + logPrefix + "] can not encryption " + err.Error())
	}

	packet := packetBody{
		action: "APIV1ScoringmgrScoreLibnameGet",
		data:   encryptBytes,
	}

	sendByte, err := json.Marshal(&packet)
	if err != nil {
		return 0, errors.New("[" + logPrefix + "] can not Marshal " + err.Error())
	}

	respBytes, err := c.helper.SendUDPRequest(endpoint, sendByte)
	if err == nil {
		respMsg, err := c.Key.DecryptByteToJSON(respBytes)
		if err != nil {
			return scoreValue, errors.New("[" + logPrefix + "] can not decryption " + err.Error())
		}

		log.Println("[JSON] : ", respMsg)

		scoreValue = respMsg["ScoreValue"].(float64)
		if scoreValue == 0.0 {
			err = errors.New("failed")
		}
		return scoreValue, err
	}
	return 0.0, err
}
