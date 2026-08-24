package cmd

import "sort"

/*
exportSpec captures the per-feature differences of an export so the shared sort + emit
stages can run both the ssh and x509 paths without each handler repeating them. The
feature-specific record loop (parse, revocation/provisioner lookup, serial extraction,
validity classification) stays in each handler; this spec supplies only what the shared
stages need.
*/
type exportSpec[T any] struct {
	columns []tColumn[T]
	// rowLabel returns a human-readable label for a row, used by the emit helpers in logging.
	rowLabel func(T) string
	// finishLess / startLess are the ordering predicates for SORT_FINISH and SORT_START.
	finishLess func(a, b T) bool
	startLess  func(a, b T) bool
	// otherEmit renders a format that is not shared across features (e.g. x509's openssl).
	// Nil for features that have no such format.
	otherEmit func(rows []T)
}

/*
sortRecords sorts records in place according to config.sortOrder, using the feature-specific
ordering predicate from spec. Shared by both export paths.
*/
func sortRecords[T any](records []T, spec exportSpec[T]) {
	less := spec.finishLess
	if config.sortOrder.Value == SORT_START {
		less = spec.startLess
	}
	sort.SliceStable(records, func(i, j int) bool { return less(records[i], records[j]) })
}

/*
emitRecords dispatches rows to the emit format selected by choice. table/json/markdown/plain
are identical across features; any other value is handed to spec.otherEmit (nil-safe). Shared
by both export paths.
*/
func emitRecords[T any](rows []T, choice *tChoice, spec exportSpec[T]) {
	switch choice.Value {
	case FORMAT_JSON:
		emitJson(rows)
	case FORMAT_TABLE:
		emitTable(rows, spec.columns, spec.rowLabel)
	case FORMAT_MARKDOWN:
		emitMarkdown(rows, spec.columns)
	case FORMAT_PLAIN:
		emitPlain(rows, spec.columns)
	default:
		if spec.otherEmit != nil {
			spec.otherEmit(rows)
		}
	}
}

/*
selectValid reports whether a record with the given validity status is shown under the current
showValid / showRevoked / showExpired selection flags. Shared by both export paths.
*/
func selectValid(validity string) bool {
	return (config.showExpired && validity == EXPIRED_STR) ||
		(config.showRevoked && validity == REVOKED_STR) ||
		(config.showValid && validity == VALID_STR)
}
