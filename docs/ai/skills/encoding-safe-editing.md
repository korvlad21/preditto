# Encoding-safe editing

Use when creating or editing text. The UTF-8 requirement is defined in [AGENTS.md](../../../AGENTS.md).

1. Strictly decode existing target bytes as UTF-8 before editing. If decoding fails, identify the actual encoding and deliberately convert and verify the whole file before changing content.
2. Prefer `apply_patch`; use `Update File` for an existing path rather than deleting and adding that same path in one patch.
3. For scripted edits, explicitly specify UTF-8 for reads/writes. Avoid shell interpolation of document contents; use a quoted heredoc or a structured tool argument.
4. Validate resulting bytes with a strict UTF-8 decoder and review the diff. Terminal appearance alone is not an encoding check.

Do not mix encodings or rely on platform-default encodings in redirection, PowerShell or scripts. A failed patch is a reason to inspect the actual file/encoding, not to repeat the same attempt.
