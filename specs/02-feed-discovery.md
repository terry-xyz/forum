# Feed Discovery
## Topic
The homepage feed remains restorable from its URL.

## JTBD
- When I arrive at the forum, I want to see recent posts first so I can judge what is current quickly.
- When I return to a saved or shared feed link, I want the same visible feed state to appear so I do not need to rebuild it.

## Acceptance Criteria
- Anonymous visitors can open the homepage feed without signing in.
- The default homepage feed shows newer posts before older posts.
- The active feed state is represented by URL query parameters and can be restored from the URL alone.
- Reloading or reopening the same feed URL returns the same visible feed state.
- When no posts match, the feed explains why the result is empty and offers a clear reset path.
- Feed browsing remains usable when client-side enhancements are unavailable.

## Out Of Scope
- Feed filter semantics.
- Registration, login, logout, or session behavior.
- Creating posts, comments, or reactions.
- Post detail pages.
