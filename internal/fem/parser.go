package fem

type Annotation struct {
	Type string
	Text string
	Line int
}

func Parse(content string) ([]Annotation, string, error) {
	return nil, content, nil
}
