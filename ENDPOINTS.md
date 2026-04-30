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

- `POST /api/auth/login`
  - Faz login do usuário.
  - Aceita username ou email no campo `username`.
  - Payload JSON:
    ```json
    {
      "username": "usuario123",
      "password": "senha_forte"
    }
    ```
  - Exemplo com email:
    ```json
    {
      "username": "user@exemplo.com",
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
- `GET /api/services/categories`
  - Lista todas as categorias de serviços disponíveis.
  - Parâmetros de query (opcionais):
    - `all`: quando `true`, lista categorias de todos os serviços (incluindo inativos)
  - Resposta JSON:
    ```json
    ["Limpeza Urbana", "Saúde", "Educação", "Segurança Pública"]
    ```
- `GET /api/services/category/{category}`
  - Retorna todos os serviços pertencentes a uma categoria específica.
  - Parâmetros de query (opcionais):
    - `all`: quando `true`, lista todos os serviços (incluindo inativos)
- `GET /api/services/{id}`
  - Retorna detalhes de um serviço específico pelo `id`.
  - Inclui estatísticas adicionais:
    - `average_service_time`: tempo médio de atendimento (em dias)
    - `rating_stats`: média de estrelas e total de avaliações
    - `status_stats`: estatísticas por status dos pedidos
    - `recent_requests`: últimos 5 pedidos relacionados ao serviço

## Configuração do App (Mobile Home)

- `GET /api/app/config`
  - Retorna a configuração completa para a Home do aplicativo mobile.
  - Estrutura modular baseada em `sections` para permitir reordenação dinâmica.
  - Resposta JSON:
    ```json
    {
      "logo_url": "https://...",
      "banners": [
        { "id": 1, "image_url": "...", "title": "...", "link_url": "...", "order_index": 0 }
      ],
      "sections": [
        { "type": "categories", "title": "Categorias", "data": [...] },
        { "type": "services", "title": "Serviços em Destaque", "data": [...] },
        { "type": "top_rated", "title": "Melhores Avaliados", "data": [...] }
      ]
    }
    ```

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
  - Quando a notícia já nasce ou passa para `published`, a API envia push notifications para os usuários cadastrados.
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
  - Campos ausentes ficam inalterados.
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

## Notificações

Todos endpoints dessa seção exigem autenticação JWT.

- `POST /api/notifications/push-tokens`
  - Registra ou atualiza o `ExponentPushToken[...]` do dispositivo do usuário autenticado.
  - Payload JSON:
    ```json
    {
      "token": "ExponentPushToken[...]"
    }
    ```
  - Esse token é usado para receber notificações automáticas quando notícias forem publicadas ou houver atualização em pedidos de serviço.

- `GET /api/notifications`
  - Lista notificações do usuário logado (somente as que ele tem permissão)
  - Parâmetros de query (opcionais):
    - `type`: filtra por tipo de notificação
    - `unread_only`: `true` para retornar somente notificações não lidas
    - `page`: número da página (padrão: 1)
    - `limit`: itens por página (padrão: 20)

- `GET /api/notifications/{id}`
  - Retorna detalhes de uma notificação específica
  - Retorna `403 Forbidden` caso usuário não tenha permissão para visualizar
  - Retorna `404 Not Found` caso notificação não exista

- `PATCH /api/notifications/{id}/read`
  - Marca notificação como lida, define a data de visualização
  - Retorna `403 Forbidden` caso usuário não seja dono da notificação

- `DELETE /api/notifications/{id}`
  - Remove notificação
  - Retorna `403 Forbidden` caso usuário não seja dono da notificação
  - Retorna status `204 No Content` em caso de sucesso

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

## Gestão de Configurações do App (Admin)

Estes endpoints exigem autenticação administrativa.

- `PUT /api/app/settings/{key}`
  - Atualiza uma configuração global do app.
  - Chaves suportadas: `logo_url`, `featured_services` (array de IDs), `featured_categories` (array de nomes).
  - Exemplo para `featured_services`:
    ```json
    [1, 5, 10]
    ```

- `POST /api/app/upload-image`
  - Faz upload de uma imagem para ser usada no logo ou banners.
  - Aceita `multipart/form-data`:
    - `image`: (file) Arquivo de imagem.
  - Retorna o URL público da imagem:
    ```json
    { "url": "https://..." }
    ```

- `GET /api/app/banners`
  - Lista todos os banners (incluindo inativos) para gestão.

- `POST /api/app/banners`
  - Cria um novo banner para o carrossel.
  - Payload JSON:
    ```json
    {
      "image_url": "https://...",
      "title": "Título Opcional",
      "link_url": "https://...",
      "order_index": 0,
      "is_active": true
    }
    ```

- `PUT /api/app/banners/{id}`
  - Atualiza um banner existente.

- `DELETE /api/app/banners/{id}`
  - Remove um banner permanentemente.

## Avaliações de Serviços

- `POST /api/ratings` (Protegido)
  - Avalia um pedido de serviço concluído (1-5 estrelas).
  - Um pedido só pode ser avaliado uma vez.
  - Payload JSON:
    ```json
    {
      "service_request_id": 123,
      "stars": 5,
      "comment": "Excelente atendimento!"
    }
    ```
- `GET /api/services/{id}/ratings` (Público)
  - Lista as avaliações de um serviço específico.
  - Parâmetros de query: `page`, `limit`.
- `GET /api/services/{id}/rating-stats` (Público)
  - Retorna a média de estrelas e o total de avaliações de um serviço.
  - Resposta JSON:
    ```json
    {
      "average": 4.5,
      "count": 120
    }
    ```
