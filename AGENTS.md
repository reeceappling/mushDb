# Repository Instructions

This repository is a full-stack monorepo featuring a Go REST API backend and a Next.js App Router frontend.

## Hardline Rules
[//]: # (TODO: ensure ok!)
- Do not touch, read, or modify any files in `/noCommit`, `/initDB`, `/etc`, or `/env`, or any file with a name prefix `noCommit_backup_`. Listing the files is fine, but reading their contents is not. Some of these files include secrets which should never be exposed in any way to anyone except the owner of the repo, including yourself. These files must NEVER be deleted or moved.

## Project Structure
- `main.go` - Entrypoint for the Go backend API.
- `/api` - Go backend API (Standard library) module. Utilized by `main.go`. Go API AGENTS.md can be found at go/AGENTS.md
- `/web` - Next.js 16+ frontend application (App Router, Tailwind CSS, TypeScript). Webserver AGENTS.md can be found at web/AGENTS.md
- `/native` - React-Native application that is extremely similar to web but as a native app. React Native AGENTS.md can be found at native/AGENTS.md
- `/scripts` - Utility scripts for development, testing, and deployment.
- Files used to connect all the parts.
  - `docker-compose.yml` - Docker Compose configuration.
  - `docker-compose-web-dev.yaml` - Docker Compose configuration for web dev deployment.
  - `docker-compose-web-prod.yaml` - Docker Compose configuration for web prod deployment.
  - `DockerfileApi` - Docker Compose file for the Go backend API.
  - `/env` - Environment variable files for different environments.
---

## Testing & Verification
- **Unit Tests:** Run unit tests for backend and frontend using the respective commands outlined in their subsections below.
- **Integration Tests:** Run integration tests for backend and frontend using the respective commands outlined in their subsections below.
- **End-to-End Tests:** Run end-to-end tests using Cypress.
- [//]: # (TODO: more?)


## 1. Backend (`main.go` and `/api` - Go)

### Commands
[//]: # (TODO- **Run Dev Server:** `cd api && go run main.go` &#40;or `air` for hot reload if configured&#41;)
[//]: # (TODO- **Run Tests:** `cd api && go test ./... -v`)
[//]: # (TODO- **Lint/Check:** `cd api && go vet ./...`)

### Rules & Conventions
- **Error Handling:** Never ignore returned errors, unless they are an impossiblity. Handle them explicitly (`if err != nil { return err }`) if there is any chance they could occur.

### Testing
[//]: # (TODO- linting)
[//]: # (TODO- run go unit tests)
[//]: # (TODO- run go integration tests)
[//]: # (TODO- see end-to-end testing area with cypress)

[//]: # (TODO- **Dependency Injection:** Explicitly pass dependencies &#40;database connections, config&#41; via structs rather than relying on global package state.)
- 
- **JSON Mapping:** Ensure JSON payload structs use explicit camelCase `json:"..."` tags matching frontend consumption.
- **Context:** Always pass `context.Context` as the first argument to database and long-running operations. When a function takes a context, it should always be the first argument.

---

## 2. Frontend (`/web` - Next.js)

### Commands
- **Install Dependencies:** `cd web && pnpm install`

[//]: # (TODO:- **Run Dev Server:** `cd web && pnpm dev`)
[//]: # (TODO:- **Build Production:** `cd web && pnpm build`)
[//]: # (TODO:- **Run Tests/Lint:** `cd web && pnpm test && pnpm lint`)

### Testing
[//]: # (TODO- linting and typechecking)
[//]: # (TODO- run ts unit tests)
[//]: # (TODO- run ts integration tests)
[//]: # (TODO- see end-to-end testing area with cypress)

### Rules & Conventions
- **App Router:** Use Next.js App Router conventions (`app/` directory). Name page and layout components correctly (`page.tsx`, `layout.tsx`).
- **API Integration:** Fetch backend endpoints from `/api` using absolute or environment-configured proxy URLs (`NEXT_PUBLIC_API_URL`). Do not hardcode localhost ports except in fallback dev configs.
- **TypeScript:** Strict mode is enforced. Avoid `any`; define explicit interfaces or types for all API response data structures.
- **Server vs Client Components:** Default to Server Components. Add `'use client'` explicitly only when utilizing hooks (`useState`, `useEffect`, event listeners) or browser-only APIs.

[//]: # (TODO: TESTS: WRITE TESTS IF NOT EXISTING!)
[//]: # (TODO: GO TESTS!)
[//]: # (TODO: VITEST for web unit tests and integration tests!)
[//]: # (TODO: end-to-end testing with cypress!)
[//]: # (TODO: jest for native!)


---

## 3. Native App (`/native` - React Native and Next.js) 
[//]: # (TODO: ok?)

### Commands

[//]: # (TODO:- **Install Dependencies:** `cd web && pnpm install`)

[//]: # (TODO:- **Run Dev Server:** `cd web && pnpm dev`)
[//]: # (TODO:- **Build Production:** `cd web && pnpm build`)
[//]: # (TODO:- **Run Tests/Lint:** `cd web && pnpm test && pnpm lint`)

### Testing
[//]: # (TODO- linting and typechecking)
[//]: # (TODO- run unit tests)
[//]: # (TODO- run integration tests)
[//]: # (TODO- see end-to-end testing area with cypress?)

### Rules & Conventions

[//]: # (TODO:- **App Router:** Use Next.js App Router conventions &#40;`app/` directory&#41;. Name page and layout components correctly &#40;`page.tsx`, `layout.tsx`&#41;.)

[//]: # (TODO:- **API Integration:** Fetch backend endpoints from `/api` using absolute or environment-configured proxy URLs &#40;`NEXT_PUBLIC_API_URL`&#41;. Do not hardcode localhost ports except in fallback dev configs.)

[//]: # (TODO:- **TypeScript:** Strict mode is enforced. Avoid `any`; define explicit interfaces or types for all API response data structures.)

[//]: # (TODO:- **Server vs Client Components:** Default to Server Components. Add `'use client'` explicitly only when utilizing hooks &#40;`useState`, `useEffect`, event listeners&#41; or browser-only APIs.)

---

## General Verification Workflow

1. **Verify changes do not break cross-service contract expectations** (API response shapes vs frontend types).
3. **Verify changes would not break any existing deployment with a fully populated database with all data types**

[//]: # (TODO: REWORD LAST LINE)

2. **Run backend tests**
3. **Run frontend tests**
4. **Run End-to-End tests**

[//]: # (TODO: does this cover linting and typechecking?)


[//]: # (TODO: Define explicit boundaries:)