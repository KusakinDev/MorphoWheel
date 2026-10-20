# Security policy

## Reporting a vulnerability

Please do not open a public issue for security problems. Report them privately
through GitHub: **Security → Report a vulnerability** on this repository.

Include the affected version or commit, steps to reproduce, and the impact you
expect. You can expect an initial reply within a few days.

## Scope

MorphoWheel runs locally, makes no network requests, and passes your files to
local converters (`ffmpeg`, `pandoc`, `soffice`, `magick`, `sips`, `tar`).
Reports about crafted files that make those invocations misbehave (for example
argument injection through file names) are in scope. Vulnerabilities inside the
converters themselves should go to their own projects.
