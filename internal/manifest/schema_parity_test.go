package manifest

import (
	"os"
	"testing"

	"github.com/nao1215/atago/internal/schematest"
)

// TestManifestSchema_StructParity is the drift guard the manifest schema lacked
// (#496): its example-conformance check could only catch a missing property
// once the committed example happened to carry one, so project_path,
// fixtures_dir, expect_fail, and deterministic all shipped unschemad.
func TestManifestSchema_StructParity(t *testing.T) {
	t.Parallel()
	const schemaPath = "../../schema/manifest.schema.json"
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read %s: %v", schemaPath, err)
	}
	schematest.CheckParity(t, schemaPath, data, Document{})
}
