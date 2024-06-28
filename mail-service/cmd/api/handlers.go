package main

import (
	"log"
	"net/http"

	"github.com/abhilashdk2016/toolkit/v2"
)

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	log.Println("Inside mail-service -> handlers.go -> SendMail")
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	var tools toolkit.Tools
	err := tools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		log.Println("Inside mail-service -> handlers.go -> SendMail -> ReadJSON error " + err.Error())
		tools.ErrorJSON(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		log.Println("Inside mail-service -> handlers.go -> SendMail -> SendSMTPMessage error " + err.Error())
		tools.ErrorJSON(w, err)
		return
	}

	payload := toolkit.JSONResponse{
		Error:   false,
		Message: "sen to " + requestPayload.To,
	}

	tools.WriteJSON(w, http.StatusAccepted, payload)
}
