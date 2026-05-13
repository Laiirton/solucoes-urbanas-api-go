# Graph Report - .  (2026-05-13)

## Corpus Check
- Corpus is ~31,568 words - fits in a single context window. You may not need a graph.

## Summary
- 1043 nodes · 1687 edges · 68 communities (56 shown, 12 thin omitted)
- Extraction: 36% EXTRACTED · 64% INFERRED · 0% AMBIGUOUS · INFERRED: 1086 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Helpers Layer|Helpers Layer]]
- [[_COMMUNITY_Service Request & Auth Handlers|Service Request & Auth Handlers]]
- [[_COMMUNITY_Push Token Repository|Push Token Repository]]
- [[_COMMUNITY_AppConfig & Stats Repository|AppConfig & Stats Repository]]
- [[_COMMUNITY_Auth & Geolocation Handlers|Auth & Geolocation Handlers]]
- [[_COMMUNITY_API Endpoints Documentation|API Endpoints Documentation]]
- [[_COMMUNITY_Upload Service & Tests|Upload Service & Tests]]
- [[_COMMUNITY_Home & Region Filter Handlers|Home & Region Filter Handlers]]
- [[_COMMUNITY_News & Notification Repositories|News & Notification Repositories]]
- [[_COMMUNITY_App Config Handler|App Config Handler]]
- [[_COMMUNITY_Service Request Repository|Service Request Repository]]
- [[_COMMUNITY_News & App Config Handlers|News & App Config Handlers]]
- [[_COMMUNITY_Notification Handler|Notification Handler]]
- [[_COMMUNITY_Service Rating & List Handlers|Service Rating & List Handlers]]
- [[_COMMUNITY_Storage Service & Tests|Storage Service & Tests]]
- [[_COMMUNITY_Category & Service Handlers|Category & Service Handlers]]
- [[_COMMUNITY_Core Bootstrap|Core Bootstrap]]
- [[_COMMUNITY_Team Handler & Models|Team Handler & Models]]
- [[_COMMUNITY_User Handler & Models|User Handler & Models]]
- [[_COMMUNITY_Team Repository|Team Repository]]
- [[_COMMUNITY_Expo Push Service & Tests|Expo Push Service & Tests]]
- [[_COMMUNITY_System Notification Repository|System Notification Repository]]
- [[_COMMUNITY_Region Repository|Region Repository]]
- [[_COMMUNITY_Handler Injection Hub|Handler Injection Hub]]
- [[_COMMUNITY_Service Repository|Service Repository]]
- [[_COMMUNITY_Team Repository (alt)|Team Repository (alt)]]
- [[_COMMUNITY_Region Repository (alt)|Region Repository (alt)]]
- [[_COMMUNITY_User Repository|User Repository]]
- [[_COMMUNITY_Home Models (AST)|Home Models (AST)]]
- [[_COMMUNITY_Home Models (Semantic)|Home Models (Semantic)]]
- [[_COMMUNITY_Service Handler|Service Handler]]
- [[_COMMUNITY_Upload Service Tests|Upload Service Tests]]
- [[_COMMUNITY_Team Handler|Team Handler]]
- [[_COMMUNITY_User Models|User Models]]
- [[_COMMUNITY_Service Repository (alt)|Service Repository (alt)]]
- [[_COMMUNITY_Service Rating Handler|Service Rating Handler]]
- [[_COMMUNITY_Storage Service Tests|Storage Service Tests]]
- [[_COMMUNITY_Category Repository|Category Repository]]
- [[_COMMUNITY_User Repository (alt)|User Repository (alt)]]
- [[_COMMUNITY_News Repository|News Repository]]
- [[_COMMUNITY_Expo Push Service|Expo Push Service]]
- [[_COMMUNITY_Core Bootstrap (alt)|Core Bootstrap (alt)]]
- [[_COMMUNITY_Service Rating Repository|Service Rating Repository]]
- [[_COMMUNITY_Service Attendance Repository (alt)|Service Attendance Repository (alt)]]
- [[_COMMUNITY_Service Attendance Repository|Service Attendance Repository]]
- [[_COMMUNITY_App Config Models|App Config Models]]
- [[_COMMUNITY_Notification Models|Notification Models]]
- [[_COMMUNITY_Team Models|Team Models]]
- [[_COMMUNITY_Category Repository (alt)|Category Repository (alt)]]
- [[_COMMUNITY_Upload & Attendance Tests|Upload & Attendance Tests]]
- [[_COMMUNITY_Service Models (AST)|Service Models (AST)]]
- [[_COMMUNITY_Middleware Auth Layer|Middleware Auth Layer]]
- [[_COMMUNITY_Service Rating Repository (alt)|Service Rating Repository (alt)]]
- [[_COMMUNITY_Service Models (Semantic)|Service Models (Semantic)]]
- [[_COMMUNITY_Geocoding & Home Service|Geocoding & Home Service]]
- [[_COMMUNITY_Upload Attachment Tests|Upload Attachment Tests]]
- [[_COMMUNITY_Database Connection|Database Connection]]
- [[_COMMUNITY_Category Models|Category Models]]
- [[_COMMUNITY_Service Rating Models|Service Rating Models]]
- [[_COMMUNITY_Service Request Models|Service Request Models]]
- [[_COMMUNITY_Category Models (Semantic)|Category Models (Semantic)]]
- [[_COMMUNITY_Region Models|Region Models]]
- [[_COMMUNITY_Service Icons Models|Service Icons Models]]
- [[_COMMUNITY_News Models (AST)|News Models (AST)]]
- [[_COMMUNITY_Service Attendance Models|Service Attendance Models]]
- [[_COMMUNITY_Request Stats Repository|Request Stats Repository]]
- [[_COMMUNITY_News Models (Semantic)|News Models (Semantic)]]
- [[_COMMUNITY_Docker Infrastructure|Docker Infrastructure]]

## God Nodes (most connected - your core abstractions)
1. `respondError()` - 63 edges
2. `ServiceRequestRepository` - 30 edges
3. `respondJSON()` - 27 edges
4. `respondError()` - 27 edges
5. `ServiceRequestRepository` - 25 edges
6. `Database Package` - 25 edges
7. `ServiceRequestHandler` - 24 edges
8. `Setup()` - 22 edges
9. `Main Function` - 19 edges
10. `ServiceHandler` - 19 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Connect()`  [INFERRED]
  cmd/api/main.go → internal/database/database.go
- `main()` --calls--> `Load()`  [INFERRED]
  cmd/api/main.go → internal/config/config.go
- `main()` --calls--> `NewUserRepository()`  [INFERRED]
  cmd/api/main.go → internal/repository/user_repository.go
- `main()` --calls--> `NewServiceRepository()`  [INFERRED]
  cmd/api/main.go → internal/repository/service_repository.go
- `main()` --calls--> `NewNewsRepository()`  [INFERRED]
  cmd/api/main.go → internal/repository/news_repository.go

## Communities (68 total, 12 thin omitted)

### Community 0 - "Helpers Layer"
Cohesion: 0.06
Nodes (20): Handler Helpers, Handler Helpers Tests, TestRespondError(), TestRespondJSON(), CategoryHandler, parseID(), parsePagination(), respondError() (+12 more)

### Community 1 - "Service Request & Auth Handlers"
Cohesion: 0.06
Nodes (68): AuthHandler.Login(), AuthHandler.Logout(), CanManageTeam(), CategoryHandler.CreateCategory(), CategoryHandler.DeleteCategory(), CategoryHandler.GetCategory(), CategoryHandler.ListCategories(), CategoryHandler.UpdateCategory() (+60 more)

### Community 2 - "Push Token Repository"
Cohesion: 0.08
Nodes (54): NewPushTokenRepository, PushTokenRepository, PushTokenRepository.DeletePushToken, PushTokenRepository.ListTokens, PushTokenRepository.ListTokensByUser, PushTokenRepository.UpsertPushToken, Database Package, Migration 000017: Fix Region System (down) (+46 more)

### Community 3 - "AppConfig & Stats Repository"
Cohesion: 0.06
Nodes (8): AppConfigRepository, GetCategoryIcon(), GetServiceIcon(), NewAppConfigRepository(), AppConfigRepository, buildBaseWhere(), calcPercentFunc(), ServiceRequestRepository

### Community 4 - "Auth & Geolocation Handlers"
Cohesion: 0.06
Nodes (27): auth_handler.go, Auth Handler, Geolocation Handler, Region Handler, Service Attendance Handler, GET /api/geolocation — Address/geolocation lookup, Geolocation Endpoints, NewAppConfigHandler() (+19 more)

### Community 5 - "API Endpoints Documentation"
Cohesion: 0.09
Nodes (35): API Root — Soluções Urbanas, GET /api/app/config — Mobile home config, App Config Endpoints, app_config_handler.go, models/app_config.go, POST /api/app/banners — Create banner, DELETE /api/app/banners/{id} — Delete banner, GET /api/app/banners — List all banners (admin) (+27 more)

### Community 6 - "Upload Service & Tests"
Cohesion: 0.19
Nodes (24): failingDeleteMock, mockUploadedFile, NewUploadService(), ParseAttachmentURLs(), createMultipartFiles(), TestFileUploadError_Format(), TestParseAttachmentURLs_Empty(), TestParseAttachmentURLs_EmptyArray() (+16 more)

### Community 7 - "Home & Region Filter Handlers"
Cohesion: 0.12
Nodes (15): Home Handler, Region Filter, NewHomeHandler(), HomeHandler, CanManageTeam(), GetRegionFilterForAdmin(), GetRegionFilterForUser(), GetTeamFilterForUser() (+7 more)

### Community 8 - "News & Notification Repositories"
Cohesion: 0.07
Nodes (7): NewNewsRepository(), nullableValue(), NewsRepository, NewPushTokenRepository(), PushTokenRepository, NewSystemNotificationRepository(), SystemNotificationRepository

### Community 9 - "App Config Handler"
Cohesion: 0.13
Nodes (28): AppConfigHandler.CreateBanner(), AppConfigHandler.DeleteBanner(), AppConfigHandler.GetMobileConfig(), AppConfigHandler.ListBanners(), AppConfigHandler.UpdateBanner(), AppConfigHandler.UpdateSetting(), AppConfigHandler.UploadImage(), AppConfigHandler.deleteFileIfInternal() (+20 more)

### Community 10 - "Service Request Repository"
Cohesion: 0.07
Nodes (29): NewServiceRequestRepository, NewServiceRequestStatsRepository, ServiceRequestRepository, ServiceRequestRepository.CountServiceRequestsByStatusByUser, ServiceRequestRepository.CountServiceRequestsByUser, ServiceRequestRepository.CreateServiceRequest, ServiceRequestRepository.DeleteServiceRequest, ServiceRequestRepository.GetAverageServiceTime (+21 more)

### Community 11 - "News & App Config Handlers"
Cohesion: 0.15
Nodes (8): News Handler, AppConfigHandler, extractSupabaseURLs(), generateSlug(), hasNewsUpdateFields(), NewNewsHandler(), NewsHandler, FileUploadError

### Community 12 - "Notification Handler"
Cohesion: 0.11
Nodes (23): Notification Handler, hasSystemNotificationUpdateFields(), NewNotificationHandler(), POST /api/news — Create news, DELETE /api/news/{id} — Delete news, GET /api/news/{id} — Get news by id/slug, News Endpoints, news_handler.go (+15 more)

### Community 13 - "Service Rating & List Handlers"
Cohesion: 0.14
Nodes (18): NewServiceRatingHandler(), POST /api/ratings — Rate service request (protected), Service Ratings Endpoints, GET /api/services/{id}/ratings — List service ratings, GET /api/services/{id}/rating-stats — Rating stats, models/service_attendance.go, service_rating_handler.go, models/service_rating.go (+10 more)

### Community 14 - "Storage Service & Tests"
Cohesion: 0.21
Nodes (11): failingMockStorage, mockStorageService, NewSupabaseStorageService(), TestSupabaseStorageService_DeleteFile_ServerError(), TestSupabaseStorageService_DeleteFile_Success(), TestSupabaseStorageService_DeleteFile_WrongBucket(), TestSupabaseStorageService_UploadFile_ServerError(), TestSupabaseStorageService_UploadFile_Success() (+3 more)

### Community 15 - "Category & Service Handlers"
Cohesion: 0.12
Nodes (18): category_handler.go, models/category.go, Category Handler, NewCategoryHandler(), NewServiceHandler(), service_handler.go, models/service_icons.go, models/service.go (+10 more)

### Community 16 - "Core Bootstrap"
Cohesion: 0.13
Nodes (20): Main Entry Point, Config Package, Config Struct, Load Config Function, Database Connect Function, Main Function, NewAppConfigRepository Constructor, NewCategoryRepository Constructor (+12 more)

### Community 17 - "Team Handler & Models"
Cohesion: 0.15
Nodes (18): TeamHandler, AddTeamMember, CreateTeam, DeleteTeam, GetMyTeam, GetTeam, GetTeamDashboard, ListTeamMembers (+10 more)

### Community 18 - "User Handler & Models"
Cohesion: 0.15
Nodes (16): UserHandler, CreateUser, DeleteProfileImage, DeleteUser, GetMe, GetUser, ListUsers, UpdateUser (+8 more)

### Community 20 - "Expo Push Service & Tests"
Cohesion: 0.23
Nodes (7): chunkStrings(), NewExpoPushService(), TestExpoPushService_SendNewsPublished_ChunksTokens(), TestExpoPushService_SendNewsPublished_EmptyTokens(), TestExpoPushService_SendNewsPublished_SendsExpectedPayload(), ExpoPushMessage, ExpoPushService

### Community 21 - "System Notification Repository"
Cohesion: 0.15
Nodes (12): NewSystemNotificationRepository, SystemNotificationRepository, SystemNotificationRepository.Create, SystemNotificationRepository.Delete, SystemNotificationRepository.GetByID, SystemNotificationRepository.List, SystemNotificationRepository.MarkAsRead, SystemNotificationRepository.Update (+4 more)

### Community 22 - "Region Repository"
Cohesion: 0.15
Nodes (12): NewRegionRepository, RegionRepository, RegionRepository.Create, RegionRepository.Delete, RegionRepository.FindByNeighborhood, RegionRepository.GetByID, RegionRepository.List, RegionRepository.ListAll (+4 more)

### Community 23 - "Handler Injection Hub"
Cohesion: 0.18
Nodes (13): AppConfigHandler, AuthHandler, CategoryHandler, GeolocationHandler, NewsHandler, NotificationHandler, RegionHandler, ServiceHandler (+5 more)

### Community 25 - "Team Repository (alt)"
Cohesion: 0.17
Nodes (12): NewTeamRepository, TeamRepository, TeamRepository.AddMember, TeamRepository.CreateTeam, TeamRepository.DeleteTeam, TeamRepository.GetTeamByID, TeamRepository.GetTeamByRegion, TeamRepository.GetTeamStats (+4 more)

### Community 27 - "User Repository"
Cohesion: 0.29
Nodes (3): formatBirthDate(), NewUserRepository(), UserRepository

### Community 28 - "Home Models (AST)"
Cohesion: 0.18
Nodes (10): CategoryStat, HomeAlert, HomeResponse, HomeStats, MapLocation, PopularService, RecentRequest, StatDetail (+2 more)

### Community 29 - "Home Models (Semantic)"
Cohesion: 0.36
Nodes (10): CategoryStat, HomeAlert, HomeResponse, HomeStats, MapLocation, PopularService, RecentRequest, StatDetail (+2 more)

### Community 30 - "Service Handler"
Cohesion: 0.18
Nodes (11): ServiceHandler, CreateService, DeleteService, GetService, ListCategories, ListServices, ListServicesByCategory, ListServicesByCategoryID (+3 more)

### Community 31 - "Upload Service Tests"
Cohesion: 0.18
Nodes (11): TestUploadService_AllowedFileTypes, TestUploadService_DisallowedExtension, TestUploadService_DisallowedMIMEType, TestUploadService_ExtensionCaseInsensitive, TestUploadService_FileTooLarge, TestUploadService_MaxFilesExceeded, TestUploadService_NoFiles, TestUploadService_RollbackOnUploadFailure (+3 more)

### Community 32 - "Team Handler"
Cohesion: 0.31
Nodes (9): NewTeamHandler(), team_handler.go, models/team.go, POST /api/teams — Create team, DELETE /api/teams/{id} — Delete team, GET /api/teams/{id} — Get team by id, Teams Endpoints, GET /api/teams — List teams (+1 more)

### Community 33 - "User Models"
Cohesion: 0.2
Nodes (8): CreateUserRequest, ErrorResponse, LoginRequest, LoginResponse, MessageResponse, UpdateUserRequest, User, UserDetailResponse

### Community 34 - "Service Repository (alt)"
Cohesion: 0.2
Nodes (10): NewServiceRepository, ServiceRepository, ServiceRepository.CreateService, ServiceRepository.DeleteService, ServiceRepository.GetServiceByID, ServiceRepository.ListCategories, ServiceRepository.ListServices, ServiceRepository.ListServicesByCategory (+2 more)

### Community 35 - "Service Rating Handler"
Cohesion: 0.27
Nodes (9): ServiceRatingHandler, CreateRating, GetRatingStats, ListRatingsByService, CreateServiceRatingRequest, ServiceRating, ServiceRatingResponse, ServiceRatingStats (+1 more)

### Community 36 - "Storage Service Tests"
Cohesion: 0.2
Nodes (10): NewSupabaseStorageService, TestSupabaseStorageService_DeleteFile_ServerError, TestSupabaseStorageService_DeleteFile_Success, TestSupabaseStorageService_DeleteFile_WrongBucket, TestSupabaseStorageService_UploadFile_ServerError, TestSupabaseStorageService_UploadFile_Success, TestSupabaseStorageService_UploadFile_UsesReader, supabaseStorageService (+2 more)

### Community 38 - "User Repository (alt)"
Cohesion: 0.22
Nodes (9): NewUserRepository, UserRepository, UserRepository.CreateUser, UserRepository.DeleteUser, UserRepository.GetUserByID, UserRepository.GetUserByUsernameOrEmail, UserRepository.ListUsers, UserRepository.UpdateUser (+1 more)

### Community 39 - "News Repository"
Cohesion: 0.22
Nodes (9): NewNewsRepository, NewsRepository, NewsRepository.CreateNews, NewsRepository.DeleteNews, NewsRepository.GetNews, NewsRepository.GetNewsBySlug, NewsRepository.ListNews, NewsRepository.UpdateNews (+1 more)

### Community 40 - "Expo Push Service"
Cohesion: 0.22
Nodes (9): ExpoPushService, ExpoPushService.SendNewsPublished, ExpoPushService.SendToUser, ExpoPushService.sendBatch, NewExpoPushService, TestExpoPushService_SendNewsPublished_ChunksTokens, TestExpoPushService_SendNewsPublished_EmptyTokens, TestExpoPushService_SendNewsPublished_SendsExpectedPayload (+1 more)

### Community 41 - "Core Bootstrap (alt)"
Cohesion: 0.25
Nodes (4): main(), Config, Load(), NewServiceRequestRepository()

### Community 43 - "Service Attendance Repository (alt)"
Cohesion: 0.25
Nodes (7): NewServiceAttendanceRepository, ServiceAttendanceRepository, ServiceAttendanceRepository.Create, ServiceAttendanceRepository.GetByID, ServiceAttendanceRepository.ListByRequestID, CreateServiceAttendanceRequest, ServiceAttendance

### Community 45 - "App Config Models"
Cohesion: 0.29
Nodes (6): AppBanner, AppConfig, CategorySummary, MobileHomeResponse, Section, ServiceSummary

### Community 46 - "Notification Models"
Cohesion: 0.29
Nodes (4): CreateSystemNotificationRequest, RegisterPushTokenRequest, SystemNotification, UpdateSystemNotificationRequest

### Community 47 - "Team Models"
Cohesion: 0.29
Nodes (6): CreateTeamRequest, MyTeamResponse, Team, TeamMember, TeamStats, UpdateTeamRequest

### Community 48 - "Category Repository (alt)"
Cohesion: 0.29
Nodes (7): CategoryRepository, CategoryRepository.Create, CategoryRepository.Delete, CategoryRepository.GetByID, CategoryRepository.List, CategoryRepository.Update, NewCategoryRepository

### Community 49 - "Upload & Attendance Tests"
Cohesion: 0.29
Nodes (7): NewUploadService, ServiceAttendanceHandler, TestRollbackFiles, TestRollbackFiles_BestEffortOnDeleteFailure, TestRollbackFiles_Empty, UploadService, UploadService.RollbackFiles

### Community 50 - "Service Models (AST)"
Cohesion: 0.33
Nodes (5): CreateServiceRequest, Service, ServiceDetailResponse, StatusStat, UpdateServiceRequest

### Community 51 - "Middleware Auth Layer"
Cohesion: 0.4
Nodes (5): Auth, RequireRole, UserIDKey, contextKey, ErrorResponse

### Community 52 - "Service Rating Repository (alt)"
Cohesion: 0.33
Nodes (6): NewServiceRatingRepository, ServiceRatingRepository, ServiceRatingRepository.Create, ServiceRatingRepository.GetByRequestID, ServiceRatingRepository.GetStatsByServiceID, ServiceRatingRepository.ListByServiceID

### Community 53 - "Service Models (Semantic)"
Cohesion: 0.47
Nodes (5): CreateServiceRequest, Service, ServiceDetailResponse, StatusStat, UpdateServiceRequest

### Community 54 - "Geocoding & Home Service"
Cohesion: 0.33
Nodes (6): GeocodingService, GeocodingService.GeocodeAddress, HomeHandler, NewGeocodingService, ServiceRequestHandler, extractBairroFromNominatim

### Community 55 - "Upload Attachment Tests"
Cohesion: 0.33
Nodes (6): ParseAttachmentURLs, TestFileUploadError_Format, TestParseAttachmentURLs_Empty, TestParseAttachmentURLs_EmptyArray, TestParseAttachmentURLs_InvalidJSON, TestParseAttachmentURLs_Valid

### Community 57 - "Category Models"
Cohesion: 0.4
Nodes (4): Category, CategoryDetailResponse, CreateCategoryRequest, UpdateCategoryRequest

### Community 58 - "Service Rating Models"
Cohesion: 0.4
Nodes (4): CreateServiceRatingRequest, ServiceRating, ServiceRatingResponse, ServiceRatingStats

### Community 59 - "Service Request Models"
Cohesion: 0.4
Nodes (4): CreateServiceRequestRequest, ServiceRequest, ServiceRequestDetailResponse, UpdateServiceRequestStatusRequest

### Community 60 - "Category Models (Semantic)"
Cohesion: 0.5
Nodes (4): Category, CategoryDetailResponse, CreateCategoryRequest, UpdateCategoryRequest

### Community 61 - "Region Models"
Cohesion: 0.5
Nodes (3): CreateRegionRequest, Region, UpdateRegionRequest

### Community 62 - "Service Icons Models"
Cohesion: 0.5
Nodes (3): GetCategoryIcon, GetServiceIcon, ServiceIconMapping

## Knowledge Gaps
- **273 isolated node(s):** `Config`, `contextKey`, `AppBanner`, `AppConfig`, `ServiceSummary` (+268 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **12 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Setup()` connect `Auth & Geolocation Handlers` to `Team Handler`, `API Endpoints Documentation`, `Upload Service & Tests`, `Home & Region Filter Handlers`, `Core Bootstrap (alt)`, `News & App Config Handlers`, `Notification Handler`, `Service Rating & List Handlers`, `Category & Service Handlers`, `Expo Push Service & Tests`?**
  _High betweenness centrality (0.375) - this node is a cross-community bridge._
- **Why does `main()` connect `Core Bootstrap (alt)` to `AppConfig & Stats Repository`, `Auth & Geolocation Handlers`, `Category Repository`, `News & Notification Repositories`, `Service Rating Repository`, `Service Attendance Repository`, `Storage Service & Tests`, `Team Repository`, `Database Connection`, `Service Repository`, `Region Repository (alt)`, `User Repository`?**
  _High betweenness centrality (0.255) - this node is a cross-community bridge._
- **Why does `ServiceRequestHandler` connect `Service Request & Auth Handlers` to `Service Rating Handler`, `Team Handler & Models`, `User Handler & Models`, `System Notification Repository`, `Home Models (Semantic)`, `Service Icons Models`?**
  _High betweenness centrality (0.183) - this node is a cross-community bridge._
- **Are the 61 inferred relationships involving `respondError()` (e.g. with `.Login()` and `.ListCategories()`) actually correct?**
  _`respondError()` has 61 INFERRED edges - model-reasoned connections that need verification._
- **Are the 30 inferred relationships involving `ServiceRequestRepository` (e.g. with `NewServiceRequestRepository` and `ServiceRequestRepository.CreateServiceRequest`) actually correct?**
  _`ServiceRequestRepository` has 30 INFERRED edges - model-reasoned connections that need verification._
- **Are the 27 inferred relationships involving `handlers package` (e.g. with `generateSlug()` and `extractSupabaseURLs()`) actually correct?**
  _`handlers package` has 27 INFERRED edges - model-reasoned connections that need verification._
- **Are the 27 inferred relationships involving `respondJSON()` (e.g. with `handlers package` and `TestRespondJSON()`) actually correct?**
  _`respondJSON()` has 27 INFERRED edges - model-reasoned connections that need verification._