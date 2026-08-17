package spec

// This file is the single field list of what each step kind ${name}-expands.
//
// Every walker returns a copy of its input with visit applied to exactly the
// author-written strings the engine expands before the step runs. The engine
// expands by passing store.Expand as visit; the variable summaries collect by
// passing a recording visit; the security notes find ${env:} reads the same
// way. One list, three consumers — because the drift between the engine's
// hand-written expand functions and the summaries' hand-written field lists is
// how run's env, http's body_file/body_to/form/files, cdp's upload/download,
// and pty's paste/exec each went uncounted in some summary while the engine
// expanded them. WalkAssertStrings pioneered the shape for assertions, where
// that drift never happened; these give every other kind the same immunity.
//
// Two rules keep the walkers honest:
//
//   - Copy first, then visit. A walker starts from a full copy of its input
//     (`c := *x`), so a field it does not know about is CARRIED, not dropped —
//     the pty expansion used to rebuild its copy from the zero value, and a
//     newly added Resize field silently vanished from the executed session.
//   - Only what the step's own expansion touches. A retry's `until:` assertion
//     is expanded per attempt by the poll loop (through WalkAssertStrings), a
//     stdin base64 payload must stay byte-exact, and durations/named keys are
//     vocabulary, not text; none of those belong here.
//
// TestWalkStepStrings_VisitsEveryExpandedField pins each walker's field list,
// so widening or narrowing what a kind expands is an explicit decision made in
// one place and reviewed in one diff.

// walkMapValues applies visit to each value of a string map, leaving keys
// untouched (mirroring store.ExpandMap), or returns the map unchanged when
// empty so nil stays nil.
func walkMapValues(m map[string]string, visit func(string) string) map[string]string {
	if len(m) == 0 {
		return m
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = visit(v)
	}
	return out
}

// WalkFixtureStrings covers a fixture's path and payload sources. base64 is
// binary and deliberately exempt.
func WalkFixtureStrings(f *Fixture, visit func(string) string) *Fixture {
	if f == nil {
		return nil
	}
	c := *f
	c.File = visit(f.File)
	c.Content = visit(f.Content)
	c.From = visit(f.From)
	c.Symlink = visit(f.Symlink)
	return &c
}

// WalkRunStrings covers what expandRun performs at step start: the command,
// cwd, stdin text and file, the redirect targets, and env values. The retry
// `until:` assertion is expanded per attempt by the poll loop and stays out.
func WalkRunStrings(r *Run, visit func(string) string) *Run {
	if r == nil {
		return nil
	}
	c := *r
	c.Command = visit(r.Command)
	c.Cwd = visit(r.Cwd)
	c.Stdin.Inline = visit(r.Stdin.Inline)
	c.Stdin.File = visit(r.Stdin.File)
	c.StdoutTo = visit(r.StdoutTo)
	c.StderrTo = visit(r.StderrTo)
	c.Env = walkMapValues(r.Env, visit)
	return &c
}

// WalkServiceStrings covers a service's command, cwd, env values, and the
// readiness probes (file, port, log regexp) expandService resolves.
func WalkServiceStrings(svc *Service, visit func(string) string) *Service {
	if svc == nil {
		return nil
	}
	c := *svc
	c.Command = visit(svc.Command)
	c.Cwd = visit(svc.Cwd)
	c.Env = walkMapValues(svc.Env, visit)
	if svc.Ready != nil {
		rc := *svc.Ready
		rc.File = visit(svc.Ready.File)
		rc.Port = visit(svc.Ready.Port)
		rc.Log = visit(svc.Ready.Log)
		c.Ready = &rc
	}
	return &c
}

// WalkStoreStrings covers the one store field the engine expands: a file
// source's path. The selectors are matchers over produced output, not text the
// engine substitutes into.
func WalkStoreStrings(s *Store, visit func(string) string) *Store {
	if s == nil {
		return nil
	}
	c := *s
	if s.From != nil && s.From.File != nil {
		fromCopy := *s.From
		fileCopy := *s.From.File
		fileCopy.Path = visit(s.From.File.Path)
		fromCopy.File = &fileCopy
		c.From = &fromCopy
	}
	return &c
}

// WalkHTTPStrings covers what expandHTTP performs: path, header values, the
// JSON body's string leaves, the raw/file bodies, the download target, form
// values, and multipart file paths. Retry stays out, as with run.
func WalkHTTPStrings(h *HTTP, visit func(string) string) *HTTP {
	if h == nil {
		return nil
	}
	c := *h
	c.Path = visit(h.Path)
	c.Header = walkMapValues(h.Header, visit)
	c.JSON = WalkJSONValueStrings(h.JSON, visit)
	c.Body = visit(h.Body)
	c.BodyFile = visit(h.BodyFile)
	c.BodyTo = visit(h.BodyTo)
	c.Form = walkMapValues(h.Form, visit)
	if len(h.Files) > 0 {
		c.Files = make([]FilePart, len(h.Files))
		for i, f := range h.Files {
			f.Path = visit(f.Path)
			c.Files[i] = f
		}
	}
	return &c
}

// WalkQueryStrings covers the one query field the engine expands: the SQL
// text. The runner name is resolved against the declared table, not expanded.
func WalkQueryStrings(q *Query, visit func(string) string) *Query {
	if q == nil {
		return nil
	}
	c := *q
	c.SQL = visit(q.SQL)
	return &c
}

// WalkGRPCStrings covers what expandGRPC performs: the method, header values,
// and the JSON request message's string leaves.
func WalkGRPCStrings(g *GRPC, visit func(string) string) *GRPC {
	if g == nil {
		return nil
	}
	c := *g
	c.Method = visit(g.Method)
	c.Header = walkMapValues(g.Header, visit)
	c.JSON = WalkJSONValueStrings(g.JSON, visit)
	return &c
}

// WalkCDPStrings covers a browser step's action arguments, one action at a
// time through WalkCDPActionStrings.
func WalkCDPStrings(c *CDP, visit func(string) string) *CDP {
	if c == nil {
		return nil
	}
	out := *c
	if len(c.Actions) > 0 {
		out.Actions = make([]CDPAction, len(c.Actions))
		for i, a := range c.Actions {
			out.Actions[i] = WalkCDPActionStrings(a, visit)
		}
	}
	return &out
}

// WalkCDPActionStrings covers every author-written argument of one browser
// action: selectors, URLs, values, key names, file paths, and the eval script.
func WalkCDPActionStrings(a CDPAction, visit func(string) string) CDPAction {
	a.Navigate = visit(a.Navigate)
	a.WaitVisible = visit(a.WaitVisible)
	a.WaitHidden = visit(a.WaitHidden)
	a.Click = visit(a.Click)
	a.Check = visit(a.Check)
	a.Uncheck = visit(a.Uncheck)
	a.Text = visit(a.Text)
	a.Eval = visit(a.Eval)
	if a.SendKeys != nil {
		sk := *a.SendKeys
		sk.Selector = visit(sk.Selector)
		sk.Value = visit(sk.Value)
		a.SendKeys = &sk
	}
	if a.Press != nil {
		p := *a.Press
		p.Selector = visit(p.Selector)
		p.Key = visit(p.Key)
		a.Press = &p
	}
	if a.Select != nil {
		s := *a.Select
		s.Selector = visit(s.Selector)
		s.Value = visit(s.Value)
		a.Select = &s
	}
	if a.Screenshot != nil {
		s := *a.Screenshot
		s.Path = visit(s.Path)
		s.Selector = visit(s.Selector)
		a.Screenshot = &s
	}
	if a.Attribute != nil {
		at := *a.Attribute
		at.Selector = visit(at.Selector)
		at.Name = visit(at.Name)
		a.Attribute = &at
	}
	if a.Upload != nil {
		up := *a.Upload
		up.Selector = visit(up.Selector)
		up.File = visit(up.File)
		a.Upload = &up
	}
	if a.Download != nil {
		dl := *a.Download
		dl.Click = visit(dl.Click)
		dl.Dir = visit(dl.Dir)
		a.Download = &dl
	}
	return a
}

// WalkPTYStrings covers a pty step's command, cwd, env values, and session.
// The engine layers scenario env and the TERM default on top after walking;
// those are runtime composition, not author-written text.
func WalkPTYStrings(p *PTY, visit func(string) string) *PTY {
	if p == nil {
		return nil
	}
	c := *p
	c.Command = visit(p.Command)
	c.Cwd = visit(p.Cwd)
	c.Env = walkMapValues(p.Env, visit)
	if len(p.Session) > 0 {
		c.Session = make([]PTYAction, len(p.Session))
		for i, a := range p.Session {
			c.Session[i] = WalkPTYActionStrings(a, visit)
		}
	}
	return &c
}

// WalkPTYActionStrings covers one session entry: the expect pattern, a send's
// verbatim text and paste payload (named keys are fixed byte sequences), an
// exec's host command, and the expect_screen matcher strings. Starting from
// the full copy is what carries resize — and whatever entry arrives next —
// without a list to forget it from.
func WalkPTYActionStrings(a PTYAction, visit func(string) string) PTYAction {
	a.Expect = visit(a.Expect)
	if a.Send != nil {
		cs := *a.Send
		if cs.Text != nil {
			txt := visit(*cs.Text)
			cs.Text = &txt
		}
		if cs.Paste != nil {
			pasted := visit(*cs.Paste)
			cs.Paste = &pasted
		}
		a.Send = &cs
	}
	if a.Exec != nil {
		ce := *a.Exec
		ce.Command = visit(ce.Command)
		a.Exec = &ce
	}
	a.ExpectScreen = WalkPTYExpectScreenStrings(a.ExpectScreen, visit)
	return a
}

// WalkSignalStrings covers the one signal field the engine expands: the target
// service name. The signal itself is vocabulary and the wait timeout a
// duration.
func WalkSignalStrings(sg *Signal, visit func(string) string) *Signal {
	if sg == nil {
		return nil
	}
	c := *sg
	c.Service = visit(sg.Service)
	return &c
}
