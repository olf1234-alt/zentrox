# Contributing

- Thanks for your interest! Please:
  - Run `make test` & `make lint` before submitting PRs
  - Keep public API changes documented in `README.md`

- With issues:
  - Use the search tool before opening a new issue.
  - Please provide source code and commit sha if you found a bug.
  - Review existing issues and provide feedback or react to them.

- With pull requests:
  - Open your pull request against `main`
  - Your pull request should have no more than two commits, if not you should squash them.
  - It should pass all tests in the available continuous integration systems such as GitHub Actions.
  - You should add/modify tests to cover your proposed code changes.
  - If your pull request contains a new feature, please document it on the README.
  
## Required conventions

Branch names **must** follow the following prefixes and be separated by `/`:

- `feat/<name>` – New feature
- `fix/<name>` – Bug fix
- `chore/<name>` – Minor / tool upgrade / non-business logic impact

**Regex:**
^(feat|fix|chore)/[a-z0-9._-]+$

**Valid examples**
- `feat/update-context`
- `fix/order-calc`
- `chore/ci-cd-cache`

**INVALID examples**
- `feature/update-context`
- `Feat/Update`
- `feat/` (empty)
- `hotfix/issue-1` (not included in allowed prefixes)

## How to create branch

```bash
# Get new from main (or develop)
git checkout main && git pull

# Create branch in correct format
git checkout -b feat/update-context
git push -u origin feat/update-context