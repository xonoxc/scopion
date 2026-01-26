package model

type ErrorByService struct {
	Service string `json:"service"`
	Count   int    `json:"count"`
}
