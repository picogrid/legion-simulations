# Package Structure

This directory contains the core packages for the Legion Simulations project.

## `/client`
**Hand-written Legion API client**

A clean, manually-written client for the Legion API with:
- Domain-organized operations (entities, users, organizations, etc.)
- OAuth2 and API key authentication support
- Context-aware operations
- Clean error handling

Key files:
- `client.go` - Core client functionality
- `entities.go` - Entity management
- `users.go` - User operations
- `organizations.go` - Organization management
- `feeds.go` - Feed operations
- `locations.go` - Entity location management
- `helpers.go` - Convenience functions

## `/auth`
**Authentication and token management**

Handles all authentication concerns:
- `keycloak.go` - Keycloak client for OAuth2 flows
- `token_manager.go` - Automatic token refresh management
- `helpers.go` - Authentication helpers and user prompts

## `/models`
**Generated API models**

Auto-generated from OpenAPI specification. Do not edit manually.

## `/simulation`
**Simulation framework**

Core interfaces and registry for simulations:
- `interface.go` - Simulation interface definition
- `registry.go` - Simulation registration and discovery
- `config.go` - Configuration structures

## `/config`
**Environment configuration**

Manages Legion environment configurations (URLs, authentication methods).

## `/utils`
**Utility functions**

Helper functions for:
- Simulation discovery
- Parameter prompting
- Common operations

## Architecture Decisions

1. **Hand-written Client**: We use a hand-written client (`/client`) because:
   - Simpler API without complex dependencies
   - Better error handling
   - More idiomatic Go code
   - Easier to understand and maintain
   - No code generation complexity

2. **Domain Organization**: Client operations are organized by domain (entities, users, etc.) for better discoverability and maintainability.

3. **Authentication Separation**: All authentication logic is centralized in `/auth` for reusability across different client implementations.

4. **Models Reuse**: We use the generated models from `/models` to ensure type safety while keeping our client implementation clean.