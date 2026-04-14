# MVP Interaction Notes

Scope for the current UI pass:

- Remove the redundant workspace header actions above the answer area.
- Show the user's question immediately when submitting, then fill the answer when the backend returns.
- When creating a child from selected text, submit the question automatically instead of asking the user to type it again.
- Keep the question list model ready for pending / streaming states so the next backend iteration can stream answer text without a redesign.

Deferred for the next pass:

- Create-tree flow should stop asking the user for a title and root question up front.
- The tree title should be derived automatically after the first user question is asked.
