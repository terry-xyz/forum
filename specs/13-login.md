# Login

## Topic
Returning users can sign in with accepted credentials.

## JTBD
- As a returning user, I want to sign in with my credentials so the forum recognizes me as authenticated.

## Acceptance Criteria
- Valid credentials sign the user in on the current browser.
- Successful sign-in leaves the current browser in an authenticated state.
- Invalid credentials leave the current browser signed out.
- Failed sign-in shows a clear outcome that the credentials were not accepted.
- A failed sign-in does not create a signed-in session for that browser.

## Out Of Scope
- Registration.
- Logout.
- Password reset or account recovery.
- Session replacement when a user signs in again.
- Broader authentication lifecycle behavior beyond login success and login failure.
