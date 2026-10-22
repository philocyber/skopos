# Skopos provenance and modifications


Skopos is maintained by PhiloCyber at [Philocyber/skopos](https://github.com/Philocyber/skopos). It derives its original code and architecture from [Autumn-27/ARTEX](https://github.com/Autumn-27/ARTEX), based on commit `160fe13`.

The Skopos modifications dated **2026-10-02** cover the product name, English-default/Korean-selectable interface and messages, documentation and bundled agent guidance, setup/build/start/update/password-reset prompts, report and agent language handling, and distribution naming. The executable is `skopos` (`skopos.exe` on Windows); archive names follow `skopos-<version>-<os>-<arch>.zip`. Docker Compose builds the checked-out source locally. The configured release workflow targets this repository and GHCR.

The original `LICENSE` is retained unchanged. Original authorship, contributor acknowledgments and upstream changelog history remain attributed to ARTEX. The Go module is `github.com/Philocyber/skopos`, with source entry point `./cmd/skopos`. Configuration/environment variables use the `SKOPOS_*` prefix, and the default database and fixed login identity are `skopos`/`SKOPOS`; no `ARTEX_*` or `artex`-named runtime identifier remains.

This modification date is not a release tag. It does not assert that binary releases or container images have been published. Historical validation documents under `sidequestion/` describe upstream runs; they are not new Skopos verification receipts. Translated historical JSON preserves technical identifiers and measurements while labeling translated narrative.

Additional modifications on **2026-10-03** complete the server translation audit and add GLM-5.3 configuration templates plus its required always-enabled reasoning compatibility. Templates use the existing provider adapter and do not create or activate a profile automatically.

Further modifications on **2026-10-21** add Vercel AI Gateway and Amazon Bedrock API-key templates, pre-flight endpoint validation and failure hints. They reuse the existing Anthropic and OpenAI formats and add no new wire adapter.
