# 🕷️ Go Concurrent Web Crawler

Um Web Crawler concorrente de alta performance construído em **Go (Golang)**, desenhado para demonstrar padrões avançados de concorrência, controlo de fluxo, injeção de dependência e manipulação segura de memória.

## 🚀 Destaques de Engenharia

- **Concorrência Controlada (Bounded Concurrency):** Utiliza *Worker Pools* baseados em semáforos (`chan struct{}`) e `sync.WaitGroup` para processar múltiplas páginas em paralelo sem sobrecarregar a rede ou o servidor de destino.
- **Thread-Safety & Prevenção de Loops:** Implementa `sync.RWMutex` e `sync.Mutex` para garantir leitura e escrita seguras na tabela de URLs visitadas, evitando *race conditions* e ciclos infinitos.
- **Gestão de Ciclo de Vida com `context`:** Suporte nativo a cancelamento em cascata e *timeouts* globais através de `context.Context`.
- **Injeção de Dependência & Testabilidade:** Desacoplamento do cliente HTTP através da interface `HTTPClient`, permitindo testes unitários rápidos e determinísticos com *mocks* (validados com a flag `-race`).
- **Standard Go Project Layout:** Organização clara separando o ponto de entrada da aplicação (`cmd/crawler`) das regras de domínio e infraestrutura (`internal/crawler`).

## 📂 Estrutura do Projeto

- `cmd/crawler/main.go`: CLI, parsing de flags e exportação JSON
- `internal/crawler/crawler.go`: Motor concorrente, goroutines e sincronização
- `internal/crawler/parser.go`: Parser DOM HTML e resolução de URLs relativas
- `internal/crawler/crawler_test.go`: Testes unitários com Mocks e validação Thread-Safe

## 🛠️ Como Executar

`go run ./cmd/crawler -url https://go.dev -depth 1 -workers 8 -timeout 30s -out report.json`

## 🧪 Executar os Testes Unitários

`go test -v -race ./...`
