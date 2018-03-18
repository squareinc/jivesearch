package parcel

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestUPSFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    string
		resp string
	}{
		{
			name: "1Z12345E6605272234",
			u:    `https://onlinetools.ups.com/rest/Track`,
			resp: `{"TrackResponse":{"Response":{"ResponseStatus":{"Code":"1", "Description":"Success"}, "TransactionReference":{"CustomerContext":"Your Test Case Summary Description"}}, "Shipment":{"InquiryNumber":{"Code":"01", "Description":"ShipmentIdentificationNumber", "Value":"1Z12345E6605272234"}, "ShipperNumber":"", "ShipmentAddress":[{"Type":{"Code":"01", "Description":"Shipper Address"}, "Address":{"AddressLine":["1 Main", "Unit 10"], "City":"Some City", "StateProvinceCode":"ID", "PostalCode":"90210", "CountryCode":"US"}}, {"Type":{"Code":"02", "Description":"ShipTo Address"}, "Address":{"City":"Another City", "StateProvinceCode":"KS", "PostalCode":"90211", "CountryCode":"US"}}], "ShipmentWeight":{"UnitOfMeasurement":{"Code":"LBS"}, "Weight":"17.50"}, "Service":{"Code":"003", "Description":"UPS GROUND"}, "ReferenceNumber":[{"Code":"01", "Value":""}, {"Code":"01", "Value":""}], "DeliveryDetail":{"Type":{"Code":"03", "Description":"Scheduled Delivery"}, "Date":"20180311"}, "Package":{"TrackingNumber":"1Z12345E6605272234", "PackageServiceOption":{"Type":{"Code":"024", "Description":"ARS"}}, "Activity":[{"ActivityLocation":{"Address":{"City":"Banahana", "StateProvinceCode":"ID", "CountryCode":"US"}}, "Status":{"Type":"I", "Description":"Departure Scan", "Code":"DP"}, "Date":"20180311", "Time":"023800"}], "Message":{"Code":"01", "Description":"On Time"}, "PackageWeight":{"UnitOfMeasurement":{"Code":"LBS"}, "Weight":"17.50"}, "ReferenceNumber":[{"Code":"01", "Value":""}, {"Code":"01", "Value":""}]}}, "Disclaimer":"You are using UPS tracking service on..."}}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("POST", tt.u, responder)

			a := &UPS{
				HTTPClient: &http.Client{},
			}
			got, err := a.Fetch(tt.name)
			if err != nil {
				t.Fatal(err)
			}

			want := Response{
				TrackingNumber: strings.ToUpper(tt.name),
				Updates: []Update{
					{
						DateTime: time.Date(2018, 3, 11, 2, 38, 0, 0, time.UTC),
						Location: Location{
							City: "Banahana", State: "ID", Country: "US",
						},
						Status: "Departure Scan",
					},
				},
				Expected: Expected{
					Delivery: "Scheduled Delivery",
					Date:     time.Date(2018, 3, 11, 0, 0, 0, 0, time.UTC),
				},
				URL: fmt.Sprintf("https://wwwapps.ups.com/WebTracking/processInputRequest?AgreeToTermsAndConditions=yes&InquiryNumber1=%v&TypeOfInquiryNumber=T&error_carried=true&loc=en-us&sort_by=status&tracknums_displayed=1", strings.ToUpper(tt.name)),
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}

	httpmock.Reset()
}
