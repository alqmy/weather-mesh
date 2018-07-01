package members

import "time"

type Member struct {
	Address string
	Name    string
	Last    time.Time
}

type Members map[string]Member

func (m Members) List() []Member {
	l := []Member{}

	for _, mem := range m {
		l = append(l, mem)
	}

	return l
}
