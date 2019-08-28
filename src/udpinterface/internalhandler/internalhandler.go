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

// Package internalhandler implements REST server functions to communication between orchestrations
package internalhandler

import (
	"log"
	"net"
	"net/http"

	"common/types/servicemgrtypes"
	"orchestrationapi"
	"restinterface/cipher"
	"restinterface/resthelper"
)

const logPrefix = "RestInternalInterface"

// Handler struct
type Handler struct {
	isSetAPI bool
	api      orchestrationapi.OrcheInternalAPI

	helper resthelper.RestHelper

	cipher.HasCipher
}

var handler *Handler

func init() {
	handler = new(Handler)
	handler.helper = resthelper.GetHelper()
}

// GetHandler returns the singleton Handler instance
func GetHandler() *Handler {
	return handler
}

// SetOrchestrationAPI sets OrcheInternalAPI
func (h *Handler) SetOrchestrationAPI(o orchestrationapi.OrcheInternalAPI) {
	h.api = o
	h.isSetAPI = true
}

// IsSetAPIInstance returns isSetAPI
func (h *Handler) IsSetAPIInstance() bool {
	return h.isSetAPI
}

// APIV1Ping handles ping request from remote orchestration
func (h *Handler) APIV1Ping(conn *net.UDPConn) {
	log.Printf("[%s] APIV1Ping", logPrefix)
	h.helper.ResponseJSON(w, nil, http.StatusOK)
}

// APIV1ServicemgrServicesPost handles service execution request from remote orchestration
func (h *Handler) APIV1ServicemgrServicesPost(conn *net.UDPConn, clientIP []byte, body []byte) {
	log.Printf("[%s] APIV1ServicemgrServicesPost", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	appInfo, err := h.Key.DecryptByteToJSON(body)
	if err != nil {
		log.Printf("[%s] can not decryption", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	appInfo["NotificationTargetURL"] = string(clientIP)
	log.Println(appInfo)

	h.api.ExecuteAppOnLocal(appInfo)

	respJSONMsg := make(map[string]interface{})
	respJSONMsg["Status"] = servicemgrtypes.ConstServiceStatusStarted

	respEncryptBytes, err := h.Key.EncryptJSONToByte(respJSONMsg)
	if err != nil {
		log.Printf("[%s] can not encryption", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	h.helper.ResponseJSON(w, respEncryptBytes, http.StatusOK)
}

// APIV1ServicemgrServicesNotificationServiceIDPost handles service notification request from remote orchestration
func (h *Handler) APIV1ServicemgrServicesNotificationServiceIDPost(conn *net.UDPConn, body []byte) {
	log.Printf("[%s] APIV1ServicemgrServicesNotificationServiceIDPost", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	statusNotification, err := h.Key.DecryptByteToJSON(body)
	if err != nil {
		log.Printf("[%s] can not decryption", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	serviceID := statusNotification["ServiceID"].(float64)
	status := statusNotification["Status"].(string)

	err = h.api.HandleNotificationOnLocal(serviceID, status)
	if err != nil {
		h.helper.Response(w, http.StatusInternalServerError)
		return
	}

	handler.helper.Response(w, http.StatusOK)
}

// APIV1ScoringmgrScoreLibnameGet handles scoring request from remote orchestration
func (h *Handler) APIV1ScoringmgrScoreLibnameGet(conn *net.UDPConn, body []byte) {
	log.Printf("[%s] APIV1ScoringmgrScoreLibnameGet", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	Info, err := h.Key.DecryptByteToJSON(body)
	if err != nil {
		log.Printf("[%s] can not decryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	devID := Info["devID"]

	scoreValue, err := h.api.GetScore(devID.(string))
	if err != nil {
		log.Printf("[%s] GetScore fail : %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusInternalServerError)
		return
	}

	respJSONMsg := make(map[string]interface{})
	respJSONMsg["ScoreValue"] = scoreValue

	respEncryptBytes, err := h.Key.EncryptJSONToByte(respJSONMsg)
	if err != nil {
		log.Printf("[%s] can not encryption %s", logPrefix, err.Error())
		h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	h.helper.ResponseJSON(w, respEncryptBytes, http.StatusOK)
}

func (h *Handler) setHelper(helper resthelper.RestHelper) {
	h.helper = helper
}
