package sheets

import "testing"

func TestParseCSV(t *testing.T) {
	csv := "domain,classification,rationale,last_review,status,source_ref,tags\n" +
		"example.com,NO_DPA,Reason text,2025-08-01,active,TCK-1,\"EDTECH,PRIVACY\"\n" +
		"*.sample.org,EXPIRED_DPA,Expired agreement,2025-07-10,active,TCK-2,SECURITY\n"
	p, err := parseCSV(csv)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(p.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(p.Records))
	}
	if p.Records[0].Domain != "example.com" {
		t.Fatalf("domain mismatch")
	}
	if len(p.Records[0].Tags) != 2 {
		t.Fatalf("expected 2 tags (EDTECH, PRIVACY) got %d", len(p.Records[0].Tags))
	}
}

func TestParseCSV_MissingColumn(t *testing.T) {
	csv := "domain,classification,rationale,last_review\nexample.com,NO_DPA,Reason,2025-08-01\n"
	_, err := parseCSV(csv)
	if err == nil {
		t.Fatalf("expected error for missing status column")
	}
}

func TestParseCSV_InvalidClassification(t *testing.T) {
	csv := "domain,classification,rationale,last_review,status\nexample.com,BAD,Reason,2025-08-01,active\n"
	_, err := parseCSV(csv)
	if err == nil {
		t.Fatalf("expected invalid classification error")
	}
}
