package output

// Output - contract for supported ouput formats
type Output interface {
	Write() error
}
