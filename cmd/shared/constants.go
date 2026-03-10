package shared

const (
	// ErrExclusiveWithSingleSelect is returned when exclusive options are used with single select
	ErrExclusiveWithSingleSelect = "exclusive options are not allowed when answer_limit is 1 (single select)"

	// ErrNoNonExclusiveOptions is returned when all options are exclusive
	ErrNoNonExclusiveOptions = "at least one non-exclusive option is required when using exclusive options"
)
