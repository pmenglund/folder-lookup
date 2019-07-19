package fetcher

import (
	"context"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		conf Config
		want *Fetcher
	}{
		{"test", Config{}, nil},
	}

	ctx := context.TODO()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(ctx, tt.conf)
			if err != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %+v, want %v", got, tt.want)
			}
		})
	}
}
