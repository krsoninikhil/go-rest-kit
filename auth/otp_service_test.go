package auth

import (
	"fmt"
	"testing"
)

func Test_generateOTP(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name string
		args args
	}{
		{"test for length 4", args{4}},
		{"test for length 6", args{6}},
		{"test for length 6", args{6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateOTP(tt.args.length); len(got) != tt.args.length {
				t.Errorf("generateOTP() = %v, length want %v", got, tt.args.length)
			} else {
				fmt.Printf("got: %v\n", got)
			}
		})
	}
}
