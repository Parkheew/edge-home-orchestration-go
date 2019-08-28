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

// Package externalhandler implements REST server functions to communication between orchestration and service applications
package externalhandler

import (
	"log"
	"net"

	"orchestrationapi"
	"udpinterface/cipher"
	"udpinterface/helper"
)

const logPrefix = "ExternalInterface"

// Handler struct
type Handler struct {
	isSetAPI bool
	api      orchestrationapi.OrcheExternalAPI

	helper helper.UDPHelper

	cipher.HasCipher
}

var handler *Handler

func init() {
	handler = new(Handler)
	handler.helper = helper.GetHelper()
}

// GetHandler returns the singleton Handler instance
func GetHandler() *Handler {
	return handler
}

// SetOrchestrationAPI sets OrcheExternalAPI
func (h *Handler) SetOrchestrationAPI(o orchestrationapi.OrcheExternalAPI) {
	h.api = o
	h.isSetAPI = true
}

// IsSetAPIInstance returns isSetAPI
func (h *Handler) IsSetAPIInstance() bool {
	return h.isSetAPI
}

// APIV1RequestServicePost handles service request from service application
func (h *Handler) APIV1RequestServicePost(conn *net.UDPConn, clientAddr *net.UDPAddr, body []byte) {
	log.Printf("[%s] APIV1RequestServicePost", logPrefix)
	if h.isSetAPI == false {
		log.Printf("[%s] does not set api", logPrefix)
		h.helper.ResponseJSON(conn, h.makeErrorBody(helper.NOTFoundError), clientAddr)
		return
	} else if h.IsSetKey == false {
		log.Printf("[%s] does not set key", logPrefix)
		h.helper.ResponseJSON(conn, h.makeErrorBody(helper.NOTFoundError), clientAddr)
		return
	}

	var (
		responseMsg  string
		responseName string
		resp         orchestrationapi.ResponseService

		executeEnvs        []interface{}
		responseTargetInfo map[string]interface{}
	)

	//request
	appCommand, err := h.Key.DecryptByteToJSON(body)
	if err != nil {
		log.Printf("[%s] can not decryption", logPrefix)
		// h.helper.Response(w, http.StatusServiceUnavailable)
		return
	}

	serviceInfos := orchestrationapi.ReqeustService{}
	name, ok := appCommand["ServiceName"].(string)
	if !ok {
		responseMsg = orchestrationapi.INVALID_PARAMETER
		responseName = ""
		goto SEND_RESP
	}
	serviceInfos.ServiceName = name

	executeEnvs, ok = appCommand["ServiceInfo"].([]interface{})
	if !ok {
		responseMsg = orchestrationapi.INVALID_PARAMETER
		responseName = name
		goto SEND_RESP
	}

	serviceInfos.ServiceInfo = make([]orchestrationapi.RequestServiceInfo, len(executeEnvs))
	for idx, executeEnv := range executeEnvs {
		tmp := executeEnv.(map[string]interface{})
		exeType, ok := tmp["ExecutionType"].(string)
		if !ok {
			responseMsg = orchestrationapi.INVALID_PARAMETER
			responseName = name
			goto SEND_RESP
		}
		serviceInfos.ServiceInfo[idx].ExecutionType = exeType

		exeCmd, ok := tmp["ExecCmd"].([]interface{})
		if !ok {
			responseMsg = orchestrationapi.INVALID_PARAMETER
			responseName = name
			goto SEND_RESP
		}

		serviceInfos.ServiceInfo[idx].ExeCmd = make([]string, len(exeCmd))
		for idy, cmd := range exeCmd {
			serviceInfos.ServiceInfo[idx].ExeCmd[idy] = cmd.(string)
		}
	}

	resp = h.api.RequestService(serviceInfos)

	responseMsg = resp.Message
	responseName = resp.ServiceName

	responseTargetInfo = make(map[string]interface{})
	responseTargetInfo["ExecutionType"] = resp.RemoteTargetInfo.ExecutionType
	responseTargetInfo["Target"] = resp.RemoteTargetInfo.Target

SEND_RESP:
	respJSONMsg := make(map[string]interface{})
	respJSONMsg["Message"] = responseMsg
	respJSONMsg["ServiceName"] = responseName
	respJSONMsg["RemoteTargetInfo"] = responseTargetInfo

	respEncryptBytes, err := h.Key.EncryptJSONToByte(respJSONMsg)
	if err != nil {
		log.Printf("[%s] can not encryption", logPrefix)
		h.helper.ResponseJSON(conn, h.makeErrorBody(helper.NOTFoundError), clientAddr)
		return
	}

	h.helper.ResponseJSON(conn, respEncryptBytes, clientAddr)
}

func (h *Handler) setHelper(helper helper.UDPHelper) {
	h.helper = helper
}

func (h *Handler) makeErrorBody(code int) []byte {
	respJSONMsg := make(map[string]interface{})
	respJSONMsg["Code"] = code

	respEncryptBytes, err := h.Key.EncryptJSONToByte(respJSONMsg)
	if err != nil {
		log.Printf("[%s] can not encryption", logPrefix)
		return nil
	}

	return respEncryptBytes
}
