package extension

import (
	"errors"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

func ParseHuman(text string, city string) (*when.Result, error) {
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	r, err := w.Parse(text, NowOnAddress(city))
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errors.New("no matches found")
	}

	return r, nil
}
