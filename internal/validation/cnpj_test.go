package validation

import "testing"

func TestValidateCNPJ(t *testing.T) {
	type args struct {
		cnpj string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "válido com pontuação",
			args:    args{cnpj: "04.252.011/0001-10"},
			wantErr: false,
		},
		{
			name:    "válido sem pontuação",
			args:    args{cnpj: "04252011000110"},
			wantErr: false,
		},
		{
			name:    "string vazia",
			args:    args{cnpj: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCNPJ(tt.args.cnpj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCNPJ() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
