package output

// Output - all output formatters (stateful or not)
// are composable by meeting this contract
type Output interface {
	Write() error
}
