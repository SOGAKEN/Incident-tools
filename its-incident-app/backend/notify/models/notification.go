package models

type NotificationRequest struct {
	IncidentID uint `json:"incident_id"`
	
	Responder string `json:"responder"`
	Content   string `json:"content"`
	Title  string `json:"title"`
	Chanel string `json:"chanel"`
	Name   string `json:"name"`
}
