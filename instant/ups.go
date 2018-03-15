package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant/contributors"
	"github.com/jivesearch/jivesearch/instant/ups"
	"github.com/jivesearch/jivesearch/log"
	"golang.org/x/text/language"
)

// UPS is an instant answer
type UPS struct {
	ups.Fetcher
	Answer
}

// PackageResponse is this instant answer's response to a triggered query
type PackageResponse struct {
	TrackingNumber string   `json:"tracking_number"`
	Updates        []Update `json:"updates"`
	Expected
	URL string `json:"url"`
}

// Expected is the expected delivery date and time
type Expected struct {
	Delivery string    `json:"delivery"`
	Date     time.Time `json:"expected"`
}

// Update is a single delivery event for a package
type Update struct {
	DateTime time.Time `json:"date_time"`
	Location `json:"location"`
	Status   string `json:"status"`
}

// Location is a location of an Update
type Location struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

func (u *UPS) setQuery(r *http.Request, qv string) answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *UPS) setUserAgent(r *http.Request) answerer {
	return u
}

func (u *UPS) setLanguage(lang language.Tag) answerer {
	u.language = lang
	return u
}

func (u *UPS) setType() answerer {
	u.Type = "ups"
	return u
}

func (u *UPS) setContributors() answerer {
	u.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return u
}

func (u *UPS) setRegex() answerer {
	/*
		UPS Format:
		https://www.ups.com/ca/en/tracking/help/tracking/tnh.page
		1Z9999999999999999
		999999999999
		T9999999999
		999999999

		The InquiryNumber element is limited to a length of 9-34 digits as it can accept tracking numbers for:
			Small Package -- Either 18 digits (starting with 1Z) or 11 digits (starting with a single letter)
			Mail Innovations -- Typically 22 digits, starting with a 9
			Ground Freight -- Either 9 digits (all numbers) or 18 digits (starting with 1Z)
			Air Freight -- Either 10 digits (all numbers) or 18 digits (starting with 1Z)
			Ocean Freight -- Possibly 17 digits (per example in Dev Guide)
	*/

	// https://stackoverflow.com/questions/619977/regular-expression-patterns-for-tracking-numbers
	// There is also an undocumented checksum. Email from UPS would not provide me how to calculate it but a couple of hints:
	// https://github.com/jkeen/tracking_number_data/blob/70065359c64996e1537c46efc5d0638b24df105b/couriers/ups.json
	// (old) https://www.codeproject.com/Articles/21224/Calculating-the-UPS-Tracking-Number-Check-Digit
	// Probably only necessary if it triggers other IA's accidentally???

	u.regex = append(u.regex, regexp.MustCompile(`(?i)\b(?P<trigger>(1Z ?[0-9A-Z]{3} ?[0-9A-Z]{3} ?[0-9A-Z]{2} ?[0-9A-Z]{4} ?[0-9A-Z]{3} ?[0-9A-Z]|T\d{3} ?\d{4} ?\d{3}|\d{22}))\b`))
	return u
}

func (u *UPS) solve() answerer {
	tn := strings.ToUpper(u.triggerWord)

	raw, err := u.Fetcher.Fetch(tn)
	if err != nil {
		u.Err = err
		return u
	}

	p := PackageResponse{
		TrackingNumber: raw.TrackResponse.Shipment.Package.TrackingNumber,
	}

	for _, a := range raw.TrackResponse.Shipment.Package.Activity {
		d := a.Date + a.Time
		dt, err := time.Parse("20060102150405", d)
		if err != nil {
			log.Info.Println(err) // this isn't serious enough to warrant a return
		}

		up := Update{
			DateTime: dt,
			Status:   a.Status.Description,
		}

		up.Location.City = a.ActivityLocation.Address.City
		up.Location.State = a.ActivityLocation.Address.StateProvinceCode
		up.Location.Country = a.ActivityLocation.Address.CountryCode
		p.Updates = append(p.Updates, up)
	}

	/*
		01 - Delivery
		02-  Estimated Delivery
		03 - Scheduled Delivery
		Mail innovations use only 01 and 02
	*/
	p.Expected.Delivery = raw.TrackResponse.Shipment.DeliveryDetail.Type.Description
	if raw.TrackResponse.Shipment.DeliveryDetail.Date != "" {
		p.Expected.Date, err = time.Parse("20060102", raw.TrackResponse.Shipment.DeliveryDetail.Date)
		if err != nil {
			u.Err = err
			return u
		}
	}

	p.URL, err = p.buildURL()
	if err != nil {
		u.Err = err
		return u
	}

	u.Solution = p

	return u
}

func (p *PackageResponse) buildURL() (string, error) {
	u, err := url.Parse("https://wwwapps.ups.com/WebTracking/processInputRequest")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("sort_by", "status")
	q.Set("error_carried", "true")
	q.Set("tracknums_displayed", "1")
	q.Set("TypeOfInquiryNumber", "T")
	q.Set("loc", "en-us")
	q.Set("InquiryNumber1", p.TrackingNumber)
	q.Set("AgreeToTermsAndConditions", "yes")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (u *UPS) setCache() answerer {
	u.Cache = true
	return u
}

func (u *UPS) tests() []test {
	testNumbers := []string{
		// Test numbers from "Tracking Web Service Developer Guide.pdf"
		// These won't return anything in the actual API but good for testing triggers
		"1z12345e0205271688",
		"1Z12345E6605272234",
		"1Z12345E0305271640",
		"1Z12345E0393657226",
		"1Z12345E1305277940",
		"1Z12345E6205277936",
		"1Z12345E1505270452",
		//"990728071",  // not working
		//"3251026119", // not working
		"9102084383041101186729",
		//"cgish000116630", // not working
		"1Z648616E192760718",
		//"5548789114",        // not working
		//"ER751105042015062", // not working
		"1ZWX0692YP40636269",
	}

	tests := []test{}

	for _, n := range testNumbers {
		t := test{
			query: n,
			expected: []Data{
				{
					Type:         "ups",
					Triggered:    true,
					Contributors: contributors.Load([]string{"brentadamson"}),
					Solution: PackageResponse{
						TrackingNumber: strings.ToUpper(n),
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
						URL: fmt.Sprintf("https://wwwapps.ups.com/WebTracking/processInputRequest?AgreeToTermsAndConditions=yes&InquiryNumber1=%v&TypeOfInquiryNumber=T&error_carried=true&loc=en-us&sort_by=status&tracknums_displayed=1", strings.ToUpper(n)),
					},
					Cache: true,
				},
			},
		}

		tests = append(tests, t)
	}

	return tests
}
