package loader

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// validateMatrix checks the shape of every scenario's matrix block before it is
// expanded (spec.md §22). It runs on the raw decoded spec, so the rows are still
// present; the concrete scenarios produced by expandMatrix are validated later by
// validate (which also catches name collisions across expanded rows).
func validateMatrix(s *spec.Spec) []string {
	var errs []string
	for i := range s.Scenarios {
		sc := &s.Scenarios[i]
		if sc.Matrix == nil {
			continue
		}
		where := fmt.Sprintf("scenarios[%d]", i)
		if sc.Name != "" {
			where = fmt.Sprintf("scenario %q", sc.Name)
		}
		if len(sc.Matrix) == 0 {
			errs = append(errs, fmt.Sprintf("%s.matrix must contain at least one row", where))
			continue
		}
		for r, row := range sc.Matrix {
			if len(row) == 0 {
				errs = append(errs, fmt.Sprintf("%s.matrix[%d] must contain at least one variable", where, r))
			}
		}
	}
	return errs
}

// expandMatrix replaces every matrix scenario with one concrete scenario per row,
// in definition order (spec.md §22, ADR-0020). Each instance carries its row as
// Vars (seeded into the store before steps run) and a templated, unique Name.
// Scenarios without a matrix are left untouched.
func expandMatrix(s *spec.Spec) {
	hasMatrix := false
	for i := range s.Scenarios {
		if s.Scenarios[i].Matrix != nil {
			hasMatrix = true
			break
		}
	}
	if !hasMatrix {
		return
	}

	out := make([]spec.Scenario, 0, len(s.Scenarios))
	for i := range s.Scenarios {
		sc := s.Scenarios[i]
		if sc.Matrix == nil {
			out = append(out, sc)
			continue
		}
		for _, row := range sc.Matrix {
			inst := sc
			inst.Matrix = nil
			inst.Vars = row
			inst.Name = matrixInstanceName(sc.Name, row)
			out = append(out, inst)
		}
	}
	s.Scenarios = out
}

// matrixInstanceName renders a unique name for one matrix row. If the template
// references any row variable with ${var}, those are substituted (and the user
// owns uniqueness); otherwise a deterministic "[k=v ...]" suffix is appended so
// every instance stays distinct.
func matrixInstanceName(template string, row map[string]string) string {
	name := template
	referenced := false
	for k, v := range row {
		token := "${" + k + "}"
		if strings.Contains(name, token) {
			referenced = true
			name = strings.ReplaceAll(name, token, v)
		}
	}
	if referenced {
		return name
	}
	return template + " " + matrixRowSuffix(row)
}

// matrixRowSuffix is a deterministic "[k1=v1 k2=v2]" rendering of a row.
func matrixRowSuffix(row map[string]string) string {
	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteByte('[')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%s=%s", k, row[k])
	}
	b.WriteByte(']')
	return b.String()
}
