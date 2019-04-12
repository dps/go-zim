package zim

import (
	"fmt"
	"testing"
)

const expectedSum = "d40ddf09cd6ef66affa415806ca5e5b7"

func TestInternalChecksum(t *testing.T) {
	var internalSum, sumErr = z.InternalChecksum()
	if sumErr != nil {
		t.Error(sumErr)
	}
	if s := fmt.Sprintf("%x", internalSum); s != expectedSum {
		t.Errorf("z.InternalChecksum() = %s; want %s", s, expectedSum)
	}
}

func TestCalculatedChecksum(t *testing.T) {
	var calculatedSum, sumErr = z.CalculateChecksum()
	if sumErr != nil {
		t.Error(sumErr)
	}
	if s := fmt.Sprintf("%x", calculatedSum); s != expectedSum {
		t.Errorf("z.CalculateChecksum() = %s; want %s", s, expectedSum)
	}
}

func TestValidateChecksum(t *testing.T) {
	if e := z.ValidateChecksum(); e != nil {
		t.Error(e)
	}
}
