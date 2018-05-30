package extension

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Response struct {
	Meta struct {
		Code          string `json:"code"`
		ExecutionTime string `json:"execution_time"`
	} `json:"meta"`
	Data struct {
		AddressesFound string `json:"addresses_found"`
		Addresses      []struct {
			PostalCode string `json:"postal_code"`
			City       string `json:"city"`
			County     string `json:"county"`
			State      string `json:"state"`
			Timezone   struct {
				ID          string `json:"id"`
				Location    string `json:"location"`
				CountryCode string `json:"country_code"`
				CountryName string `json:"country_name"`
			} `json:"timezone"`
			Datetime struct {
				Date          string `json:"date"`
				DateTime      string `json:"date_time"`
				DateTimeTxt   string `json:"date_time_txt"`
				DateTimeYmd   string `json:"date_time_ymd"`
				Time          string `json:"time"`
				Month         string `json:"month"`
				Day           string `json:"day"`
				Year          string `json:"year"`
				OffsetSeconds string `json:"offset_seconds"`
				OffsetGmt     string `json:"offset_gmt"`
			} `json:"datetime"`
		} `json:"addresses"`
	} `json:"data"`
}

func GetTimezoneByAddress(addr string) string {
	url := "https://timezoneapi.io/api/address/?" + addr

	result := Response{}
	err := GetJson(url, &result)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if result.Meta.Code == "200" {
		return result.Data.Addresses[0].Timezone.ID
	} else {
		return ""
	}
}

func GetJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func NowOnAddress(addr string) time.Time {
	now := time.Now()
	timezone := GetTimezoneByAddress(addr)
	if timezone != "" {
		loc, _ := time.LoadLocation(timezone)
		now = now.In(loc)
	}
	return now
}
