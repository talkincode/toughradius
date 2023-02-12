package des

import (
	"reflect"
	"testing"
)

func TestDesDecrypt(t *testing.T) {
	type args struct {
		src []byte
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "1",
			args:    args{
				src: []byte{108,116,65,67,105,72,106,86,106,73,110,43,117,86,109,51,49,71,81,118,121,119,61,61},
				key: []byte("12345678"),
			},
			want:    []byte("12345678"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DesDecrypt(tt.args.src, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DesDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DesDecrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDesEncrypt(t *testing.T) {
	type args struct {
		src []byte
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "1",
			args:    args{
				src: []byte("12345678"),
				key: []byte("12345678"),
			},
			want:    []byte{108,116,65,67,105,72,106,86,106,73,110,43,117,86,109,51,49,71,81,118,121,119,61,61},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DesEncrypt(tt.args.src, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DesEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DesEncrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
