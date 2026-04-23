# Post Publishing
## Topic
Signed-in users can publish a post that other readers can see.

## JTBD
- When I have something to share, I want to publish it as a post so other people can read it later.
- When I choose categories, I want the post to appear in those categories so readers can find it in the right places.

## Acceptance Criteria
- A signed-in user can publish a post with a title, body, and one or more categories from the forum's available category set.
- A successful submission creates exactly one new post attributed to the signed-in author.
- After a successful submission, the new post is visible in the forum feed.
- After a successful submission, the new post can be opened on its post detail page.
- The published post shows the categories chosen during submission.
- A post is not created when the title is missing.
- A post is not created when the body is missing.
- A post is not created when no category is selected.
- A post is not created when any selected category is unavailable.
- A signed-out visitor who attempts to publish is sent to login before any post is created.

## Out Of Scope
- Editing a post after publication.
- Creating categories.
- Comments, reactions, moderation, or deletion.
