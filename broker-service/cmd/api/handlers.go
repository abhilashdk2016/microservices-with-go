package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"

	"github.com/abhilashdk2016/my-broker/event"
	"github.com/abhilashdk2016/toolkit/v2"
)

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	var tools toolkit.Tools
	payload := toolkit.JSONResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = tools.WriteJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var tools toolkit.Tools
	var requestPayload RequestPayload
	err := tools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItemViaRpc(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		tools.ErrorJSON(w, errors.New("unknown action"))

	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	var tools toolkit.Tools
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Println("About to call auth service")
	client := &http.Client{}
	response, err := client.Do(request)
	log.Println(response)
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		tools.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	} else if response.StatusCode != http.StatusAccepted {
		tools.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromAuthService toolkit.JSONResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromAuthService)
	if err != nil {
		tools.ErrorJSON(w, err)
		return
	}

	if jsonFromAuthService.Error {
		tools.ErrorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload toolkit.JSONResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromAuthService.Data

	tools.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logWriter(w http.ResponseWriter, l LogPayload) {
	var tools toolkit.Tools
	jsonData, _ := json.MarshalIndent(l, "", "\t")
	logServiceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Println("About to call log service")
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	log.Println(response)
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	var payload toolkit.JSONResponse
	payload.Error = false
	payload.Message = "logged"

	tools.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	var tools toolkit.Tools
	jsonData, _ := json.MarshalIndent(msg, "", "\t")
	mailServiceURL := "http://mail-service/send"
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error converting msg to json ", err.Error())
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Println("About to call mail service")
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	log.Println(response)
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	var payload toolkit.JSONResponse
	payload.Error = false
	payload.Message = "mail sent"

	tools.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbitMQ(w http.ResponseWriter, l LogPayload) {
	var tools toolkit.Tools
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
	}

	var payload toolkit.JSONResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ Message"

	tools.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, message string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

type RPCPaylod struct {
	Name string
	Data string
}

func (app *Config) logItemViaRpc(w http.ResponseWriter, l LogPayload) {
	var tools toolkit.Tools
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		tools.ErrorJSON(w, err)
	}

	rpcPayload := RPCPaylod{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		tools.ErrorJSON(w, err)
	}

	payload := toolkit.JSONResponse{
		Error:   false,
		Message: result,
	}

	tools.WriteJSON(w, http.StatusAccepted, payload)
}
