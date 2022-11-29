package smaps

import "github.com/CPU-commits/Intranet_BNews/src/services"

type SingleNewsMap struct {
	News *services.NewsResponse `json:"news"`
}

type NewsMap struct {
	News  []services.NewsResponse `json:"news"`
	Total int                     `json:"total" example:"15"`
}
