# Delete Cascade Visibility
## Topic
Deleting content removes dependent artifacts from forum views.

## JTBD
- When I delete a post or comment, I want anything that should no longer be visible because of that deletion to disappear from forum views.

## Acceptance Criteria
- When a post is deleted, it no longer appears in feed views.
- When a post is deleted, opening its former post detail URL no longer shows that post content.
- When a post is deleted, comments that belonged to that post no longer appear in forum views.
- When a post is deleted, category-based views no longer surface that post through earlier category links.
- When a post is deleted, reactions tied to that post no longer appear in visible counts or visible reaction state.
- When a comment is deleted, it no longer appears on the post detail page.
- When a comment is deleted, reactions tied to that comment no longer appear in visible counts or visible reaction state on that post.

## Out Of Scope
- How deletion is stored or propagated internally.
- Restoring deleted content.
- Visibility rules for administrative, diagnostic, or database-level views.
