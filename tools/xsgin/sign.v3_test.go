package xsign

import (
	"net/url"
	"testing"
)

func TestSignV3_CheckSum(t *testing.T) {
	values := make(url.Values)
	values.Set("a", "15")
	values.Set("b", "1/2")
	values.Set("c", "1")
	values.Set("x-ts", "12318472123")

	body := "{\"b\":\"hello\",\"a\":123}"

	type fields struct {
		data    string
		signKey string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "1",
			fields: fields{
				data:    values.Encode() + body,
				signKey: "xxxxx",
			},
			want: "8663c6e9707bfaa0fecd73a938aacf24",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewSignV3(tt.fields.signKey)
			x.Write([]byte(tt.fields.data))
			if got := x.CheckSum(); got != tt.want {
				t.Errorf("CheckSum() = %v, want %v", got, tt.want)
			}
		})
	}
}
