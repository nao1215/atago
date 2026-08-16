package report

import (
	"os"
	"testing"

	"github.com/nao1215/atago/internal/schematest"
)

// TestReportSchema_StructParity is the drift guard the report schema lacked
// (#496). The schema's only check was that a committed example validates, and
// that example carries neither a load failure, nor a suite-setup failure, nor an
// expect_fail scenario — so three fields shipped in this writer while the schema
// kept rejecting any real report that used them.
func TestReportSchema_StructParity(t *testing.T) {
	t.Parallel()
	const schemaPath = "../../schema/report.schema.json"
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read %s: %v", schemaPath, err)
	}
	schematest.CheckParity(t, schemaPath, data, jsonDocument{})
}
