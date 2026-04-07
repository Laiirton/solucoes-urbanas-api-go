# API Endpoints para o serviço Soluções Urbanas

ROTA DA API que já está funcionando: https://solucoes-urbanas-api-go.onrender.com/api/

## Autenticação pública

- `POST /api/auth/register`
  - Registra um novo usuário.
  - Payload JSON:
    ```json
    {
      "username": "usuario123",
      "email": "user@exemplo.com",
      "password": "senha_forte",
      "full_name": "Nome Completo",
      "cpf": "123.456.789-00",
      "birth_date": "01/01/1990",
      "type": "user"
    }
    ```
- `POST /api/auth/login`
  - Faz login do usuário.
  - Payload JSON:
    ```json
    {
      "username": "usuario123",
      "password": "senha_forte"
    }
    ```

## Geolocalização

- `GET /api/geolocation`
  - Busca informações de endereço/geolocalização a partir de parâmetros de query, como `street`.


## Notícias públicas

- `GET /api/news`
  - Lista todas as notícias.
- `GET /api/news/{id}`
  - Retorna uma notícia específica pelo `id`.

## Serviços públicos

- `GET /api/services`
  - Lista todos os serviços ativos por padrão.
- `GET /api/services/{id}`
  - Retorna detalhes de um serviço específico pelo `id`.

## Rotas protegidas (JWT obrigatório)

- `GET /api/auth/me`
  - Retorna dados do usuário autenticado.
- `POST /api/auth/logout`
  - Realiza logout simples e retorna mensagem de sucesso.
- `GET /api/home`
  - Retorna estatísticas/resumo da home com base no usuário autenticado e nos pedidos de serviço.
## Usuários

- `GET /api/users`
  - Lista todos os usuários.
- `GET /api/users/me`
  - Retorna o perfil do usuário autenticado.
- `GET /api/users/{id}`
  - Retorna dados do usuário especificado por `id`.
- `PUT /api/users/{id}`
  - Atualiza usuário por `id`.
  - Payload JSON (todos campos opcionais):
    ```json
    {
      "username": "novo_nick",
      "full_name": "Novo Nome Completo",
      "cpf": "111.222.333-44",
      "birth_date": "10/05/1995",
      "type": "admin"
    }
    ```
- `DELETE /api/users/{id}`
  - Exclui usuário por `id`.


## Notícias (escrita)

- `POST /api/news`
  - Cria uma notícia.
  - Aceita `multipart/form-data`:
    - `title`: (string) Título da notícia.
    - `content`: (string) Conteúdo formatado.
    - `files`: (file) Um ou mais arquivos de imagem.
- `PUT /api/news/{id}`
  - Atualiza notícia existente por `id`.
  - Aceita `multipart/form-data` similar ao POST.
- `DELETE /api/news/{id}`
  - Exclui notícia por `id`.

## Serviços (escrita)

- `POST /api/services`
  - Cria um novo serviço no catálogo.
  - Payload JSON:
    ```json
    {
      "title": "Coleta de Lixo Especial",
      "description": "Pedido de retirada de grandes volumes",
      "category": "Limpeza Urbana",
      "is_active": true,
      "form_schema": [
        {"name": "logradouro", "label": "Rua/Avenida", "type": "text", "required": true},
        {"name": "bairro", "label": "Bairro", "type": "text", "required": true},
        {"name": "volume", "label": "Qtd de Sacos", "type": "number", "required": false}
      ]
    }
    ```
- `PUT /api/services/{id}`
  - Atualiza um serviço existente.
  - Payload JSON (campos opcionais):
    ```json
    {
      "title": "Novo Título",
      "form_schema": [
        {"name": "ponto_referencia", "label": "Ponto de Referência", "type": "text", "required": false}
      ]
    }
    ```
- `DELETE /api/services/{id}`
  - Remove um serviço permanentemente.

## Pedidos de serviço

- `POST /api/service-requests`
  - Cria um novo pedido de serviço.
  - Payload JSON:
    ```json
    {
      "service_title": "Buraco na via",
      "category": "Infraestrutura",
      "request_data": {
        "address": "Av. Brasil, 450",
        "description": "Fisura no asfalto"
      }
    }
    ```
  - Também aceita `multipart/form-data` para anexos (`files`).
- `GET /api/service-requests`
  - Lista pedidos de serviço do usuário autenticado.
  - Use `?all=true` para listar todos os pedidos (modo administrativo simples).
- `GET /api/service-requests/{id}`
  - Retorna um pedido de serviço específico pelo `id`.
- `PATCH /api/service-requests/{id}/status`
  - Atualiza o status de um pedido de serviço por `id`.
  - Payload JSON:
    ```json
    {
      "status": "in_progress"
    }
    ```
    - Status permitidos: `pending`, `in_progress`, `completed`, `cancelled`.
- `DELETE /api/service-requests/{id}`
  - Exclui um pedido de serviço por `id`.
