package spec

// CDP drives a headless browser through a named browser runner. The
// action list runs in order against one browser session; the value captured by
// the last `text`/`eval` action feeds the `value` assertion target and
// `store from.value`.
type CDP struct {
	Runner  string      `yaml:"runner"`
	Actions []CDPAction `yaml:"actions"`
}

// CDPAction is one browser action; exactly one field is set. The action set is
// intentionally small and declarative — no conditions, loops, or expression
// language (#50).
type CDPAction struct {
	Navigate    string         `yaml:"navigate,omitempty"`     // load a URL
	WaitVisible string         `yaml:"wait_visible,omitempty"` // wait until a selector is visible
	WaitHidden  string         `yaml:"wait_hidden,omitempty"`  // wait until a selector is hidden/absent
	Click       string         `yaml:"click,omitempty"`        // click a selector
	Press       *CDPPress      `yaml:"press,omitempty"`        // press a key on a selector
	Select      *CDPSelect     `yaml:"select,omitempty"`       // choose an <option> in a <select>
	Check       string         `yaml:"check,omitempty"`        // tick a checkbox selector
	Uncheck     string         `yaml:"uncheck,omitempty"`      // untick a checkbox selector
	Screenshot  *CDPScreenshot `yaml:"screenshot,omitempty"`   // write a PNG into the workdir
	Text        string         `yaml:"text,omitempty"`         // capture a selector's text
	Title       bool           `yaml:"title,omitempty"`        // capture the page title
	Attribute   *CDPAttribute  `yaml:"attribute,omitempty"`    // capture an element attribute
	Eval        string         `yaml:"eval,omitempty"`         // evaluate JS, capture the result as JSON
	SendKeys    *CDPSendKeys   `yaml:"send_keys,omitempty"`    // type into a selector
	Upload      *CDPUpload     `yaml:"upload,omitempty"`       // set a file on an <input type=file>
	Download    *CDPDownload   `yaml:"download,omitempty"`     // click to trigger a download, capture the file
}

// CDPUpload sets File on the <input type=file> matched by Selector (#75). File is
// resolved against the scenario workdir and must exist there; the browser surface
// stays black-box (no scripted file dialogs).
type CDPUpload struct {
	Selector string `yaml:"selector"`
	File     string `yaml:"file"`
}

// CDPDownload triggers a download by clicking Click and captures the downloaded
// file into Dir (a workdir-relative directory, default the workdir root) using
// the server-suggested filename (#75). The captured value is the final filename,
// so existing file/dir/pdf/image assertions can validate the downloaded artifact.
type CDPDownload struct {
	Click string `yaml:"click"`
	Dir   string `yaml:"dir,omitempty"`
}

// CDPSendKeys types Value into the element matched by Selector.
type CDPSendKeys struct {
	Selector string `yaml:"selector"`
	Value    string `yaml:"value"`
}

// CDPPress presses a single key (e.g. "Enter", "Tab", or a printable character)
// on the element matched by Selector.
type CDPPress struct {
	Selector string `yaml:"selector"`
	Key      string `yaml:"key"`
}

// CDPSelect chooses the option whose value is Value in the <select> matched by
// Selector.
type CDPSelect struct {
	Selector string `yaml:"selector"`
	Value    string `yaml:"value"`
}

// CDPScreenshot writes a PNG of the page (or of Selector when set) to Path,
// resolved against the scenario workdir, so existing file/image assertions can
// inspect it.
type CDPScreenshot struct {
	Path     string `yaml:"path"`
	Selector string `yaml:"selector,omitempty"`
}

// CDPAttribute captures the Name attribute of the element matched by Selector
// into the value assertion path (like text/eval).
type CDPAttribute struct {
	Selector string `yaml:"selector"`
	Name     string `yaml:"name"`
}
