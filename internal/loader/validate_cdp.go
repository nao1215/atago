package loader

import (
	"fmt"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/spec"
)

// validateCDPActions checks that each browser action sets exactly one action key
// and supplies its required sub-fields, so a malformed cdp step fails at load
// time with a clear message rather than mid-run (#50).
func validateCDPActions(add addFunc, where string, actions []spec.CDPAction) {
	for i, a := range actions {
		aw := fmt.Sprintf("%s.cdp.actions[%d]", where, i)
		if n := cdpActionCount(&a); n == 0 {
			add(diag.ChooseExactlyOne, "%s sets no recognized action", aw)
			continue
		} else if n > 1 {
			add(diag.ChooseExactlyOne, "%s sets multiple actions; set exactly one", aw)
		}
		switch {
		case a.Press != nil:
			if a.Press.Selector == "" || a.Press.Key == "" {
				add(diag.RequiredKey, "%s.press requires selector and key", aw)
			}
		case a.Select != nil:
			if a.Select.Selector == "" {
				add(diag.RequiredKey, "%s.select requires a selector", aw)
			}
		case a.Screenshot != nil:
			if a.Screenshot.Path == "" {
				add(diag.RequiredKey, "%s.screenshot requires a path", aw)
			}
		case a.Attribute != nil:
			if a.Attribute.Selector == "" || a.Attribute.Name == "" {
				add(diag.RequiredKey, "%s.attribute requires selector and name", aw)
			}
		case a.SendKeys != nil:
			if a.SendKeys.Selector == "" {
				add(diag.RequiredKey, "%s.send_keys requires a selector", aw)
			}
		case a.Upload != nil:
			if a.Upload.Selector == "" || a.Upload.File == "" {
				add(diag.RequiredKey, "%s.upload requires selector and file", aw)
			}
		case a.Download != nil:
			if a.Download.Click == "" {
				add(diag.RequiredKey, "%s.download requires a click selector", aw)
			}
		}
	}
}

// cdpActionCount reports how many action keys are set on one browser action.
func cdpActionCount(a *spec.CDPAction) int {
	n := 0
	for _, set := range []bool{
		a.Navigate != "", a.WaitVisible != "", a.WaitHidden != "", a.Click != "",
		a.Press != nil, a.Select != nil, a.Check != "", a.Uncheck != "",
		a.Screenshot != nil, a.Text != "", a.Title, a.Attribute != nil,
		a.Eval != "", a.SendKeys != nil, a.Upload != nil, a.Download != nil,
	} {
		if set {
			n++
		}
	}
	return n
}
