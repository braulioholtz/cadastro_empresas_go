package models

type Empresa struct {
	ID              string `json:"id,omitempty" bson:"_id,omitempty"`
	CNPJ            string `json:"cnpj" bson:"cnpj"`
	NomeFantasia    string `json:"nome_fantasia" bson:"nome_fantasia"`
	RazaoSocial     string `json:"razao_social" bson:"razao_social"`
	Endereco        string `json:"endereco" bson:"endereco"`
	NumFuncionarios int    `json:"num_funcionarios" bson:"num_funcionarios"`
	NumMinPCD       int    `json:"num_min_pcd" bson:"num_min_pcd"`
}
