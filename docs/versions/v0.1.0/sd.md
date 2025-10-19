# System Design — Prein v0.1.0

Last updated: 2025-10-13

This System Design (SD) expands the System Analysis (SA) into concrete design artifacts for implementation, including architecture diagrams, data models (DDL), API contracts, authentication and session handling, deployment considerations, testing plans, acceptance criteria, and operational runbook notes.

## Goals

- Implement a minimal, secure, and maintainable system that provides Google sign-in and persists Google account ID and source.
- After sign-in, redirect users to a homepage with a header (logo) and hero section.
- Provide a small, extensible backend (Go + Gin) and frontend (Next.js) integration with PostgreSQL as the datastore.

## Design Constraints and Assumptions

- Backend: Go (Gin). Frontend: Next.js (React + TypeScript).
- PostgreSQL is the primary persistent store.
- Development may use HTTP and localhost; production must use HTTPS and secure cookie settings.
- Sessions will be implemented as secure, HttpOnly cookies issued by the backend for v0.1.0.
- OAuth with Google will use server-side flow (authorization code grant) to ensure backend receives profile data and can persist provider id.

## High-level Architecture

The system has three main components:

- Frontend (Next.js)
  - UI pages: Login page with "Sign in with Google" button, Homepage (post-login) with header and hero.
  - Initiates server-side OAuth by redirecting the browser to backend `/auth/google/start` endpoint.

- Backend (Gin)
  - Auth routes: `/auth/google/start`, `/auth/google/callback`.
  - API routes: `/api/v1/users/current`, `/health`.
  - Session management: create/validate HttpOnly secure cookies.
  - DB access layer: upsert users, query current user.

- Database (PostgreSQL)
  - Tables: users.

Interaction flow:

1. User visits the frontend and clicks "Sign in with Google".
2. Frontend navigates to backend `/auth/google/start`.
3. Backend generates a short-lived CSRF `state`, stores it (in-memory map for v0.1.0), and redirects user to Google's OAuth 2.0 authorization endpoint with `client_id`, `redirect_uri`, `scope`, and `state`.
4. User authenticates at Google and consents. Google redirects to `/auth/google/callback` with `code` and `state`.
5. Backend validates `state`, exchanges `code` for tokens, decodes/validates an ID token (or requests userinfo), extracts `sub` (provider id), email, name, avatar.
6. Backend upserts a `users` record using `provider_id` and `source = "google"`.
7. Backend issues a secure HttpOnly session cookie and redirects browser back to frontend homepage (`/`).
8. Frontend can call `/api/v1/users/current` to fetch the current user (authenticated via session cookie).

### Diagram (ASCII)

Frontend (Next.js)
  |
  |-- GET /auth/google/start --> Backend (Gin)
                                |-- redirect to Google OAuth
  |<-- 302 redirect -----------

Google OAuth
  |-- redirects to /auth/google/callback (backend)

Backend (Gin)
  |-- exchange code -> tokens
  |-- upsert users -> Postgres
  |-- set session cookie
  |-- redirect to frontend /

Frontend
  |-- GET /api/v1/users/current (with cookie) -> Backend -> returns user

## Data Models (DDL)

Below are SQL schemas sufficient for v0.1.0. They are intentionally small and clear so they can be used with sqlc/pgx or any SQL migration tooling.

-- users table

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  provider_id TEXT NOT NULL UNIQUE,
  email TEXT,
  name TEXT,
  family_name TEXT,
  given_name TEXT,
  avatar_url TEXT,
  source users_source NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

Notes:
- Use a migration tool (golang-migrate) to apply these DDLs. For local development the repository should include migration files under `backend/db/migrations` (future step).
- For production, consider a stronger uniqueness and normalization strategy for emails if supporting multiple providers.

## API Contracts

General notes:
- All API responses are JSON.
- Backend uses cookie-based sessions; the client need not include Authorization headers for authenticated routes.
- Time fields are RFC3339 (ISO 8601) strings.

1) GET /health
- Purpose: Health check.
- Auth: none
- Response 200:
  { "status": "ok" }

2) GET /auth/google/start
- Purpose: Initiate OAuth 2.0 authorization code flow with Google.
- Auth: none
- Behavior: Redirects (302) to Google's OAuth endpoint.
- Query/Headers: optional `redirect_to` param (frontend route to return to after auth).

3) GET /auth/google/callback
- Purpose: OAuth callback endpoint.
- Auth: none
- Query: code, state
- Behavior: Validate state, exchange code for tokens (and/or verify ID token), extract profile, upsert user, create session, redirect back to frontend (e.g., /).
- Error cases:
  - 400: missing code/state or invalid state
  - 401: token exchange failed
  - 500: DB error or session creation failed

4) GET /api/v1/users/current
- Purpose: Return current authenticated user profile.
- Auth: session cookie required
- Response 200:
  {
    "id": "uuid",
    "provider_id": "google-sub-id",
    "email": "user@example.com",
    "name": "User Name",
    "avatar_url": "https://...",
    "source": "google",
    "created_at": "2025-10-13T12:00:00Z",
    "updated_at": "2025-10-13T12:00:00Z"
  }
- Response 401: when session is missing or invalid.

Session contract
- Session cookie name: `prein_sess`
- Cookie flags: HttpOnly, Secure (in production), SameSite=Lax, Path=/, Max-Age configured (7 days by default).
- Session store: in-memory map for v0.1.0; supports mapping session ID -> user ID.

## Authentication & Session Handling Details

Auth flow details
- Use OAuth 2.0 Authorization Code grant plus OIDC ID token when possible (request `openid email profile` scopes).
- Prefer validating the ID token's signature and claims (issuer, audience, expiry) rather than calling userinfo, but either is acceptable for v0.1.0.
- The backend is responsible for upserting users and creating a server-side session.

CSRF/State handling
- Generate cryptographically random `state` values and store them server-side in a short-lived map keyed by `state` -> { created_at, redirect_to }.
- Validate the `state` in callback and expire after use or after a short TTL (e.g., 10 minutes).

Session lifecycle
- On successful auth, create a new session ID (secure random, e.g., 32 byte base64) and store { user_id, created_at, expires_at } in an in-memory store.
- Set session cookie with session ID, HttpOnly flag, and appropriate SameSite and Secure flags.
- For initial development this store can be an in-memory Go map with a background cleaner. For production, replace with Redis or another durable store.
- Provide a logout endpoint (later) that deletes the session server-side and clears cookie.

## Upsert semantics for users

- Use a single SQL statement that inserts or updates on conflict by `provider_id` (Postgres `ON CONFLICT (provider_id) DO UPDATE SET ...`) to avoid duplicates on concurrent sign-ins.
- Update fields such as `email`, `name`, `avatar_url`, `updated_at` on each sign-in.

Example SQL (simplified):

INSERT INTO users (provider_id, email, name, avatar_url, source)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (provider_id) DO UPDATE
SET email = EXCLUDED.email,
    name = EXCLUDED.name,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = NOW();

Return the user's `id` for session creation.

## Observability and Logging

- Log authentication attempts (success/failure), reasons for failures, and user creation/upsert events.
- Record minimal telemetry: counts of sign-ins, failed sign-ins, and sign-in latency.
- Use structured logs (JSON) at minimum: timestamp, level, event, user_id (if available), error.
- Expose `/health` and consider a metrics endpoint for Prometheus in future iterations.

## Testing Strategy

Unit tests
- Token exchange logic: mock Google's token endpoint responses, test ID token parsing and validation, test error handling.
- DB upsert logic: test upsert produces single user and updates fields.
- Session store: test session creation, retrieval, expiration.

Integration tests
- Start backend with test configuration and a fake Google OAuth server (or mock HTTP) to simulate code/token exchange and userinfo.
- Verify callback flow creates user and session cookie and redirects properly.

End-to-end (E2E)
- Use Playwright or Cypress to run a test that:
  1. Navigates to the frontend login page.
  2. Clicks "Sign in with Google" which redirects to backend.
  3. Simulates Google consent (using test client or stub) and returns to callback.
  4. Verifies the frontend receives a session and user is redirected to homepage.

Manual test cases
- Happy path sign-in and redirect.
- Repeat sign-in with same Google account should not create duplicates.
- Callback with invalid or missing state should be rejected.
- Missing email in Google's profile should cause an error (v0.1.0 requires email).

## Security Considerations

- Use HTTPS in production and set cookie Secure flag.
- Validate ID tokens from Google (issuer, audience, expiry, signature) instead of trusting userinfo responses when possible.
- Rotate and protect OAuth client secrets. Store them in environment variables or a secret manager; do not commit them.
- Limit the data stored: only provider_id, email, name, avatar_url, and source.
- Protect against session fixation by issuing new session IDs after login.
- Rate limit auth endpoints to mitigate abuse.

## Deployment & Operations

Environment variables (minimum):
- GOOGLE_CLIENT_ID
- GOOGLE_CLIENT_SECRET
- OAUTH_REDIRECT_URL (e.g., https://example.com/auth/google/callback)
- SESSION_SECRET (for signing session IDs or cookies if applying HMAC)
- DATABASE_URL (Postgres DSN)
- ENV (development|production)

Local development notes
- For local OAuth testing, configure a Google OAuth client with redirect URI pointing to localhost (e.g., http://localhost:8080/auth/google/callback) and set the client ID/secret in a `.env` file.
- Backend run command (dev): `go run .` (see `backend/README.md`).

Production notes
- Deploy backend behind TLS termination.
- Use a persistent session store (Redis) and migrate in-memory maps to Redis.
- Use a managed Postgres instance and run migrations via `golang-migrate` during deployment.
- Enable structured logging and a centralized log sink (e.g., Cloud Logging, Datadog).

## Acceptance Criteria (mapped)

- Users can sign in with Google via the server-side flow.
- Backend stores `provider_id` and `source = "google"` for each user.
- Repeated sign-ins with the same Google account do not create duplicate users (upsert behavior).
- After sign-in, users are redirected to the frontend homepage which shows a header with logo and hero section.
- `GET /api/v1/users/current` returns the authenticated user's profile.

## Risks & Mitigations

- OAuth misconfiguration (redirect URI mismatch): document step-by-step setup and test locally.
- Duplicate users due to race conditions: use DB unique constraint and `ON CONFLICT` upsert.
- Session state loss across services: use Redis for sessions in production.
- Missing email fields from Google: request `email` scope and fail early if absent.

## Next Steps (Implementation Plan)

1. Add `.env.example` entries for `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `OAUTH_REDIRECT_URL`, `SESSION_SECRET`, and `DATABASE_URL`.
2. Implement backend endpoints in `backend/routes` and controller logic in `backend/controllers`:
   - `/auth/google/start`
   - `/auth/google/callback`
   - `/api/v1/users/current`
3. Add DB migrations for `users` under `backend/db/migrations` and run `go run .` locally to verify.
4. Create a minimal frontend login page with "Sign in with Google" button that navigates to `/auth/google/start` and a homepage showing header + hero.
5. Add unit and integration tests for token exchange, upsert logic, and session handling.

## Runbook & Operational Playbook (short)

- How to rotate Google client secret: create new client credentials in Google console, update `GOOGLE_CLIENT_SECRET` in secrets manager, deploy, and verify sign-ins.
- How to debug sign-in failures: check backend logs for token exchange errors, inspect `state` storage, confirm redirect URI configured in Google console matches `OAUTH_REDIRECT_URL`.
- How to recover from duplicate users: safely merge user records by matching `provider_id`; prevent recurrence via unique constraint + upsert.
