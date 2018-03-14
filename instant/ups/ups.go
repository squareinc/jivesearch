// Package ups fetches ups tracking data
package ups

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Fetcher retrieves package info from the UPS API
type Fetcher interface {
	Fetch(number string) (Response, error)
}

// API holds settings for the UPS API
type API struct {
	HTTPClient *http.Client
	User       string
	Password   string
	Key        string
}

// Response is UPS's raw json response
type Response struct {
	TrackResponse struct {
		Response struct {
			ResponseStatus struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"ResponseStatus"`
			TransactionReference struct {
				CustomerContext string `json:"CustomerContext"`
			} `json:"TransactionReference"`
		} `json:"Response"`
		Shipment struct {
			InquiryNumber struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
				Value       string `json:"Value"`
			} `json:"InquiryNumber"`
			ShipperNumber   string `json:"ShipperNumber"`
			ShipmentAddress []struct {
				Type struct {
					Code        string `json:"Code"`
					Description string `json:"Description"`
				} `json:"Type"`
				Address struct {
					AddressLine       []string `json:"AddressLine"`
					City              string   `json:"City"`
					StateProvinceCode string   `json:"StateProvinceCode"`
					PostalCode        string   `json:"PostalCode"`
					CountryCode       string   `json:"CountryCode"`
				} `json:"Address"`
			} `json:"ShipmentAddress"`
			ShipmentWeight struct {
				UnitOfMeasurement struct {
					Code string `json:"Code"`
				} `json:"UnitOfMeasurement"`
				Weight string `json:"Weight"`
			} `json:"ShipmentWeight"`
			Service struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"Service"`
			ReferenceNumber []struct {
				Code  string `json:"Code"`
				Value string `json:"Value"`
			} `json:"ReferenceNumber"`
			DeliveryDetail struct {
				Type struct {
					Code        string `json:"Code"`
					Description string `json:"Description"`
				} `json:"Type"`
				Date string `json:"Date"`
			} `json:"DeliveryDetail"`
			Package struct {
				TrackingNumber       string `json:"TrackingNumber"`
				PackageServiceOption struct {
					Type struct {
						Code        string `json:"Code"`
						Description string `json:"Description"`
					} `json:"Type"`
				} `json:"PackageServiceOption"`
				Activity []struct {
					ActivityLocation struct {
						Address struct {
							City              string `json:"City"`
							StateProvinceCode string `json:"StateProvinceCode"`
							CountryCode       string `json:"CountryCode"`
						} `json:"Address"`
					} `json:"ActivityLocation"`
					Status struct {
						Type        string `json:"Type"`
						Description string `json:"Description"`
						Code        string `json:"Code"`
					} `json:"Status"`
					Date string `json:"Date"`
					Time string `json:"Time"`
				} `json:"Activity"`
				Message struct {
					Code        string `json:"Code"`
					Description string `json:"Description"`
				} `json:"Message"`
				PackageWeight struct {
					UnitOfMeasurement struct {
						Code string `json:"Code"`
					} `json:"UnitOfMeasurement"`
					Weight string `json:"Weight"`
				} `json:"PackageWeight"`
				ReferenceNumber []struct {
					Code  string `json:"Code"`
					Value string `json:"Value"`
				} `json:"ReferenceNumber"`
			} `json:"Package"`
		} `json:"Shipment"`
		Disclaimer string `json:"Disclaimer"`
	} `json:"TrackResponse"`
}

func (a *API) buildRequest(number string) (*http.Request, error) {
	// test url := "https://wwwcie.ups.com/rest/Track"
	url := "https://onlinetools.ups.com/rest/Track"

	jsonStr := []byte(fmt.Sprintf(`{
		"UPSSecurity": {
			"UsernameToken": {
				"Username": "%v",
				"Password": "%v"
			},
			"ServiceAccessToken": {
				"AccessLicenseNumber": "%v"
			}
		},
		"TrackRequest": {
			"Request": {
				"RequestOption": "1",
				"TransactionReference": {
					"CustomerContext": "Your Test Case Summary Description"
				}
			},
			"InquiryNumber": "%v"
		}
	}`, a.User, a.Password, a.Key, number))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	// Docs say the following headers are required but API seems to work w/out them
	req.Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	req.Header.Set("Access-Control-Allow-Methods", "POST")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Content-Type", "application/json")
	return req, err
}

// Fetch retrieves a UPS API response
func (a *API) Fetch(trackingNumber string) (Response, error) {
	r := Response{}

	req, err := a.buildRequest(trackingNumber)
	if err != nil {
		return r, err
	}

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}
