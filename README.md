Passo a passo para execução da aplicação:

* Para subir as dependencias do projeto execute o comando:
```
    docker-compose up -d --build
```
A aplicação estará de pé respondendo na porta: 8080

### 1. Testes:
   * Há 2 testes em nossa pilha, 
   * O primeiro teste é baseado em IP (limite de requisições: 10, tempo de bloqueio 2s).
     * Envia 10 requisições recebendo response status code 200.
     * Envia 2 requisições recebendo response status code 429.
     * Executa sleep de 1 segundo.
     * Envia 1 requisição recebendo response status code 429.
     * Executa sleep de 1 segundo.
     * Envia 1 requisição recebendo response status code 200.
   * O segundo teste é baseado em Token (limite de requisições: 100, tempo de bloqueio 3s).
     * Envia 100 requisições recebendo response status code 200.
     * Envia 2 requisições recebendo response status code 429.
     * Executa sleep de 1 segundo.
     * Envia 1 requisição recebendo response status code 429.
     * Executa sleep de 2 segundo.
     * Envia 1 requisição recebendo response status code 200.

### 2. Rodando de testes:
    * Com o ambiente e as dependencias em pé, execute o comando para entrar no container da aplicação:
```
   docker-compose exec goapp bash
```
   * Dentro do container execute o comando:
```
   go test ./...
```

   * Também é possivel ver o fluxo de request dos testes executando fora do container:
```
   docker-compose logs goapp
```