# Feed Filter Semantics
## Topic
Homepage filters narrow the feed by strict intersection.

## JTBD
- When I apply feed filters, I want the visible results to reflect the filter state I chose so I can narrow the feed to what I mean.

## Acceptance Criteria
- Category filtering shows only posts that match every selected category.
- The `my posts` filter shows only posts authored by the current signed-in user.
- The `liked posts` filter shows only posts included by the forum's defined liked-posts behavior for the current signed-in user.
- When multiple filters are active, only posts matching every active criterion remain visible.

## Out Of Scope
- Feed URL restoration and shared-link behavior.
- Post creation, comments, reactions, or moderation actions.
- Search, tags, category directories, or alternate home screens.
