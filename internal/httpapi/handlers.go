package httpapi

import (
	"encoding/json"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"matriz/internal/messaging"
	"matriz/internal/models"
	"matriz/internal/repository"
	"matriz/internal/validation"

	"github.com/go-chi/chi/v5"
)

// parseEmpresaFromRequest converte o corpo da requisição em models.Empresa,
// suporta application/x-www-form-urlencoded e multipart/form-data.
// Limite de memória para multipart: ~10MB.
// Retorna http.ErrNotSupported quando o Content-Type não é suportado.
func parseEmpresaFromRequest(r *http.Request) (models.Empresa, error) {
	var e models.Empresa
	ct := r.Header.Get("Content-Type")
	mediatype, _, _ := mime.ParseMediaType(ct)
	switch strings.ToLower(mediatype) {
	case "application/x-www-form-urlencoded", "multipart/form-data", "":
		// Parse de form values. Para multipart, ParseMultipartForm; para urlencoded, ParseForm.
		if strings.HasPrefix(strings.ToLower(mediatype), "multipart/") {
			_ = r.ParseMultipartForm(10 << 20) // 10MB
		} else {
			_ = r.ParseForm()
		}
		form := r.Form
		e.CNPJ = strings.TrimSpace(form.Get("cnpj"))
		e.NomeFantasia = strings.TrimSpace(form.Get("nome_fantasia"))
		e.RazaoSocial = strings.TrimSpace(form.Get("razao_social"))
		e.Endereco = strings.TrimSpace(form.Get("endereco"))
		if v := strings.TrimSpace(form.Get("num_funcionarios")); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				e.NumFuncionarios = n
			}
		}
		if v := strings.TrimSpace(form.Get("num_min_pcd")); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				e.NumMinPCD = n
			}
		}
		return e, nil
	default:
		return e, http.ErrNotSupported
	}
}

type Server struct {
	repo repository.EmpresaStore
	pub  *messaging.Publisher
}

// NewServer cria uma instância do servidor HTTP com as dependências de
// repositório e publisher (opcional). Se pub for nil, eventos não serão publicados.
func NewServer(repo repository.EmpresaStore, pub *messaging.Publisher) *Server {
	return &Server{repo: repo, pub: pub}
}

// Routes registra e retorna as rotas HTTP do serviço de empresas.
// Endpoints:
// - POST   /empresas
// - GET    /empresas
// - GET    /empresas/{id}
// - PUT    /empresas/{id}
// - DELETE /empresas/{id}
func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/empresas", s.create)
	r.Get("/empresas", s.list)
	r.Get("/empresas/{id}", s.get)
	r.Put("/empresas/{id}", s.update)
	r.Delete("/empresas/{id}", s.delete)
	return r
}

// writeJSON escreve uma resposta JSON com o status informado.
// Se v for nil, apenas os headers e o status são enviados.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError padroniza respostas de erro no formato {"error": "..."}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// create trata POST /empresas.
// Status:
// - 201 em caso de sucesso (retorna {"id": "<novo_id>"}).
// - 400 para erros de validação/negócio (ex.: "cnpj já cadastrado").
// publica mensagem de "Cadastro de EMPRESA ..." se Publisher estiver configurado.
func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	e, err := parseEmpresaFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "form inválido")
		return
	}
	if err := validation.ValidateCNPJ(e.CNPJ); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := s.repo.Create(r.Context(), &e)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if s.pub != nil {
		_ = s.pub.Publish("Cadastro de EMPRESA " + e.NomeFantasia)
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// list trata GET /empresas.
// Status:
// - 200 com a lista de empresas.
// - 500 em caso de falha no repositório.
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	items, err := s.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// get trata GET /empresas/{id}.
// Status:
// - 200 com a empresa encontrada.
// - 404 se não existir.
func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := s.repo.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "não encontrado")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// update trata PUT /empresas/{id}.
// Status:
// - 200 em caso de sucesso.
// - 400 para erros de validação/negócio (inclui unicidade de CNPJ).
// publica "Edição da EMPRESA ..." se Publisher estiver configurado.
func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	e, err := parseEmpresaFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "form inválido")
		return
	}

	if err := validation.ValidateCNPJ(e.CNPJ); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.repo.Update(r.Context(), id, &e); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if s.pub != nil {
		_ = s.pub.Publish("Edição da EMPRESA " + e.NomeFantasia)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// delete trata DELETE /empresas/{id}.
// Status:
// - 200 em caso de sucesso.
// - 500 em falha de exclusão.
func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Buscar o item antes para obter o nome na mensagem de evento.
	item, _ := s.repo.Get(r.Context(), id)
	if err := s.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	name := ""
	if item != nil {
		name = item.NomeFantasia
	}
	if s.pub != nil {
		_ = s.pub.Publish("Exclusão da EMPRESA " + name)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
