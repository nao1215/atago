package diag

// ATG5xxx — internal errors, which exit 5. Reaching one of these is a bug in
// atago, not something a spec can be written around.
//
// The family is deliberately blunt. A finer taxonomy would imply the reader can
// act on the distinction, and they cannot: every one of these means atago got
// into a state it does not handle, and the useful response is the same in all
// of them. What matters is that the code is stable enough to search an issue
// tracker with.
const internalSince = "v0.21.0"

// InternalError is atago failing in a way no spec can cause.
var InternalError = register(5001, "InternalError", Entry{
	Summary: "atago hit a state it does not handle",
	Detail:  "This is a bug in atago. It means an invariant the code relies on did not hold — a step reaching a runner that cannot serve it, a generated document that could not be produced, a resource that should exist and does not. A spec cannot be at fault: everything a spec can get wrong is refused while loading, with an ATG2xxx code.",
	Fix:     "Please report it at https://github.com/nao1215/atago/issues with the message, the atago version from `atago version`, and the spec if you can share it. There is no workaround to try first.",
	Since:   internalSince,
})
