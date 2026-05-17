# Agent Collaboration Rules

These rules apply to every conversation and every coding session in this repository.

## Git And Deployment

- Do not push frontend changes directly to `main`.
- After making frontend changes, run the appropriate local checks and summarize the result for the user.
- Ask the user for explicit approval before pushing frontend changes, especially when the target branch is `main`.
- If the user has not explicitly approved a push, leave the changes local or committed locally only.

## Communication

- Be proactive with implementation and verification, but pause before actions that affect shared remote state.
- Clearly state when code is ready to push and what has been verified.
