package validation

import (
	"errors"
	"strings"
)

// ValidateCNPJ por agora apenas verifica se não é vazio
// se necessário pode ser incluido validações de CNPJ válido.
func ValidateCNPJ(cnpj string) error {
	cnpj = strings.TrimSpace(cnpj)
	if cnpj == "" {
		return errors.New("cnpj obrigatório")
	}

	return nil
}
