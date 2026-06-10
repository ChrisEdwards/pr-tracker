# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Layout

This is a single-context repo.

Read these files when present:

* `CONTEXT.md` at the repo root
* `docs/adr/` at the repo root

If either path does not exist, proceed silently. Do not flag its absence or suggest creating it upfront. Producer skills such as `/grill-with-docs` create these files lazily when terms or decisions are actually resolved.

## Expected file structure

```text
/
|-- CONTEXT.md
|-- docs/adr/
`-- internal/
```

## Use the glossary's vocabulary

When output names a domain concept in an issue title, refactor proposal, hypothesis, or test name, use the term as defined in `CONTEXT.md`.

If the concept is not in the glossary yet, either the language is being invented too early or the repo has a real terminology gap. Note the gap for `/grill-with-docs`.

## Flag ADR conflicts

If output contradicts an existing ADR, surface it explicitly instead of silently overriding it.
