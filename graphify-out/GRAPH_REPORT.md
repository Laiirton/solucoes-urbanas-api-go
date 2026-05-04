# Graph Report - E:\git\solucoes-urbanas-api-go  (2026-05-04)

## Corpus Check
- Corpus is ~24,643 words - fits in a single context window. You may not need a graph.

## Summary
- 314 nodes · 624 edges · 18 communities (16 shown, 2 thin omitted)
- Extraction: 53% EXTRACTED · 47% INFERRED · 0% AMBIGUOUS · INFERRED: 295 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b0abfc2a`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_respondError()|respondError()]]
- [[_COMMUNITY_NewUploadService()|NewUploadService()]]
- [[_COMMUNITY_Setup()|Setup()]]
- [[_COMMUNITY_NewsHandler|NewsHandler]]
- [[_COMMUNITY_ServiceRequestRepository|ServiceRequestRepository]]
- [[_COMMUNITY_Service request lifecycle|Service request lifecycle]]
- [[_COMMUNITY_main()|main()]]
- [[_COMMUNITY_API surface|API surface]]
- [[_COMMUNITY_File uploads and storage|File uploads and storage]]
- [[_COMMUNITY_AppConfigHandler|AppConfigHandler]]
- [[_COMMUNITY_CreateSystemNotificationRequest|CreateSystemNotificationRequest]]
- [[_COMMUNITY_ServiceHandler|ServiceHandler]]
- [[_COMMUNITY_AppBanner|AppBanner]]
- [[_COMMUNITY_ExpoPushService|ExpoPushService]]
- [[_COMMUNITY_CreateTeamRequest|CreateTeamRequest]]
- [[_COMMUNITY_News|News]]
- [[_COMMUNITY_CreateServiceAttendanceRequest|CreateServiceAttendanceRequest]]

## God Nodes (most connected - your core abstractions)
1. `respondError()` - 42 edges
2. `ServiceRequestRepository` - 20 edges
3. `Setup()` - 18 edges
4. `API surface` - 17 edges
5. `NewsHandler` - 16 edges
6. `SystemNotification` - 16 edges
7. `main()` - 15 edges
8. `ServiceRequestHandler` - 15 edges
9. `AppConfigHandler` - 14 edges
10. `AppConfigRepository` - 14 edges

## Surprising Connections (you probably didn't know these)
- `ENDPOINTS: Configuração do App (Mobile Home)` --references--> `AppConfigHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/app_config_handler.go
- `ENDPOINTS: Gestão de Configurações do App (Admin)` --references--> `AppConfigHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/app_config_handler.go
- `File uploads and storage` --conceptually_related_to--> `AppConfigHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/app_config_handler.go
- `ENDPOINTS: Autenticação pública` --references--> `AuthHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/auth_handler.go
- `Authentication and JWT` --conceptually_related_to--> `AuthHandler`  [INFERRED]
  ENDPOINTS.md → internal/handlers/auth_handler.go

## Hyperedges (group relationships)
- **Service request flow** — concept_service_request_lifecycle, concept_geocoding_and_maps, concept_notifications_and_push_tokens, concept_service_catalog_and_ratings [INFERRED 0.86]
- **Media and upload workflow** — concept_file_uploads_and_storage, concept_profile_image_uploads, concept_news_publishing_pipeline [INFERRED 0.84]

## Communities (18 total, 2 thin omitted)

### Community 0 - "respondError()"
Cohesion: 0.09
Nodes (17): Authentication and JWT, Teams and categories, parseID(), parsePagination(), respondError(), hasSystemNotificationUpdateFields(), NewNotificationHandler(), NotificationHandler (+9 more)

### Community 1 - "NewUploadService()"
Cohesion: 0.13
Nodes (24): failingDeleteMock, failingMockStorage, FileUploadError, mockStorageService, mockUploadedFile, supabaseStorageService, NewUploadService(), ParseAttachmentURLs() (+16 more)

### Community 2 - "Setup()"
Cohesion: 0.08
Nodes (22): Geocoding and maps, NewAppConfigHandler(), NewAuthHandler(), AuthHandler, GeolocationHandler, NewHomeHandler(), NewNewsHandler(), NewServiceRequestHandler() (+14 more)

### Community 3 - "NewsHandler"
Cohesion: 0.12
Nodes (11): News publishing pipeline, Notifications and push tokens, extractSupabaseURLs(), generateSlug(), hasNewsUpdateFields(), NewsHandler, SystemNotification, nullableValue() (+3 more)

### Community 4 - "ServiceRequestRepository"
Cohesion: 0.13
Nodes (11): HomeHandler, extractAddressFromRequestData(), ServiceRequestHandler, CreateServiceRequest, Service, ServiceDetailResponse, ServiceRequest, ServiceRequestDetailResponse (+3 more)

### Community 5 - "Service request lifecycle"
Cohesion: 0.09
Nodes (13): Service catalog and ratings, Service request lifecycle, NewServiceAttendanceHandler(), NewServiceRatingHandler(), ServiceAttendanceHandler, CreateServiceRatingRequest, ServiceRating, ServiceRatingResponse (+5 more)

### Community 6 - "main()"
Cohesion: 0.12
Nodes (16): main(), Database bootstrap, Deployment topology, Config, Load(), Connect(), DB, NewAppConfigRepository() (+8 more)

### Community 7 - "API surface"
Cohesion: 0.25
Nodes (17): API surface, ENDPOINTS: Autenticação pública, ENDPOINTS: Avaliações de Serviços, ENDPOINTS: Configuração do App (Mobile Home), ENDPOINTS: Equipes (Teams), ENDPOINTS: Geolocalização, ENDPOINTS: Gestão de Configurações do App (Admin), ENDPOINTS: Health Check (+9 more)

### Community 8 - "File uploads and storage"
Cohesion: 0.1
Nodes (19): File uploads and storage, NewUserHandler(), PROFILE: 1. Banco de Dados (Migração), PROFILE: 2. Model (User), PROFILE: 3. Repositório, PROFILE: 4. Handler, PROFILE: 5. Rotas, PROFILE: Armazenamento (+11 more)

### Community 9 - "AppConfigHandler"
Cohesion: 0.22
Nodes (4): App configuration and banners, Mobile home configuration, AppConfigHandler, AppConfigRepository

### Community 10 - "CreateSystemNotificationRequest"
Cohesion: 0.15
Nodes (11): CreateSystemNotificationRequest, CreateUserRequest, ErrorResponse, LoginRequest, LoginResponse, MessageResponse, RegisterPushTokenRequest, UpdateSystemNotificationRequest (+3 more)

### Community 11 - "ServiceHandler"
Cohesion: 0.24
Nodes (5): NewServiceHandler(), ServiceHandler, GetCategoryIcon(), NewServiceRepository(), ServiceRepository

### Community 12 - "AppBanner"
Cohesion: 0.29
Nodes (6): AppBanner, AppConfig, CategorySummary, MobileHomeResponse, Section, ServiceSummary

### Community 13 - "ExpoPushService"
Cohesion: 0.43
Nodes (4): chunkStrings(), NewExpoPushService(), ExpoPushMessage, ExpoPushService

### Community 14 - "CreateTeamRequest"
Cohesion: 0.5
Nodes (3): CreateTeamRequest, Team, UpdateTeamRequest

## Knowledge Gaps
- **60 isolated node(s):** `Config`, `contextKey`, `AppBanner`, `AppConfig`, `ServiceSummary` (+55 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `respondError()` connect `respondError()` to `NewUploadService()`, `Setup()`, `NewsHandler`, `ServiceRequestRepository`, `Service request lifecycle`, `ServiceHandler`?**
  _High betweenness centrality (0.137) - this node is a cross-community bridge._
- **Why does `Setup()` connect `Setup()` to `respondError()`, `NewUploadService()`, `Service request lifecycle`, `main()`, `API surface`, `File uploads and storage`, `ServiceHandler`, `ExpoPushService`?**
  _High betweenness centrality (0.119) - this node is a cross-community bridge._
- **Why does `main()` connect `main()` to `respondError()`, `Setup()`, `NewsHandler`, `Service request lifecycle`, `API surface`, `ServiceHandler`?**
  _High betweenness centrality (0.107) - this node is a cross-community bridge._
- **Are the 40 inferred relationships involving `respondError()` (e.g. with `.Login()` and `.ListNews()`) actually correct?**
  _`respondError()` has 40 INFERRED edges - model-reasoned connections that need verification._
- **Are the 7 inferred relationships involving `ServiceRequestRepository` (e.g. with `ENDPOINTS: Usuários` and `ENDPOINTS: Serviços (escrita)`) actually correct?**
  _`ServiceRequestRepository` has 7 INFERRED edges - model-reasoned connections that need verification._
- **Are the 17 inferred relationships involving `Setup()` (e.g. with `main()` and `NewAuthHandler()`) actually correct?**
  _`Setup()` has 17 INFERRED edges - model-reasoned connections that need verification._
- **Are the 17 inferred relationships involving `API surface` (e.g. with `ENDPOINTS: Health Check` and `ENDPOINTS: Autenticação pública`) actually correct?**
  _`API surface` has 17 INFERRED edges - model-reasoned connections that need verification._