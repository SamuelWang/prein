# Implementation Todo — Prein v0.1.0

Last updated: 2025-10-13

This actionable todo list is derived from the System Design ([sd.md](./sd.md)) for Prein v0.1.0. Tasks are organized by priority and grouped into frontend, backend, DB, testing, and operations. Each task includes a short description, acceptance criteria, and dependencies.

## Progress checklist

Use this checklist to trace implementation progress. Update the checkboxes as work is completed.

- [x] 1. Add environment example file
- [x] 2. DB migrations: `users`
- [ ] 3. Backend — session store (in-memory)
- [ ] 4. Backend — `/auth/google/start` route
- [ ] 5. Backend — `/auth/google/callback` route
- [ ] 6. Backend — `GET /api/v1/users/current`
- [ ] 7. Frontend — Login page and button
- [ ] 8. Frontend — Homepage with header & hero
- [ ] 9. Unit tests for token exchange and upsert
- [ ] 10. Integration test for OAuth callback

## Priority 1 — Minimal Auth + Homepage (MVP)

1. Add environment example file

- Description: update `backend/.env.example` with `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `OAUTH_REDIRECT_URL`, `SESSION_SECRET`.
- Acceptance criteria: File exists and documents required env variables and example values or placeholders.
- Dependencies: none

2. DB migrations: `users`

- Description: Add SQL migration files under `backend/db/migrations` for creating `users` table (DDL from [sd.md](./sd.md)).
- Acceptance criteria: Migration files present; running migrations (locally) creates tables.
- Dependencies: `backend` project build, `golang-migrate` setup

3. Backend — session store (in-memory)

- Description: Implement a simple session store module in `backend` that supports create/get/delete sessions and TTL cleanup.
- Acceptance criteria: Sessions can be created and retrieved; expired sessions are cleaned; unit tests exist.
- Dependencies: `users` table available for session creation tests

4. Backend — `/auth/google/start` route

- Description: Implement the route that generates `state`, stores it, and redirects to Google's OAuth 2.0 endpoint.
- Acceptance criteria: Visiting `/auth/google/start` redirects to Google with correct query params and generated `state` stored server-side.
- Dependencies: session store or `state` store module

5. Backend — `/auth/google/callback` route

- Description: Implement callback handling: validate `state`, exchange `code` for tokens (or validate ID token), extract profile, upsert user, create session cookie, and redirect to frontend home.
- Acceptance criteria: Successful OAuth flow creates or updates a `users` row and sets `prein_sess` cookie; redirects to `/` on success. Handles error cases.
- Dependencies: Google OAuth credentials, state store, DB migrations, session store

6. Backend — `GET /api/v1/users/current`

- Description: Return current authenticated user's profile from session cookie.
- Acceptance criteria: Authenticated requests return user JSON; unauthenticated requests return 401.
- Dependencies: session store, user persistence

7. Frontend — Login page and button

- Description: Add a minimal login page in the Next.js frontend that navigates to `GET /auth/google/start` when the user clicks "Sign in with Google".
- Acceptance criteria: Button triggers redirect to backend start endpoint; behavior works in local dev.
- Dependencies: Backend start endpoint

8. Frontend — Homepage with header & hero

- Description: Implement a simple homepage showing a header (logo) and hero section, and fetch `GET /api/v1/users/current` to show user info if logged in.
- Acceptance criteria: Homepage renders header/hero; when logged in shows user name/avatar.
- Dependencies: `/api/v1/users/current`

## Priority 2 — Hardening, Tests, Observability

9. Unit tests for token exchange and upsert

- Description: Add unit tests covering token exchange parsing (mock HTTP), ID token validation logic, and upsert DB logic.
- Acceptance criteria: Tests run and pass locally; critical branches covered.
- Dependencies: Implementation of callback, DB schema

10. Integration test for OAuth callback

- Description: Add an integration test that spins up backend with a fake Google OAuth server or stubbed HTTP responses to assert full callback -> user upsert -> session cookie flow.
- Acceptance criteria: Integration test runs locally and asserts DB row created and cookie set.
- Dependencies: Backend auth implementation, test harness
