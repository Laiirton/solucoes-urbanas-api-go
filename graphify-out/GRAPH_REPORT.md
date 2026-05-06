# Graph Report - solucoes-urbanas-api-go  (2026-05-06)

## Corpus Check
- 60 files · ~24,657 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 399 nodes · 782 edges · 21 communities (19 shown, 2 thin omitted)
- Extraction: 60% EXTRACTED · 40% INFERRED · 0% AMBIGUOUS · INFERRED: 315 edges (avg confidence: 0.84)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `724d6f77`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]

## God Nodes (most connected - your core abstractions)
1. `respondError()` - 46 edges
2. `ServiceRequestRepository` - 28 edges
3. `AppConfigRepository` - 19 edges
4. `Setup()` - 19 edges
5. `NewsHandler` - 18 edges
6. `ServiceRequestHandler` - 18 edges
7. `ServiceRepository` - 18 edges
8. `NewsRepository` - 17 edges
9. `NewUploadService()` - 17 edges
10. `API surface` - 17 edges

## Surprising Connections (you probably didn't know these)
- `ENDPOINTS: Configuração do App (Mobile Home)` --references--> `AppConfigHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/app_config_handler.go
- `ENDPOINTS: Gestão de Configurações do App (Admin)` --references--> `AppConfigHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/app_config_handler.go
- `File uploads and storage` --conceptually_related_to--> `AppConfigHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/app_config_handler.go
- `ENDPOINTS: Autenticação pública` --references--> `AuthHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/auth_handler.go
- `ENDPOINTS: Notícias públicas` --references--> `NewsHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/news_handler.go

## Hyperedges (group relationships)
- **Service request flow** — concept_service_request_lifecycle, concept_geocoding_and_maps, concept_notifications_and_push_tokens, concept_service_catalog_and_ratings [INFERRED 0.86]
- **Media and upload workflow** — concept_file_uploads_and_storage, concept_profile_image_uploads, concept_news_publishing_pipeline [INFERRED 0.84]

## Communities (21 total, 2 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.07
Nodes (14): parseID(), parsePagination(), respondError(), respondJSON(), hasSystemNotificationUpdateFields(), NewNotificationHandler(), NotificationHandler, NewServiceHandler() (+6 more)

### Community 1 - "Community 1"
Cohesion: 0.09
Nodes (27): failingDeleteMock, failingMockStorage, FileUploadError, mockStorageService, mockUploadedFile, supabaseStorageService, NewUploadService(), ParseAttachmentURLs() (+19 more)

### Community 2 - "Community 2"
Cohesion: 0.09
Nodes (15): extractAddressFromRequestData(), NewServiceRequestHandler(), ServiceRequestHandler, CreateServiceRequest, CreateServiceRequestRequest, Service, GetServiceIcon(), ServiceDetailResponse (+7 more)

### Community 3 - "Community 3"
Cohesion: 0.11
Nodes (11): extractSupabaseURLs(), generateSlug(), hasNewsUpdateFields(), NewNewsHandler(), NewsHandler, SystemNotification, NewNewsRepository(), nullableValue() (+3 more)

### Community 4 - "Community 4"
Cohesion: 0.13
Nodes (6): App configuration and banners, Mobile home configuration, NewAppConfigHandler(), AppConfigHandler, NewAppConfigRepository(), AppConfigRepository

### Community 5 - "Community 5"
Cohesion: 0.13
Nodes (13): Authentication and JWT, Geocoding and maps, NewAuthHandler(), AuthHandler, NewGeolocationHandler(), GeolocationHandler, Auth(), respondJSON() (+5 more)

### Community 6 - "Community 6"
Cohesion: 0.11
Nodes (9): main(), Database bootstrap, Deployment topology, Config, Load(), Connect(), DB, NewUserRepository() (+1 more)

### Community 7 - "Community 7"
Cohesion: 0.25
Nodes (17): API surface, ENDPOINTS: Autenticação pública, ENDPOINTS: Avaliações de Serviços, ENDPOINTS: Configuração do App (Mobile Home), ENDPOINTS: Equipes (Teams), ENDPOINTS: Geolocalização, ENDPOINTS: Gestão de Configurações do App (Admin), ENDPOINTS: Health Check (+9 more)

### Community 8 - "Community 8"
Cohesion: 0.1
Nodes (19): File uploads and storage, NewUserHandler(), PROFILE: 1. Banco de Dados (Migração), PROFILE: 2. Model (User), PROFILE: 3. Repositório, PROFILE: 4. Handler, PROFILE: 5. Rotas, PROFILE: Armazenamento (+11 more)

### Community 9 - "Community 9"
Cohesion: 0.13
Nodes (8): Service catalog and ratings, Service request lifecycle, NewServiceAttendanceHandler(), ServiceAttendanceHandler, NewServiceAttendanceRepository(), NewServiceRatingRepository(), ServiceAttendanceRepository, ServiceRatingRepository

### Community 10 - "Community 10"
Cohesion: 0.17
Nodes (8): News publishing pipeline, Notifications and push tokens, NewPushTokenRepository(), PushTokenRepository, chunkStrings(), NewExpoPushService(), ExpoPushMessage, ExpoPushService

### Community 11 - "Community 11"
Cohesion: 0.13
Nodes (11): CreateSystemNotificationRequest, CreateUserRequest, ErrorResponse, LoginRequest, LoginResponse, MessageResponse, RegisterPushTokenRequest, UpdateSystemNotificationRequest (+3 more)

### Community 12 - "Community 12"
Cohesion: 0.14
Nodes (12): NewHomeHandler(), HomeHandler, CategoryStat, HomeAlert, HomeResponse, HomeStats, MapLocation, PopularService (+4 more)

### Community 13 - "Community 13"
Cohesion: 0.2
Nodes (4): Teams and categories, NewTeamHandler(), NewTeamRepository(), TeamRepository

### Community 14 - "Community 14"
Cohesion: 0.31
Nodes (6): NewServiceRatingHandler(), ServiceRatingHandler, CreateServiceRatingRequest, ServiceRating, ServiceRatingResponse, ServiceRatingStats

### Community 15 - "Community 15"
Cohesion: 0.43
Nodes (7): NewSupabaseStorageService(), TestSupabaseStorageService_DeleteFile_ServerError(), TestSupabaseStorageService_DeleteFile_Success(), TestSupabaseStorageService_DeleteFile_WrongBucket(), TestSupabaseStorageService_UploadFile_ServerError(), TestSupabaseStorageService_UploadFile_Success(), TestSupabaseStorageService_UploadFile_UsesReader()

### Community 16 - "Community 16"
Cohesion: 0.29
Nodes (6): AppBanner, AppConfig, CategorySummary, MobileHomeResponse, Section, ServiceSummary

### Community 17 - "Community 17"
Cohesion: 0.5
Nodes (3): CreateTeamRequest, Team, UpdateTeamRequest

## Knowledge Gaps
- **52 isolated node(s):** `Config`, `contextKey`, `AppBanner`, `AppConfig`, `ServiceSummary` (+47 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `respondError()` connect `Community 0` to `Community 1`, `Community 2`, `Community 5`, `Community 9`, `Community 14`?**
  _High betweenness centrality (0.139) - this node is a cross-community bridge._
- **Why does `main()` connect `Community 6` to `Community 0`, `Community 2`, `Community 3`, `Community 4`, `Community 5`, `Community 7`, `Community 9`, `Community 10`, `Community 13`, `Community 15`?**
  _High betweenness centrality (0.126) - this node is a cross-community bridge._
- **Why does `Setup()` connect `Community 5` to `Community 0`, `Community 1`, `Community 2`, `Community 3`, `Community 4`, `Community 6`, `Community 7`, `Community 8`, `Community 9`, `Community 10`, `Community 12`, `Community 13`, `Community 14`?**
  _High betweenness centrality (0.124) - this node is a cross-community bridge._
- **Are the 43 inferred relationships involving `respondError()` (e.g. with `.Login()` and `.ListNews()`) actually correct?**
  _`respondError()` has 43 INFERRED edges - model-reasoned connections that need verification._
- **Are the 7 inferred relationships involving `ServiceRequestRepository` (e.g. with `ENDPOINTS: Usuários` and `ENDPOINTS: Serviços (escrita)`) actually correct?**
  _`ServiceRequestRepository` has 7 INFERRED edges - model-reasoned connections that need verification._
- **Are the 4 inferred relationships involving `AppConfigRepository` (e.g. with `ENDPOINTS: Configuração do App (Mobile Home)` and `ENDPOINTS: Gestão de Configurações do App (Admin)`) actually correct?**
  _`AppConfigRepository` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Are the 18 inferred relationships involving `Setup()` (e.g. with `main()` and `NewAuthHandler()`) actually correct?**
  _`Setup()` has 18 INFERRED edges - model-reasoned connections that need verification._