package models

type Date struct {
	Date int `json:"$date"`
}

type OID struct {
	OID string `json:"$oid"`
}

type FileDB struct {
	ID          OID    `json:"_id"`
	Filename    string `json:"filename"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Status      bool   `json:"status"`
	Permissions string `json:"permissions"`
	Date        Date   `json:"date"`
}
