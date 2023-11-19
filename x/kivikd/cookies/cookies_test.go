package cookies

import "testing"

func TestDecodeCookie(t *testing.T) {
	type cookieTest struct {
		TestName string
		Input    string
		Name     string
		Created  int64
		Err      string
	}
	tests := []cookieTest{
		{TestName: "Valid", Input: "YWRtaW46NThDNTQzN0Y697rnaWCa_rarAm25wbOg3Gm3mqc", Name: "admin", Created: 1489322879},
		{TestName: "InvalidBasae64", Input: "bogus base64", Err: "illegal base64 data at input byte 5"},
		{TestName: "Extra colons", Input: "Zm9vOjEyMzQ1OmJhcjpiYXoK", Name: "foo", Created: 74565},
		{TestName: "InvalidTimestamp", Input: "Zm9vOmJhcjpiYXoK", Err: `invalid timestamp: strconv.ParseInt: parsing "bar": invalid syntax`},
	}
	for _, test := range tests {
		func(test cookieTest) {
			t.Run(test.TestName, func(t *testing.T) {
				name, created, err := DecodeCookie(test.Input)
				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}
				if name != test.Name {
					t.Errorf("Name: Expected '%s', got '%s'", test.Name, name)
				}
				if created != test.Created {
					t.Errorf("Created: Expected %d, got %d", test.Created, created)
				}
				if errMsg != test.Err {
					t.Errorf("Error: Expected '%s', got '%s'", test.Err, errMsg)
				}
			})
		}(test)
	}
}
