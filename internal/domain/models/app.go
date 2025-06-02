package models

type App struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}
