# Registration Username Uniqueness

## Topic
Registration rejects a username that is already in use by another account.

## JTBD
- When I choose a username during registration, I want to know if it is already taken so I can pick a distinct identity before the account is created.

## Acceptance Criteria
- A registration attempt that uses an existing `username` is rejected and creates no new account.
- A duplicate-username rejection leaves the existing account unchanged.
- Duplicate-username detection is enforced independently of duplicate-email detection.
- The rejection clearly indicates that the username is already taken.

## Out Of Scope
- Username format, normalization, or reserved-name rules.
- Username changes after registration.
- Login identifier rules.
