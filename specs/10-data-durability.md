# Data Durability
## Topic
Forum data survives container replacement when the same persistent storage is reused.

## JTBD
- When I replace a running container with another one using the same persistent storage, I want the forum to open with the same data instead of starting over.

## Acceptance Criteria
- Replacing a running container with another one that reuses the same persistent storage does not reset the forum to an empty state.
- Users, sessions, categories, posts, comments, and reactions that existed before replacement remain available afterward.
- A browser session that was valid before replacement continues to work afterward while its stored session remains present.

## Out Of Scope
- Backup, restore, or data migration workflows.
- Durability guarantees without persistent storage.
