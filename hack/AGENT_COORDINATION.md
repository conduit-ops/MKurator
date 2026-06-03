# Parallel agent coordination (Kurator)

Agents working on this repo in parallel must read and update
`hack/.agent-coordination.json` so only one push window and one exclusive test
run happen at a time, and pipeline checks do not race pushes.

The **coordinator** agent maintains CI status in `notes` and the final JSON
state after sibling work completes. All other agents follow the rules below.

## Before `git push`

1. Read `hack/.agent-coordination.json` from the **current branch** (prefer
   `origin/main` after `git fetch` if you are about to push to `main`).
2. If `push_allowed` is `false`, wait or poll every **60 seconds** (maximum
   **30 minutes**). Proceed with `git push` only when `push_allowed` is `true`.
3. If you time out after 30 minutes, stop and report in chat; do not force-push.

## Pipeline check (`gh run watch` / `gh run list`)

The agent checking GitHub Actions on `main` must:

1. Set `push_allowed` to `false` and `pipeline_check_in_progress` to `true`.
2. Update `last_updated` (UTC ISO-8601) and append a short status line to
   `notes`.
3. When finished, set `pipeline_check_in_progress` to `false`. Restore
   `push_allowed` to `true` unless the 5-minute post-push cooldown applies (see
   below).

Do **not** run `task cluster:up`, `task test:e2e`, or `task ci:e2e` unless
explicitly assigned that work.

## Exclusive test lock

Path: `hack/kind-cluster/.state/locks/exclusive-test.lock`

When your agent acquires this lock (via `hack/ci/suite-lock.sh` / test tasks):

1. Set `lock_holder` to your agent name (e.g. `agent-e2e-fix`).
2. Clear `lock_holder` to `null` when the lock is released.

If the lock file exists but the PID is dead, treat the lock as stale; clear
`lock_holder` and note stale lock in `notes` before taking the lock.

## Push window (one pusher at a time)

1. Only **one** agent may `git push` per window.
2. Immediately **before** push: confirm `push_allowed` is `true` and you are not
   blocked by `pipeline_check_in_progress` unless you are the coordinator.
3. Immediately **after** a successful push: set `push_allowed` to `false`, set
   `last_updated`, and commit or amend the JSON on the branch you pushed (or
   include the post-push state in the same commit when possible).
4. After **5 minutes**, set `push_allowed` back to `true` (and commit/push that
   JSON update if other agents need it on `main`).

## Renovate and dependency merges

- Do **not** merge Renovate PRs without recording intent in `notes` (PR number,
  branch, and who merged).
- The coordinator does not merge Renovate; it only documents status.

## Coordinator duties

- Poll `gh run list --branch main` every few minutes; refresh `notes` with
  **CI**, **Integration**, **E2E**, and **Release** workflow status.
- Poll `git log origin/main` every **5 minutes** for up to **20 minutes** after
  sibling agents start; when activity quiets and `exclusive-test.lock` is
  absent, set final JSON state (`push_allowed`, flags, `lock_holder`) and push
  updates.
- Initial bootstrap: one small `chore` commit adding this file and the JSON.

## JSON schema

| Field | Type | Meaning |
|-------|------|---------|
| `push_allowed` | boolean | `false` blocks all agents from pushing |
| `pipeline_check_in_progress` | boolean | `true` while an agent runs `gh` pipeline checks |
| `lock_holder` | string or null | Agent name holding `exclusive-test.lock` |
| `last_updated` | string | UTC ISO-8601 timestamp |
| `notes` | string | Free text: CI summary, Renovate, agent handoffs |

Example:

```json
{
  "push_allowed": false,
  "pipeline_check_in_progress": true,
  "lock_holder": null,
  "last_updated": "2026-06-03T02:31:35Z",
  "notes": "CI: success @ abc1234; Integration: failure @ def5678; ..."
}
```
