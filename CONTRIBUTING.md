# Contributing

Thanks for helping build the DCGG Intelligence Platform.

## Proposing Changes

1. Open an issue describing the problem or feature (or find an existing one).
2. Fork / branch from `main`:
   - `feat/<scope>` — new capability
   - `fix/<scope>` — bug fix
   - `docs/<scope>` — docs only
   - `chore/<scope>` — tooling, deps, refactor
3. Implement, run `make lint test typecheck` locally.
4. Open a PR against `main`.

## Commit Style

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(radar): add jurisdiction filter
fix(auth): refresh JWT before expiry
docs(readme): clarify MinIO bucket layout
chore(deps): bump pgx to v5.6
```

Scope is usually the module or package (`radar`, `auth`, `members`, `api`, `ui`, ...).

## PR Checklist

- [ ] Branch name follows `type/scope`.
- [ ] Conventional-commit messages.
- [ ] `make lint` passes.
- [ ] `make test` passes.
- [ ] `make typecheck` passes.
- [ ] DB changes include a migration in `backend/migrations/`.
- [ ] New env vars added to `.env.example` and README.
- [ ] Docs updated (README / ARCHITECTURE) if behavior or surface changed.
- [ ] No secrets, PII, or real HR data in fixtures.

## Review

PRs require green CI and at least one approving review before merge. Prefer squash-merge with a conventional-commit title.
