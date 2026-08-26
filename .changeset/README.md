# Changesets

Every PR with a user-visible change adds one fragment here. `CHANGELOG.md` is written at release time by
folding these fragments into a section for the upcoming tag, after which the folded fragments are deleted.
Do not edit `CHANGELOG.md` in a feature or fix PR.

Filename: descriptive kebab-case, e.g. `clear-rewards-on-maintainer-deregister.md`.

```markdown
---
'@axelar-network/axelar-core': patch
---

One or two sentences describing the change, in the imperative, as it should read in the changelog.
```

Bump level follows the release the change is headed for: `patch` for a bug fix or dependency bump, `minor`
for a new feature or module, `major` for a breaking change. Nothing consumes these fragments
programmatically, so the level is a hint to the release owner rather than an input to version selection.

Purely internal changes (comments, test refactors, CI) do not need a fragment.
