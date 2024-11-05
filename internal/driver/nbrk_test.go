package driver

import (
	"testing"
)

func Test_canPerformCurrency(t *testing.T) {
	type args struct {
		currency string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "USD",
			args: args{currency: "USD"},
			want: true,
		},
		{
			name: "EUR",
			args: args{currency: "EUR"},
			want: true,
		},
		{
			name: "RUB",
			args: args{currency: "RUB"},
			want: true,
		},
		{
			name: "unsupported",
			args: args{currency: "unsupported"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canPerformCurrency(tt.args.currency); got != tt.want {
				t.Errorf("canPerformCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}
