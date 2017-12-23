package redisstats

import (
	"reflect"
	"testing"

	"github.com/sniperkit/stats"
)

func validateMeasure(t *testing.T, found stats.Measure, expect stats.Measure) {
	if !reflect.DeepEqual(found, expect) {
	}
}
