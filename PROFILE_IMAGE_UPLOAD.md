# Upload de Imagem de Perfil - Implementação

## Visão Geral
Implementado sistema completo de upload de imagem de perfil para usuários da API.

## Mudanças Realizadas

### 1. Banco de Dados (Migração)

**Arquivo:** `internal/database/migrations/000007_add_profile_image_to_users.up.sql`
- Adiciona coluna `profile_image_url` VARCHAR NULL na tabela `users`
- Cria índice para otimizar consultas

**Arquivo:** `internal/database/migrations/000007_add_profile_image_to_users.down.sql`
- Rollback da migração (remove coluna e índice)

### 2. Model (User)

**Arquivo:** `internal/models/user.go`
- Adicionado campo `ProfileImageURL *string` no struct `User`
- Adicionado campo `ProfileImageURL *string` no struct `CreateUserRequest`
- Adicionado campo `ProfileImageURL *string` no struct `UpdateUserRequest`

### 3. Repositório

**Arquivo:** `internal/repository/user_repository.go`
- Atualizado `CreateUser` para incluir `profile_image_url`
- Atualizado `GetUserByUsername` para retornar `profile_image_url`
- Atualizado `GetUserByID` para retornar `profile_image_url`
- Atualizado `ListUsers` para retornar `profile_image_url`
- Atualizado `UpdateUser` para permitir update de `profile_image_url`

### 4. Handler

**Arquivo:** `internal/handlers/user_handler.go`
- Adicionado `storage services.StorageService` ao UserHandler
- Criado endpoint `POST /users/{id}/profile-image` - Upload de imagem de perfil
- Criado endpoint `DELETE /users/{id}/profile-image` - Remover imagem de perfil
- Validações implementadas:
  - Apenas imagens (jpg, jpeg, png, webp, gif)
  - Tamanho máximo: 5MB
  - Usuário só pode editar sua própria imagem (ou admin)
  - Rollback em caso de falha na atualização do banco

### 5. Rotas

**Arquivo:** `internal/routes/routes.go`
- Adicionada rota POST `/api/users/{id}/profile-image`
- Adicionada rota DELETE `/api/users/{id}/profile-image`
- Inicializado UserHandler com storageService

## Como Usar

### Upload de Imagem de Perfil

```bash
POST /api/users/{id}/profile-image
Content-Type: multipart/form-data

Parâmetros:
- image: arquivo da imagem (jpg, jpeg, png, webp, gif, max 5MB)

Resposta (sucesso):
{
  "url": "https://.../storage/v1/object/public/bucket/profile_images/123/uuid.jpg"
}
```

### Remover Imagem de Perfil

```bash
DELETE /api/users/{id}/profile-image

Resposta (sucesso):
{
  "message": "Profile image removed successfully"
}
```

### Criar Usuário com Imagem de Perfil

```bash
POST /api/users
Content-Type: application/json

{
  "username": "joao",
  "password": "senha123",
  "email": "joao@example.com",
  "full_name": "João Silva",
  "cpf": "123.456.789-00",
  "birth_date": "01/01/2000",
  "type": "user",
  "profile_image_url": "https://.../photo.jpg" // opcional
}
```

### Atualizar Imagem de Perfil via PATCH/PUT

```bash
PUT /api/users/{id}
Content-Type: application/json

{
  "profile_image_url": "https://.../nova-foto.jpg"
}
```

## Armazenamento

As imagens de perfil são armazenadas em:
- **Path:** `profile_images/{userID}/{uuid}.{ext}`
- **Bucket:** Mesmo bucket utilizado para news e service_requests
- **Acesso:** Público via URL

## Segurança

- **Autenticação:** Requer token JWT válido
- **Autorização:** 
  - Usuários podem editar apenas sua própria imagem
  - Admins podem editar qualquer imagem
- **Validação de Tipo:** Apenas imagens (jpg, jpeg, png, webp, gif)
- **Tamanho Máximo:** 5MB por arquivo
- **Rollback:** Em caso de falha, a imagem é removida do storage

## Estrutura do Sistema de Upload

````
┌─────────────────────────────────────────────────────────────┐
│                    UserHandler                               │
│  ┌────────────────┐  ┌────────────────┐                     │
│  │ UploadProfile  │  │ DeleteProfile  │                     │
│  │   Image()      │  │   Image()      │                     │
│  └───────┬────────┘  └───────┬────────┘                     │
│          │                   │                               │
│          ▼                   ▼                               │
│  ┌────────────────────────────────────────┐                 │
│  │      Validações                        │                 │
│  │  - Auth (JWT)                          │                 │
│  │  - Permissão (user/admin)              │                 │
│  │  - Tipo de arquivo                     │                 │
│  │  - Tamanho (max 5MB)                   │                 │
│  └────────────────────────────────────────┘                 │
│          │                                                   │
│          ▼                                                   │
│  ┌──────────────────┐     ┌──────────────────────┐         │
│  │  StorageService  │     │  UserRepository      │         │
│  │  - UploadFile    │     │  - UpdateUser        │         │
│  │  - DeleteFile    │     │                      │         │
│  └──────────────────┘     └──────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
````

## Próximos Passos Sugeridos

1. **Executar migração** no banco de dados:
   ```bash
   # Exemplo com golang-migrate
   migrate -path internal/database/migrations -database "postgres://..." up
   ```

2. **Testar endpoints** com ferramentas como Postman/curl

3. **Adicionar testes unitários** para os novos handlers

4. **Configurar policies** no Supabase Storage para a pasta `profile_images/`

## Endpoints da API

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/api/users/{id}/profile-image` | Upload de imagem de perfil |
| DELETE | `/api/users/{id}/profile-image` | Remover imagem de perfil |
| PUT | `/api/users/{id}` | Atualizar dados do usuário (incluindo profile_image_url) |
| GET | `/api/users/{id}` | Buscar dados do usuário (retorna profile_image_url) |
| GET | `/api/users/me` | Buscar dados do usuário autenticado (retorna profile_image_url) |
