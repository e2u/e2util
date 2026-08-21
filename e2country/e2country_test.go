package e2country

import "testing"

func TestCountryCodes(t *testing.T) {
	if len(CountryCodes) == 0 {
		t.Fatal("CountryCodes is empty")
	}

	want := map[string]CountryCode{
		"US": {Country: "United States", CountryCode: "1", ISOCode2: "US", ISOCode3: "USA"},
		"CN": {Country: "China", CountryCode: "86", ISOCode2: "CN", ISOCode3: "CHN"},
		"JP": {Country: "Japan", CountryCode: "81", ISOCode2: "JP", ISOCode3: "JPN"},
		"GB": {Country: "United Kingdom", CountryCode: "44", ISOCode2: "GB", ISOCode3: "GBR"},
	}

	found := map[string]*CountryCode{}
	for _, c := range CountryCodes {
		if c == nil {
			t.Fatal("nil CountryCode entry")
		}
		if _, ok := want[c.ISOCode2]; ok {
			found[c.ISOCode2] = c
		}
	}
	for iso, exp := range want {
		got, ok := found[iso]
		if !ok {
			t.Errorf("missing ISO %s", iso)
			continue
		}
		if got.Country != exp.Country || got.CountryCode != exp.CountryCode || got.ISOCode3 != exp.ISOCode3 {
			t.Errorf("ISO %s: got %+v, want %+v", iso, *got, exp)
		}
	}
}
