package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const GitHubAPIURL = "https://api.github.com/search/repositories"
// SearchResult mapeia os campos principais da resposta da API do GitHub
type SearchResult struct {
	TotalCount int          `json:"total_count"`
	Items      []Repository `json:"items"` // Um slice de repositórios
}

// Repository mapeia os campos de um item de repositório individual
// Estamos interessados apenas em alguns campos (as "features").
type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	URL         string `json:"html_url"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"` // A "feature" que usaremos para ordenar
	Forks       int    `json:"forks_count"`
}

/**
 * searchRepositories é a função principal que consome a API.
 * Ela é responsável por construir a query, fazer a chamada e decodificar a resposta.
 */
func searchRepositories(client *http.Client, query, sortBy, order string) (*SearchResult, error) {
	// 1. Construir a URL com parâmetros de forma segura
	params := url.Values{}
	params.Add("q", query)   // O termo de busca (ex: "language:go")
	params.Add("sort", sortBy) // A "FEATURE" para ordenar (ex: "stars")
	params.Add("order", order) // A direção (ex: "desc")

	fullURL := fmt.Sprintf("%s?%s", GitHubAPIURL, params.Encode())
	log.Printf("Querying GitHub API: %s\n", fullURL)

	// 2. Criar a requisição GET
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição: %w", err)
	}

	// 3. Definir Headers OBRIGATÓRIOS da API do GitHub
	// Sem eles, a API retornará 403 Forbidden.
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "my-golang-app") // A API exige um User-Agent

	// 4. Executar a requisição
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao executar requisição: %w", err)
	}
	defer resp.Body.Close() // Boa prática: sempre fechar o corpo da resposta

	// 5. Tratar códigos de status não-OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API do GitHub retornou status não-OK: %s", resp.Status)
	}

	// 6. Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler corpo da resposta: %w", err)
	}

	// 7. Decodificar (Unmarshal) o JSON na nossa struct
	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON: %w", err)
	}

	return &result, nil
}

func main() {
	// Criamos um cliente HTTP com um timeout. Isso é uma boa prática
	// para evitar que nossa aplicação fique presa indefinidamente.
	client := &http.Client{Timeout: 10 * time.Second}

	// --- Definição da nossa busca ---
	// Termo de busca: repositórios da linguagem Go
	query := "language:go"
	// A "FEATURE" pela qual queremos ordenar
	sortByFeature := "stars"
	// A ordem da ordenação
	order := "desc"
	// ---------------------------------

	fmt.Printf("Buscando repositórios no GitHub...\nQuery: '%s', Sort By: '%s', Order: '%s'\n\n", query, sortByFeature, order)

	// Chama nossa função
	result, err := searchRepositories(client, query, sortByFeature, order)
	if err != nil {
		log.Fatalf("ERRO: %v", err) // `log.Fatalf` encerra o programa em caso de erro
	}

	// --- Aqui "tratamos os dados de resposta" ---
	// Vamos apenas imprimir os 10 primeiros de forma organizada.
	fmt.Printf("Encontrados %d repositórios. Mostrando os 10 primeiros:\n", result.TotalCount)
	fmt.Println("---------------------------------------------------------")

	// Itera sobre os items (repositórios) retornados
	for i, repo := range result.Items {
		if i >= 10 { // Limita a 10 resultados
			break
		}
		fmt.Printf("#%d: %s\n", i+1, repo.FullName)
		fmt.Printf("   ⭐ Estrelas: %d\n", repo.Stars)
		fmt.Printf("   🍴 Forks:    %d\n", repo.Forks)
		fmt.Printf("   🔗 URL:       %s\n", repo.URL)
		fmt.Printf("   %s\n\n", repo.Description)
	}
}