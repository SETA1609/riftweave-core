# PO_S27_01_indexeddb-theory-and-storage-issues A theory page on IndexedDB, and the storage issues

Date: 2026-11-25

- **Decision** (owner, 2026-11-25): a theory page, `docs/theory/indexeddb.md`, on how the web app should keep data in IndexedDB and what data-oriented design means for it, and story S27 for the changes it calls for.
- **Sources:**
  - The W3C specification (*Indexed Database API 3.0*, `w3c/IndexedDB` at `f9f83f196f85`) and MDN (`mdn/content` at `eeebbb874aeb`), cloned into `cache/docs/` and read there [V 2026-11-25]. `devops-tools fetch-docs github` was refused by the GitHub API (HTTP 403) in this session, so both were cloned through git and copied under `cache/docs/<owner>_<repo>/<commit>/`.
  - Web.dev's IndexedDB best practices and persistent storage pages, Chrome's article on relaxed durability, and Fabian's *Data-Oriented Design*, read on the web [V 2026-11-25].
- **Findings in the code** [V 2026-11-25]:
  - Two databases at version 1, with no indexes; reads are `getAll()`.
  - Autosave rewrites the whole save after every step.
  - No `versionchange` or `blocked` handling, and no `persist()` request.
- **Struct of arrays, where:** in the Go core only for catalog and bestiary lists, behind a benchmark (BE_S27_05). Commands read most fields of few rows, so an array of structs stays. [A, measured in BE_S27_05]
- **Store plan** [A]: `characters` plus `character_summaries`, `profile`, and `drafts` in `riftweave-characters`; `saves` plus `steps` in `riftweave`. The S25 frontend issues now need FE_S27_04, which brings the profile and drafts stores, so neither page needs an upgrade of its own.
- **Open:** whether to merge the two databases into one, so a session's write-back to a character is one transaction [U].
