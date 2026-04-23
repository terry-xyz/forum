# Write Completion
## Topic
Successful forum writes land on a finished destination that can be revisited without repeating the action.

## JTBD
- When I complete a write successfully, I want to see the finished result so I know the action took effect.
- When I refresh or return to that result, I want the forum to stay on the completed state instead of writing again.

## Acceptance Criteria
- Every successful write ends on a readable destination that reflects the completed change.
- The successful destination can be refreshed or revisited without repeating the write.
- Reloading or replaying the destination does not create duplicate visible posts, comments, reactions, or deletions from the completed action.
- The visible result of the completed write remains stable across navigation back to the success destination.

## Out Of Scope
- Validation rules for write inputs.
- Failed, malformed, or unauthorized requests.
- How writes are stored, processed, or deduplicated internally.
