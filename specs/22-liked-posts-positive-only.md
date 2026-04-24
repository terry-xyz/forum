# Liked Posts Positive-Only Semantics

## Topic
The homepage `liked posts` filter includes only posts the current user has actively liked.

## JTBD
- When I filter to `liked posts`, I want to see the posts I have positively endorsed so the feed matches the plain-language meaning of liked content.

## Acceptance Criteria
- The `liked posts` filter includes a post only when the current signed-in user has an active positive reaction on that post.
- A negative reaction does not make a post appear in `liked posts`.
- Switching a post reaction from positive to negative removes that post from `liked posts`.
- Clearing a positive reaction removes that post from `liked posts`.
- Feed behavior, UI copy, and tests use this same positive-only definition consistently.

## Out Of Scope
- Comment-level liked-content views.
- Ranking or recommendation behavior based on reactions.
- Alternate names for the existing homepage filter.
