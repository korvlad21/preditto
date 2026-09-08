# Skill: Encoding-Safe Editing

## Purpose

Ensures that text edits are saved as valid UTF-8 and prevents mixing different encodings.

## When to Use

- when creating or editing any text file;
- mandatory when modifying Cyrillic or other non-ASCII text;
- mandatory when the file encoding is unclear;
- mandatory when you want to edit a file through the shell instead of using `apply_patch`.

## What It Covers

- checking the current file encoding;
- safely converting legacy-encoded files to UTF-8;
- rules for shell-based edits involving non-ASCII text;
- final strict byte-level validation.

## Workflow

1. First, confirm that the file is a text file and that its encoding is known.
2. If there is any doubt about the encoding, treat the file as unsafe for direct editing.
3. If the file is not valid UTF-8, first deliberately convert the entire file to UTF-8 and verify the result at the byte level.
4. Edit the content only after that. Prefer `apply_patch` for manual edits.
5. If a shell-based edit is unavoidable, pass non-ASCII literals only in an encoding-safe form and immediately perform strict UTF-8 validation after writing.
6. After any edit involving non-ASCII text, validate the resulting file again using a strict UTF-8 decoder.

## What Not to Do

- do not append UTF-8 fragments to a CP1251/ANSI-encoded file;
- do not append CP1251/ANSI fragments to a UTF-8-encoded file;
- do not use `Set-Content`, `Out-File`, `>`, or `>>` for non-ASCII text without explicitly using UTF-8 and performing subsequent validation;
- do not rely solely on how the text appears in the terminal;
- do not continue using `apply_patch` if it fails because of a legacy encoding.

## Related Files

- `AGENTS.md`
