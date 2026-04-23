# Session Access
## Topic
The forum keeps one valid session per user at a time.

## JTBD
- When I sign in, I want this browser to become my active forum session.
- When I sign in on another browser, I want the older browser to stop working.
- When a session is no longer valid, I want the forum to explain why I need to sign in again.

## Acceptance Criteria
- A successful sign-in makes the current browser the active signed-in session for that user.
- The next protected request from an older browser session is treated as signed out.
- An expired, revoked, superseded, or unknown session does not reach protected content or protected actions.
- A browser with an expired, revoked, superseded, or unknown session is sent to login before the protected request completes.
- A browser with an invalid session is no longer shown as signed in.
- The login flow shows a human-readable explanation when an earlier session is no longer valid.

## Out Of Scope
- Registration.
- Password reset.
- Email verification.
- Multi-factor authentication.
- Account deletion.
- Session controls in user settings.
- Authorization rules beyond login-gated access.
