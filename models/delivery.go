package models

type DeliverableAddress struct {
	UserName     string `json:"userName,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
	ProvinceName string `json:"provinceName,omitempty"`
	CityName     string `json:"cityName,omitempty"`
	CountyName   string `json:"countyName,omitempty"`
	DetailInfo   string `json:"detailInfo,omitempty"`
	NationalCode string `json:"nationalCode,omitempty"`
	TelNumber    string `json:"telNumber,omitempty"`
}
