package log

import (
	"testing"
)

func Test_newPatternRender(t *testing.T) {
	newPatternRender("%T%t")
}
