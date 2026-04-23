# Flat Commenting
## Topic
Readers can add a single-level comment to a post without creating reply threads.

## JTBD
- When I am reading a post, I want to add a comment directly to that post so I can respond without creating reply chains.

## Acceptance Criteria
- A signed-in user can add a comment directly to a post.
- A successful submission adds the comment to that post as a top-level comment.
- The comment experience does not offer replies to existing comments.
- No comment is shown as a child of another comment.
- A signed-out visitor who attempts to comment is sent to login before any comment is created.
- A comment is not created when the comment body is empty.
- A failed comment submission leaves the post without any new comment from that submission.

## Out Of Scope
- Comment ordering.
- Comment editing.
- Comment deletion.
- Comment reactions.
