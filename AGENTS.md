# AI Agents Project Instructions

This project is developed using Domain-Driven Design (DDD) principles.

## Architectural Guidelines

- We use Clean Architecture combined with Hexagonal (Ports & Adapters) Architecture.
- The domain layer is the core of the system and must not depend on external layers.
- We use a **rich domain model**, not an anemic one.
- Business logic must reside in the domain and application layers, never in the UI or infrastructure layers.

## Layering Rules and Code Generation Order

Code must always be generated **sequentially by layers**, respecting dependencies:

1. **Domain layer**
2. **Infrastructure layer**
3. **Application layer**
4. **UI (Transport) layer**

Never generate or implement an upper layer before the lower layers are clearly defined.
Skipping layers (e.g. Domain -> UI) is strictly forbidden. The Application layer is mandatory for all use cases.

### Domain Layer

- Uses **Aggregates**, **Entities**, and **Value Objects**.
- Aggregates define transactional boundaries.
- Entities have identity and lifecycle.
- Value Objects are immutable and have no identity.
- Contains business rules and invariants.
- Domain models must be persistence-agnostic.
- No references to infrastructure, frameworks, or transport concerns.
- Use private fields (lowercase) in domain structs.
- Expose data through getter methods only.
- Never add `json` tags to domain structs.
- Never add `MarshalJSON()` to domain entities.
- Serialization is UI layer responsibility.

### Infrastructure Layer

- Implements persistence, external services, and adapters.
- Depends on domain abstractions.
- Contains repositories, database mappings, message brokers, and external API clients.
- Must not contain business logic.
- Repository implementations MUST use ORM abstractions (GORM) and the project's `rdbms` package. Avoid raw SQL.
- Always extract database-specific models (GORM structs) into a dedicated `infrastructure/models.go` file. Do not reuse Domain models as DB models.

### Application Layer

- Uses **Commands** to represent use cases.
- Orchestrates domain logic and coordinates aggregates.
- Manages transactions and consistency boundaries.
- Depends on domain and infrastructure abstractions.
- Contains no transport- or framework-specific code.
- Use standard `log/slog` for logging instead of custom loggers.
- MUST define and return custom domain/application error types instead of raw error strings.

### UI / Transport Layer

- Uses **DTOs only**.
- Responsible for HTTP/API concerns, validation, mapping, and response formatting.
- Must not contain business logic.
- Calls application commands.
- Keep controllers thin. Extract DTOs into a separate `dto.go` file and mappers logic into `mapper.go` within the UI package.

## Swagger / OpenAPI Rules

- Descriptions of methods in Swagger must always be in Russian.
- `swagger.json` is **generated automatically**.
- Swagger is derived from:
  - Annotations in controller methods
  - DTO definitions in the UI layer
- Swagger files must **never** be edited manually.
- The UI layer is the single source of truth for API contracts.
- We use annotation in code of controllers and dto <https://github.com/go-swagger/go-swagger> for generation spec. Params and queries always should be in dto struct.
- Enum values in Swagger annotations MUST use the JSON array format: `// Enum: ["value1","value2"]`. Comma-separated lists are deprecated and cause generation errors.
- For array fields, use `items.enum` instead of `Enum` to specify allowed values.

## General Rules

- All error messages must be in English.
- Follow strict separation of concerns.
- Avoid anemic domain models.
- Do not leak infrastructure or transport details into the domain.
- Interfaces are defined only at architectural boundaries.
- Dependencies must always point inward.
- Context must be passed explicitly.
- Use `ErrorHandler` from `handlers` package for error responses.
- For IDs we use UUID.
- Always check README.md file for project description.
- Always check code syntax and run build after changes.
- When creating something new (services, repositories, controllers, etc.), it MUST be registered in `dependencies.go` for Dependency Injection (DI).
- Repository interfaces go in `domain/{domain}_repository_interface.go`.
- Commands and Queries go in `application/commands.go`.
- GORM models go in `infrastructure/models.go`.
- Value Objects go in `domain/vo.go`.
- Use `BodyResponse` (with `status` and `errors`) for controllers **ONLY** for endpoints that return no data or for error responses.
- For endpoints returning data, **DO NOT** wrap the payload in a `data` field alongside `status` and `errors`. Return the DTO directly as the JSON body.
- Use `unixtimestamp_migration_name.up|down.sql` mask for migration name.
- One migration file per domain (up + down), not one per table.
- Include proper foreign keys, CASCADE deletes, and indexes.
- Do not use meaningless comments.

## HTTP Status Codes

Use appropriate HTTP status codes for different error types:

- `404 Not Found` - for "not found" errors (resource, event, calendar, etc.)
- `403 Forbidden` - for "access denied" errors
- `400 Bad Request` - for validation errors and JSON parsing errors
- `500 Internal Server Error` - for unexpected internal errors

## Access Control

- All user-facing operations must check access before executing.
- Create a helper method `checkAccess(resource, userID)` in service layer.
- Check if user is owner OR in access list.
- Access check applies to: Create, Update, Delete, Get, List operations.
- Webhooks and callbacks must be listed in `access.whitelist` in config so they work without a user JWT.

## Pagination

- For all "listing" (list) endpoints, always use pagination via `limit`, `offset` query parameters.
- The response must always include `totalItems` (total count of matching records) alongside the list of items.
- This pattern applies to every list/search endpoint in the project.

## JSONB Fields

- Use `string` type in GORM models for JSONB columns.
- Provide helper functions for serialization/deserialization of JSONB data.
- Handle both empty string and `"null"` as nil when parsing.

## GORM Associations

- Declare `HasMany` relationships via `gorm:"foreignKey:FieldName;constraint:OnDelete:CASCADE"` tags.
- For updates that replace all children (variants/activities), use a transaction:
  1. Delete all existing children.
  2. Save the parent (GORM will create new children with the saved parent's ID).
- Use `r.db.WithContext(ctx)` for context propagation.

## Repository Interface

- Always use `context.Context` in repository interface methods (even if some older code doesn't).
- The `rdbms.Executor` interface supports `WithContext(ctx)` and `Transaction()`.
- Use `rdbms.TxStorage` / `rdbms.Tx` for transactional storage patterns with generics.

## Branch Workflow

- **Никогда не пушить коммиты напрямую в `dev` без MR.**
- Все изменения в `dev` попадают только через Merge Request (MR).
- MR должен быть аппрувнут перед мержем.
- Ветки именуются по шаблону: `feature/{TICKET_KEY}`, `bugfix/{TICKET_KEY}`, `hotfix/{TICKET_KEY}`.
- После мержа ветка удаляется.
- Если задача требует дополнительных исправлений — создаётся новый MR, не пуш в тот же `dev`.

## Scope Boundary

This file contains **project-level architectural guidelines and conventions only**. It must not include:

- Agent-specific custom commands, workflow steps, or personal preferences.
- Skill content, skill instructions, or tool configuration.
- Self-reflection instructions, self-improvement directives, or post-implementation routines.
- OpenCode configuration details (custom commands, subagent prompts, plugin setup).

If you discover something useful that is not project-wide, put it in your own agent configuration or skill files.
