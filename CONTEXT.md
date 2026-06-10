# PR Tracker

PR Tracker helps a developer decide which pull requests across local repositories need attention.

## Language

**Team Member**:
A GitHub user whose pull requests matter to the developer's immediate review workflow.
_Avoid_: Coworker, teammate

**Bot PR**:
A pull request authored by an automation account rather than a person.
_Avoid_: Automated update, dependency PR

**Bot Author**:
A known automation account; bot authors are not treated as Team Members for filtering purposes.
_Avoid_: Team bot

**My PR**:
A pull request authored by the developer using PR Tracker.
_Avoid_: Outgoing PR

**My PRs Section**:
The dashboard section that shows My PRs and is preserved independently of active pull-request filters.
_Avoid_: Filtered result

**Matching PRs Section**:
The dashboard section that shows pull requests matching the active filters or View.
_Avoid_: Filtered PRs

**Review Decision**:
GitHub's aggregate review state for a pull request, such as approved, changes requested, or review required.
_Avoid_: Approval count, my approval, team approval

**Filter**:
A single independent criterion that includes or excludes pull requests from a View.
_Avoid_: Toggle, option

**Ready PR**:
A pull request that is not a draft and can be reviewed.
_Avoid_: Non-draft PR

**Review-Needed PR**:
A non-draft Team Member pull request whose Review Decision is not approved, including changes-requested and unknown review decisions, and is not a Bot PR.
_Avoid_: Needs My Attention

**Needs My Attention PR**:
A pull request that explicitly requests action from the developer, regardless of whether it was authored by a Team Member.
_Avoid_: Review-Needed PR

**View**:
A named composition of filters for a task-oriented PR dashboard; additional filters narrow a View.
_Avoid_: Profile, saved filter
