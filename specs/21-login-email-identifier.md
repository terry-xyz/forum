# Login Email Identifier

## Topic
Returning users sign in with email and password.

## JTBD
- As a returning user, I want to use my registered email and password to sign in so the forum authenticates me the same way registration identifies my account.

## Acceptance Criteria
- The login form accepts `email` and `password`.
- A valid registered `email` plus matching `password` signs the user in on the current browser.
- An unknown `email` or wrong `password` leaves the browser signed out.
- Failed sign-in clearly indicates that the provided credentials were not accepted.

## Out Of Scope
- Password reset or account recovery.
- Alternate identifiers such as username-based login.
- Registration field validation rules beyond what registration already defines.
