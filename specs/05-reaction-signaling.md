# Reaction Signaling
## Topic
Users can place one visible reaction on a post or comment.

## JTBD
- When I read a post or comment, I want to mark it once so my stance is visible without replying.

## Acceptance Criteria
- Signed-in users can like posts.
- Signed-in users can dislike posts.
- Signed-in users can like comments.
- Signed-in users can dislike comments.
- Signed-out visitors cannot add a reaction.
- A user has at most one active reaction for a given post or comment.
- Changing a reaction leaves exactly one active visible reaction for that user and target.
- Repeated rapid submissions for the same user and target leave one final visible reaction state and correct visible totals.
- Any visitor can see the current reaction totals on posts and comments.
- Visible totals reflect only the active reactions for that target.

## Out Of Scope
- Reaction effects on ranking, sorting, summaries, or notifications.
- Public display of who reacted.
- Extra reaction types beyond like and dislike.
- Undo timers, trust levels, or moderation behavior.
