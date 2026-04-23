# Logout
## Topic
Logging out ends the current signed-in browser session.

## JTBD
- When I choose to log out, I want the current signed-in state to end so this browser no longer acts as my account.

## Acceptance Criteria
- A signed-in user can log out from the current browser session.
- Successful logout ends the signed-in state for the current browser.
- Later protected requests from the same browser are treated as signed out until a new login occurs.
- Successful logout gives a clear completion outcome.

## Out Of Scope
- Registration or login.
- Session creation, rotation, expiry, or revocation rules.
- Return-path behavior after authentication.
- Account deletion.
- Password reset.
- Multi-factor authentication.
