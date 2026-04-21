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
  - Retorna uma notícia específica pelo `id` ou `slug`.

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
      "full_name": "Novo Nome Completo",
      "cpf": "111.222.333-44",
      "birth_date": "10/05/1995",
      "type": "admin",
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
