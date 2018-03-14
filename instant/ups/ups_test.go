package ups

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestUPSAPIFetch(t *testing.T) {
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

			a := &API{
				HTTPClient: &http.Client{},
			}
			got, err := a.Fetch(tt.name)
			if err != nil {
				t.Fatal(err)
			}

			want := Response{}
			if err := json.Unmarshal([]byte(tt.resp), &want); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}

	httpmock.Reset()
}
