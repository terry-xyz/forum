# Protected Action Return
## Topic
Protected requests return users to the original destination after login.

## JTBD
- When I reach a protected page or action while signed out, I want to continue where I started after I log in.

## Acceptance Criteria
- A signed-out user who requests a protected page or action is sent to login before the protected result is shown or executed.
- The original destination, including query parameters when present, is preserved through the login flow.
- A successful login returns the user to the same protected destination they originally requested.
- A failed or abandoned login does not reach the protected destination.
- The preserved return path is consumed by the successful login flow and does not redirect later unrelated navigation.

## Out Of Scope
- Registration, credential validation, or account recovery.
- Permission rules for signed-in users.
- Navigation to non-protected pages.
