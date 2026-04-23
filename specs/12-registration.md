# Registration

## Topic
Visitors can create a forum account by submitting the required registration fields.

## JTBD
- When I want to join the forum, I want to submit my email, username, and password so I can create an account.

## Acceptance Criteria
- Registration requires an email, a username, and a password.
- A submission with all required fields and an unused email creates exactly one new account.
- A successful registration gives the visitor a clear completion outcome for the new account.
- A registration attempt that uses an existing email is rejected and leaves the existing account unchanged.
- A registration attempt missing any required field is rejected and creates no account.

## Out Of Scope
- Login or logout behavior.
- Session creation or persistence.
- Password recovery or reset flows.
- Email verification or confirmation messages.
- Username uniqueness rules.
- Profile editing after registration.
