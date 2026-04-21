# API Endpoints para o serviço Soluções Urbanas

ROTA DA API que já está funcionando: https://solucoes-urbanas-api-go.onrender.com/api/

## Health Check

- `GET /health`
  - Endpoint para verificar se a API está funcionando.
  - Resposta JSON:
    ```json
    {
      "status": "ok",
      "timestamp": "2026-04-21T18:00:00Z"
    }
    ```

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
  - Parâmetros de query (opcionais):
    - `search`: busca por título, slug ou conteúdo
    - `status`: filtra por status (draft, published)
    - `page`: número da página (padrão: 1)
    - `limit`: itens por página (padrão: 10)
- `GET /api/news/{id}`
  - Retorna uma notícia específica pelo `id` ou `slug`.

## Serviços públicos

- `GET /api/services`
  - Lista todos os serviços ativos por padrão.
  - Parâmetros de query (opcionais):
    - `all`: quando `true`, lista todos os serviços (incluindo inativos)
    - `search`: busca por texto
    - `page`: número da página (padrão: 1)
    - `limit`: itens por página (padrão: 10)
- `GET /api/services/{id}`
  - Retorna detalhes de um serviço específico pelo `id`.
  - Inclui estatísticas adicionais:
    - `average_service_time`: tempo médio de atendimento (em dias)
    - `status_stats`: estatísticas por status dos pedidos
    - `recent_requests`: últimos 5 pedidos relacionados ao serviço

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
  - Parâmetros de query (opcionais):
    - `search`: busca por ID, username, nome completo, email, tipo ou CPF
    - `type`: filtra por tipo de usuário (user, admin)
    - `page`: número da página (padrão: 1)
    - `limit`: itens por página (padrão: 10)
- `POST /api/users`
  - Cria um novo usuário (administrativo).
  - Payload JSON:
    ```json
    {
      "username": "usuario123",
      "email": "user@exemplo.com",
      "password": "***",
      "full_name": "Nome Completo",
      "cpf": "123.456.789-00",
      "birth_date": "01/01/1990",
      "type": "user",
      "team_id": 1,
      "profile_image_url": "https://.../foto.jpg"  // opcional
    }
    ```
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
      "email": "novo@email.com",
      "full_name": "Novo Nome Completo",
      "cpf": "111.222.333-44",
      "birth_date": "10/05/1995",
      "type": "admin",
      "team_id": 2,
      "profile_image_url": "https://.../nova-foto.jpg"
    }
    ```
- `DELETE /api/users/{id}`
  - Exclui usuário por `id`.
- `POST /api/users/{id}/profile-image`
  - Faz upload de imagem de perfil para o usuário.
  - Aceita `multipart/form-data`:
    - `image`: (file) Arquivo de imagem (jpg, jpeg, png, max 5MB).
  - Resposta:
    ```json
    {
      "url": "https://.../storage/v1/object/public/bucket/profile_images/123/uuid.jpg"
    }
    ```
- `DELETE /api/users/{id}/profile-image`
  - Remove a imagem de perfil do usuário.
  - Resposta:
    ```json
    {
      "message": "Profile image removed successfully"
    }
    ```


## Notícias (escrita)

- `POST /api/news`
  - Cria uma notícia.
  - Campos automáticos:
    - `author_id`: preenchido automaticamente com o ID do usuário autenticado
    - `slug`: gerado automaticamente a partir do título se não fornecido
    - `status`: padrão é "draft" se não fornecido
    - `published_at`: definido automaticamente quando status é "published"
  - Payload JSON:
    ```json
    {
      "title": "Título da notícia",
      "slug": "slug-da-noticia",
      "summary": "Resumo da notícia",
      "content": "Conteúdo formatado",
      "image_urls": ["https://exemplo.com/imagem1.jpg"],
      "status": "published",
      "category": "Categoria",
      "tags": ["tag1", "tag2"]
    }
    ```
- `POST /api/news/upload-image`
  - Faz upload de uma imagem para notícia.
  - Aceita `multipart/form-data`:
    - `image`: (file) Arquivo de imagem.
- `PUT /api/news/{id}`
  - Atualiza notícia existente por `id`.
  - Campo automático:
    - `published_at`: definido automaticamente quando status muda para "published"
  - Payload JSON (campos opcionais):
    ```json
    {
      "title": "Novo título",
      "content": "Novo conteúdo",
      "status": "published"
    }
    ```
- `DELETE /api/news/{id}`
  - Exclui notícia por `id`.

## Equipes (Teams)

- `GET /api/teams`
  - Lista todas as equipes.
- `POST /api/teams`
  - Cria uma nova equipe.
  - Payload JSON:
    ```json
    {
      "name": "Nome da Equipe",
      "service_category": "Categoria de Serviço",
      "description": "Descrição da equipe"
    }
    ```
- `GET /api/teams/{id}`
  - Retorna dados de uma equipe específica pelo `id`.
- `PUT /api/teams/{id}`
  - Atualiza equipe por `id`.
  - Payload JSON (campos opcionais):
    ```json
    {
      "name": "Novo Nome",
      "service_category": "Nova Categoria",
      "description": "Nova descrição"
    }
    ```
- `DELETE /api/teams/{id}`
  - Exclui equipe por `id`.

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
      "service_id": 1,
      "service_title": "Buraco na via",
      "request_data": {
        "address": "Av. Brasil, 450",
        "description": "Fisura no asfalto"
      }
    }
    ```
  - Também aceita `multipart/form-data` para anexos:
    - `service_id`: (number) ID do serviço
    - `service_title`: (string) Título do serviço
    - `request_data`: (string) JSON string com dados do formulário
    - `files`: (file[]) Arquivos de anexo
- `GET /api/service-requests`
  - Lista pedidos de serviço do usuário autenticado.
  - Parâmetros de query (opcionais):
    - `search`: busca por texto
    - `all`: quando `true`, lista todos os pedidos (modo administrativo)
    - `page`: número da página (padrão: 1)
    - `limit`: itens por página (padrão: 10)
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
