package date_util_test

import (
	"github.com/keep94/toolbox/date_util"
	"testing"
	"time"
)

func TestYMD(t *testing.T) {
	actual := date_util.YMD(2013, 11, 14)
	expected, _ := time.Parse(date_util.YMDFormat, "20131114")
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
