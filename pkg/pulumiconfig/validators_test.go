// Package pulumiconfig contains tests for validating the behavior of utility functions.
// This file specifically tests the `string2Number` function, which converts strings to numbers.

package pulumiconfig

import (
	"reflect"
	"testing"
)

// It ensures that the function correctly converts strings to numeric types (e.g., int64, uint64, float64)
// and handles invalid inputs or unsupported conversion types.
func Test_string2Number(t *testing.T) {
	type args struct {
		s string
		t ConvertType
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Int64 Conversion Successful",
			args: args{
				s: "12345",
				t: Int64,
			},
			want:    int64(12345),
			wantErr: false,
		},
		{
			name: "Uint64 Conversion Successful",
			args: args{
				s: "12345",
				t: Uint64,
			},
			want:    uint64(12345),
			wantErr: false,
		},
		{
			name: "Float64 Conversion Successful",
			args: args{
				s: "123.45",
				t: Float64,
			},
			want:    float64(123.45),
			wantErr: false,
		},
		{
			name: "Int64 Conversion Failure",
			args: args{
				s: "invalid",
				t: Int64,
			},
			want:    int64(0),
			wantErr: true,
		},
		{
			name: "Unsupported Conversion Type",
			args: args{
				s: "12345",
				t: ConvertType("invalidType"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := string2Number(tt.args.s, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("string2Number() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("string2Number() = %v, want %v", got, tt.want)
			}
		})
	}
}
