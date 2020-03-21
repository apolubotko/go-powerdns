package powerdns

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/joeig/go-powerdns/v2/lib"
)

func randomString(length int) string {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}

	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for i, b := range bytes {
		character := b % byte(len(characters))
		bytes[i] = character
	}

	return characters
}

func generateTestRecord(client *Client, domain string, autoAddRecord bool) string {
	name := fmt.Sprintf("test-%s.%s", randomString(16), domain)

	if mock.Disabled() && autoAddRecord {
		if err := client.Records.Add(domain, name, lib.RRTypeTXT, 300, []string{"\"Testing...\""}); err != nil {
			fmt.Printf("Error creating record: %s\n", name)
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("Created record %s\n", name)
		}
	}

	return name
}

func TestAddRecord(t *testing.T) {
	testDomain := generateTestZone(true)
	p := initialisePowerDNSTestClient(&mock)

	mock.RegisterRecordMockResponder(testDomain)

	testRecordNameTXT := generateTestRecord(p, testDomain, false)
	if err := p.Records.Add(testDomain, testRecordNameTXT, lib.RRTypeTXT, 300, []string{"\"bar\""}); err != nil {
		t.Errorf("%s", err)
	}

	testRecordNameCNAME := generateTestRecord(p, testDomain, false)
	if err := p.Records.Add(testDomain, testRecordNameCNAME, lib.RRTypeCNAME, 300, []string{"foo.tld"}); err != nil {
		t.Errorf("%s", err)
	}
}

func TestAddRecordError(t *testing.T) {
	p := initialisePowerDNSTestClient(&mock)
	p.Port = "x"
	testDomain := generateTestZone(false)

	testRecordName := generateTestRecord(p, testDomain, false)
	if err := p.Records.Add(testDomain, testRecordName, lib.RRTypeTXT, 300, []string{"\"bar\""}); err == nil {
		t.Error("error is nil")
	}
}

func TestChangeRecord(t *testing.T) {
	testDomain := generateTestZone(true)

	p := initialisePowerDNSTestClient(&mock)
	mock.RegisterRecordMockResponder(testDomain)

	testRecordName := generateTestRecord(p, testDomain, true)
	if err := p.Records.Change(testDomain, testRecordName, lib.RRTypeTXT, 300, []string{"\"bar\""}); err != nil {
		t.Errorf("%s", err)
	}
}

func TestChangeRecordError(t *testing.T) {
	p := initialisePowerDNSTestClient(&mock)
	p.Port = "x"
	testDomain := generateTestZone(false)

	testRecordName := generateTestRecord(p, testDomain, false)
	if err := p.Records.Change(testDomain, testRecordName, lib.RRTypeTXT, 300, []string{"\"bar\""}); err == nil {
		t.Error("error is nil")
	}
}

func TestDeleteRecord(t *testing.T) {
	testDomain := generateTestZone(true)
	p := initialisePowerDNSTestClient(&mock)
	mock.RegisterRecordMockResponder(testDomain)

	testRecordName := generateTestRecord(p, testDomain, true)
	if err := p.Records.Delete(testDomain, testRecordName, lib.RRTypeTXT); err != nil {
		t.Errorf("%s", err)
	}
}

func TestDeleteRecordError(t *testing.T) {
	p := initialisePowerDNSTestClient(&mock)
	p.Port = "x"
	testDomain := generateTestZone(false)

	testRecordName := generateTestRecord(p, testDomain, false)
	if err := p.Records.Delete(testDomain, testRecordName, lib.RRTypeTXT); err == nil {
		t.Error("error is nil")
	}
}

func TestCanonicalResourceRecordValues(t *testing.T) {
	testCases := []struct {
		records     []lib.Record
		wantContent []string
	}{
		{[]lib.Record{{Content: lib.String("foo.tld")}}, []string{"foo.tld."}},
		{[]lib.Record{{Content: lib.String("foo.tld.")}}, []string{"foo.tld."}},
		{[]lib.Record{{Content: lib.String("foo.tld")}, {Content: lib.String("foo.tld.")}}, []string{"foo.tld.", "foo.tld."}},
	}

	for i, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("TestCase%d", i), func(t *testing.T) {
			canonicalResourceRecordValues(tc.records)

			for j := range tc.records {
				isContent := *tc.records[j].Content
				wantContent := tc.wantContent[j]

				if isContent != wantContent {
					t.Errorf("Comparison failed: %s != %s", isContent, wantContent)
				}
			}
		})
	}
}

func TestFixRRset(t *testing.T) {
	testCases := []struct {
		rrset                     lib.RRset
		wantFixedCanonicalRecords bool
	}{
		{lib.RRset{Type: lib.RRTypePtr(lib.RRTypeMX), Records: []lib.Record{{Content: lib.String("foo.tld")}}}, true},
		{lib.RRset{Type: lib.RRTypePtr(lib.RRTypeCNAME), Records: []lib.Record{{Content: lib.String("foo.tld")}}}, true},
		{lib.RRset{Type: lib.RRTypePtr(lib.RRTypeA), Records: []lib.Record{{Content: lib.String("foo.tld")}}}, false},
	}

	for i, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("TestCase%d", i), func(t *testing.T) {
			fixRRset(&tc.rrset)

			if tc.wantFixedCanonicalRecords {
				for j := range tc.rrset.Records {
					isContent := *tc.rrset.Records[j].Content
					wantContent := lib.MakeDomainCanonical(*tc.rrset.Records[j].Content)

					if isContent != wantContent {
						t.Errorf("Comparison failed: %s != %s", isContent, wantContent)
					}
				}
			} else {
				for j := range tc.rrset.Records {
					isContent := *tc.rrset.Records[j].Content
					wrongContent := lib.MakeDomainCanonical(*tc.rrset.Records[j].Content)

					if isContent == wrongContent {
						t.Errorf("Comparison failed: %s == %s", isContent, wrongContent)
					}
				}
			}
		})
	}
}
