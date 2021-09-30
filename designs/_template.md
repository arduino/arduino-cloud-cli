# Title: Template for design documents

Use this as a template to write a design document when adding new
major features to the project. It helps other developers
understand the scope of the project, validate technical complexity and
feasibility. It also serves as a public documentation of how the feature
actually works.

## Review Period

Best before Month, day year.

## What is the problem?

## Out-of-Scope

## User Experience Walkthrough

## Implementation

### Project Changes

_Explain the changes to api or command line interface, including adding new
commands, modifying arguments etc_

### Breaking Change

_Are there any breaking changes to the interface? Explain_

### Design

_Explain how this feature will be implemented. Highlight the components
of your implementation, relationships_ _between components, constraints,
etc._

### Security

_Tip: How does this change impact security? Answer the following
questions to help answer this question better:_

**What new dependencies (libraries/cli) does this change require?**

**What other Docker container images are you using?**

**Are you creating a new HTTP endpoint? If so explain how it will be
created & used**

**Are you connecting to a remote API? If so explain how is this
connection secured**

**Are you reading/writing to a temporary folder? If so, what is this
used for and when do you clean up?**

### Documentation Changes

_Explain the changes required to internal and public documentation (API reference, tutorial, etc)_

## Additional Notes

_Link any useful metadata: Jira task, GitHub issue, …_
