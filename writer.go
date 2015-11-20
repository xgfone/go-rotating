package rotating

type NullWriter struct {
}

func NewNullWriter() NullWriter {
	return NullWriter{}
}

func (w NullWriter) Write(d []byte) (int, error) {
	return len(d), nil
}
