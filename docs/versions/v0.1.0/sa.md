# System Analysis — Prein v0.1.0

Last updated: 2025-10-13

This document expands the Product Requirements Document (PRD) for Prein v0.1.0 into a technical System Analysis (SA). It translates the high-level requirements into scope, architecture, data models, API endpoints, authentication flow, non-functional considerations, acceptance criteria, risks, and a testing strategy.

## Objective

Deliver a minimal, secure, and maintainable implementation of Prein v0.1.0 that satisfies the PRD goals:
- Provide Google Identity (OAuth) login and persist Google account ID and source.
- Offer a homepage (post-login redirect) containing a header with logo and a hero section.

## Scope and Constraints

In-scope for v0.1.0:
- Google sign-in (OAuth2 / OIDC) integration for user authentication.
- Backend endpoints to create and retrieve user records that include google account ID and source.
- Minimal frontend pages: login flow and homepage (header + hero).
- Storage of users metadata.

Out of scope for v0.1.0:
- Full resume editor features, multi-language UI, advanced authorization rules, and external integrations beyond Google login.

Assumptions:
- PostgreSQL will be used.
- The backend is a Go Gin application; frontend is Next.js.
- HTTPS termination and OAuth redirect URIs will be configured in deployment (development can use localhost with appropriate OAuth client settings).

## Stakeholders

- Product owner: defines acceptance criteria (see PRD).
- Engineers (frontend/backend): implement OAuth flow, persist user data, and create homepage UI.
- QA: verify auth, data persistence, and redirect behavior.

## High-level Architecture

- Frontend: Next.js app handles UI and initiates sign-in. It will either use Google Identity client library (for OIDC on client) or redirect to backend to start OAuth flow. For v0.1.0, prefer server-assisted OAuth to ensure backend records the Google account reliably.
- Backend: Gin-based API that implements OAuth callback handling, user persistence, and simple session management for protected routes.
- Database: PostgreSQL to store `users` and `resumes` tables.

Flow:

1. User clicks "Sign in with Google" on frontend.
2. Frontend redirects to backend endpoint `/auth/google/start` which starts OAuth by redirecting to Google's OAuth endpoint with required scopes and a CSRF state.
3. Google redirects back to backend `/auth/google/callback` with `code` and `state`.
4. Backend exchanges `code` for tokens, obtains the user's Google profile (sub = Google account ID, email, name, picture).
5. Backend upserts a `users` record with fields including `google_id` and `source = "google"` and issues a server session or JWT.
6. User is redirected to frontend homepage with session cookie.

## Data Models

Minimal models needed for v0.1.0. The examples use typical column names suitable for sqlc/pgx.

User
- id: UUID (PK)
- provider_id: TEXT (unique, not null)
- email: TEXT
- name: TEXT
- family_name: TEXT
- given_name: TEXT
- avatar_url: TEXT
- source: TEXT (e.g., "google")
- created_at: TIMESTAMP
- updated_at: TIMESTAMP

Indexes & constraints
- Unique index on `users.provider_id`

## API Endpoints

Note: Routes described are backend/Gin endpoints. They return JSON and use typical HTTP status codes.

Public
- GET /health
	- Returns 200 OK (service health).

Auth flow (server-side)
- GET /auth/google/start
	- Redirects to Google's OAuth2 consent screen. Generates and stores CSRF `state` in a short-lived store (in-memory) tied to the user's session.
- GET /auth/google/callback
	- Accepts `code` and `state`. Exchanges code for access token, fetches user profile, upserts user record (persist `provider_id` and `source`), then issues a session cookie and redirects to frontend homepage (`/`).

User API (authenticated)
- GET /api/v1/users/current
	- Returns current user profile (id, provider_id, email, name, avatar_url, source).

Security and session management
- For v0.1.0 prefer secure, HttpOnly session cookies issued by backend to simplify client state and avoid storing secrets in the browser.

## Error Cases and Edge Conditions

- OAuth errors: invalid state, expired code, token exchange failure. Return informative errors and log telemetry.
- Duplicate provider_id: Upsert semantics must avoid creating duplicate users on repeat sign-ins.
- Missing profile fields: Google's profile may omit email; define fallback behavior (reject login or create partial record and prompt for email). For v0.1.0 prefer requiring email scope and failing if email is absent.

## Acceptance Criteria (mapped to PRD)

- Google Identity Integration:
	- Users can sign in with Google.
	- Backend stores the Google account ID (`provider_id`) and `source` = "google" for each user.
	- Repeated sign-ins do not create duplicate user records.

- Homepage:
	- After successful sign-in, users are redirected to the homepage.
	- Homepage displays header with logo and a hero section.

Verification steps (QA):

1. Start app locally, visit the frontend.
2. Click Sign-in with Google, complete consent.
3. Confirm backend has a `users` record with `provider_id` and `source="google"`.
4. Confirm frontend lands on homepage and shows header + hero.

## Non-functional Requirements

- Security: use HTTPS in production; use secure HttpOnly cookies; validate OAuth state to prevent CSRF.
- Privacy: store only minimally required profile fields; encrypt or protect tokens at rest.
- Observability: log auth events (success/failure), capture user creation events.
- Performance: v0.1.0 is low-traffic; simple in-memory session store is acceptable for dev, use Redis for production.

## Testing Strategy

Unit tests
- Backend: token exchange handling, user upsert logic, error handling for missing fields.

Integration tests
- Simulate OAuth callback with a stubbed Google token exchange (or use Google's test credentials) and verify user persistence and redirect.

End-to-end (E2E)
- Use a headless browser (Playwright or Cypress) to run a login flow against a test Google OAuth client configured for localhost. Verify redirect and homepage UI.

Manual QA
- Verify Google account ID and source are stored in the DB after sign-in.

Suggested test cases

1. Happy path: Sign in with Google, user created, redirected to homepage.
2. Repeat sign-in: Sign in again with same Google account, backend re-uses existing user (no duplicate).
3. Invalid state: Callback with wrong state must be rejected.
4. Missing email: If Google doesn't return email, login fails with a clear error.

## Risks & Mitigations

- Risk: OAuth configuration issues (redirect URIs, client secrets) during development.
	- Mitigation: Provide step-by-step README for creating Google OAuth credentials and a `.env.example` with required variables.

- Risk: Session management inconsistencies across frontend/backend domains.
	- Mitigation: For development, host frontend and backend under the same top-level domain or use secure CORS and cookie settings; document required cookie settings.

- Risk: Duplicate users due to race conditions on upsert.
	- Mitigation: Use database unique constraint on `google_id` and implement upsert/update logic in a transaction.

## Operational Concerns

- Deployment: Ensure OAuth client uses correct production redirect URIs. Rotate client secrets carefully.
- Monitoring: Alert on repeated auth failures, and log unusual spikes in user creation.

## Implementation Notes & Next Steps

Minimal implementation plan (small steps):

1. Add environment variables and `.env.example` entries for `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `OAUTH_REDIRECT_URL`, `SESSION_SECRET`.
2. Implement backend routes `/auth/google/start` and `/auth/google/callback` in Gin; add `users` table migration.
3. Implement `GET /api/v1/users/current` to return current user.
4. Implement frontend Sign-in button that calls `/auth/google/start` and homepage UI.
5. Add unit and integration tests for auth flow and user persistence.

Optional enhancements (out of scope for 0.1.0 but low risk):

- Use OpenID Connect (OIDC) ID token validation instead of fetching profile via People API.
- Add Redis-backed session store for production.
