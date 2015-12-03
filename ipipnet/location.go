package ipipnet

import (
	"fmt"
	"strings"
)

type Place struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type ISP struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type Location struct {
	Country  *Place `json:"country"`
	Province *Place `json:"province"`
	City     *Place `json:"city"`
	ISPs     []*ISP `json:"isps"`
}

func (location Location) String() string {
	items := make([]string, 0, 4)
	if location.Country != nil {
		items = append(items, fmt.Sprintf("%s(%d)", location.Country.Name, location.Country.ID))
	}
	if location.Province != nil {
		items = append(items, fmt.Sprintf("%s(%d)", location.Province.Name, location.Province.ID))
	}
	if location.City != nil {
		items = append(items, fmt.Sprintf("%s(%d)", location.City.Name, location.City.ID))
	}
	for _, isp := range location.ISPs {
		items = append(items, fmt.Sprintf("%s(%d)", isp.Name, isp.ID))
	}
	return strings.Join(items, "/")
}

func (location Location) GetCountryID() uint64 {
	if location.Country != nil {
		return location.Country.ID
	} else {
		return 0
	}
}

func (location Location) GetCountryName() string {
	if location.Country != nil {
		return location.Country.Name
	} else {
		return ""
	}
}

func (location Location) GetProvinceID() uint64 {
	if location.Province != nil {
		return location.Province.ID
	} else {
		return 0
	}
}

func (location Location) GetProvinceName() string {
	if location.Province != nil {
		return location.Province.Name
	} else {
		return ""
	}
}

func (location Location) GetCityID() uint64 {
	if location.City != nil {
		return location.City.ID
	} else {
		return 0
	}
}

func (location Location) GetCityName() string {
	if location.City != nil {
		return location.City.Name
	} else {
		return ""
	}
}

func (location Location) GetISPIDs() []uint64 {
	ids := make([]uint64, len(location.ISPs))
	for i, isp := range location.ISPs {
		ids[i] = isp.ID
	}
	return ids
}

func (location Location) GetISPNames() []string {
	names := make([]string, len(location.ISPs))
	for i, isp := range location.ISPs {
		names[i] = isp.Name
	}
	return names
}
