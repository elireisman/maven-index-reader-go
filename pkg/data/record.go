package data

type Record struct {
	Name         string
	RepositoryID int64
	// TODO(eli): fill in the well-formed Record fields!

}

func NewRecord(raw map[string]string) (Record, error) {
	// TODO(eli): validate well-formed fields before applying to Record!
	return Record{}, nil
}
