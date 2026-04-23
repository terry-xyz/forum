# Error Feedback
## Topic
The forum shows safe error outcomes when a request cannot be completed.

## JTBD
- When a requested resource is missing, I want to see that it is unavailable so I understand why it did not load.
- When my input is invalid, I want to see that the request was rejected so I know it needs correction.
- When the system cannot complete a request, I want to see a generic failure outcome so I know it did not succeed.

## Acceptance Criteria
- A request for a missing forum page, post, or comment returns `404` and shows that the resource is unavailable.
- A malformed or incomplete request on routes that validate input returns `400` where applicable and shows that the request was rejected.
- An unexpected server failure returns a generic `500`-class outcome.
- User-visible error responses do not expose internal diagnostics.
- Error outcomes remain understandable without revealing implementation details.

## Out Of Scope
- Detailed logging, monitoring, tracing, or alerting behavior.
- Recovery tooling for operators or administrators.
- Custom failure flows for every possible status category.
