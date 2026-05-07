# TSF Documentation

This directory contains the [Antora](https://antora.org/)-based documentation for the Trusted Software Factory (TSF).

**Live site**: https://redhat-appstudio.github.io/tsf-cli/

## Branches

| Branch | Purpose |
|--------|---------|
| `main` | Development branch. All new doc work starts here. |
| `release-X.Y` | Release branches (e.g., `release-0.1`). The **latest** release branch is the source for the published site. |

The `docs` **tag** (not a branch) is managed automatically by CI. Do not create or move it manually.

## Contributing documentation changes

### Day-to-day workflow

1. Create a feature branch from `main`:
   ```sh
   git checkout -b docs/my-change upstream/main
   ```
2. Edit files under `docs/modules/ROOT/pages/`.
3. Build locally to verify (see [Local build](#local-build) below).
4. Open a PR against `main`. CI validates the Antora build.
5. Merge the PR.

Merging to `main` does **not** update the live site.

### Getting changes onto the live site

After merging to `main`, backport the doc changes to the current release branch (e.g., `release-0.1`):

```sh
# Cherry-pick from main to the release branch
git checkout release-0.1
git cherry-pick <commit-hash>
git push upstream release-0.1
```

Or open a separate PR targeting the release branch directly.

Once the push lands on the latest release branch, the live site updates automatically within a few minutes.

### When a new release is cut

When a new release branch is created (e.g., `release-0.2`):

- The new branch automatically becomes the publishing source.
- Pushes to older release branches (e.g., `release-0.1`) stop publishing.
- No configuration or variable updates are needed.

## How publishing works

Publishing is fully automated via GitHub Actions (`.github/workflows/docs.yml`). There are three jobs:

### 1. `tag` job (runs on push to any `release-*` branch)

- Queries all remote `release-*` branches.
- Sorts them by semver to find the latest (e.g., `release-0.2` > `release-0.1`).
- If the pushed branch **is** the latest: force-tags the commit as `docs` and pushes the tag.
- If the pushed branch is **not** the latest: skips silently.

### 2. `build` job (runs on `docs` tag push or PR to `main`)

- Checks out the tagged commit.
- Installs Antora 3.1 and the Lunr search extension.
- Runs `antora --fetch antora-playbook.yml` to build the site.
- Uploads the build output as a GitHub Pages artifact.
- On PRs to `main`, it only validates the build (no artifact upload, no deploy).

### 3. `deploy` job (runs on `docs` tag push, after `build` succeeds)

- Deploys the artifact to GitHub Pages.

```
PR to main ──> build (validate only, no deploy)

Push to release-* ──> tag job ──> is this the latest? ──yes──> force-tag as `docs`
                                                        ──no──> skip

docs tag push ──> build ──> deploy to GitHub Pages
```

## Local build

From the `docs/` directory:

```sh
npm install
npm run build
```

Or from the repo root:

```sh
npx antora generate --clean --fetch antora-playbook.yml
```

The built site is output to `build/site/`. Open `build/site/index.html` in a browser to preview.

## Directory structure

```
docs/
├── antora.yml                         # Component descriptor (name, version, AsciiDoc attributes)
├── package.json                       # Local build dependencies (Antora 3.1, Lunr extension)
├── modules/ROOT/
│   ├── nav.adoc                       # Navigation sidebar (flat single-level list)
│   └── pages/                         # Content pages
│       ├── index.adoc                 #   Home page
│       ├── preparing-to-install.adoc  #   Prerequisites and preparation
│       ├── installing.adoc            #   Installation procedure
│       ├── verifying-and-accessing.adoc # Post-install verification
│       ├── getting-started.adoc       #   First-use walkthrough
│       └── troubleshooting.adoc       #   Common issues and fixes
└── supplemental-ui/                   # UI customizations (override default Antora UI)
    ├── css/main.css                   #   Navbar, sidebar, pagination styles
    └── partials/
        ├── header-content.hbs         #   Navbar with Home link, Raise Issue, search
        ├── head-styles.hbs            #   CSS includes (site.css + main.css)
        ├── footer-content.hbs         #   Footer with GitHub and issue links
        └── pagination.hbs            #   Prev/next page navigation

antora-playbook.yml                    # Antora playbook (lives at repo root, not in docs/)
.github/workflows/docs.yml            # CI workflow for build and deploy
```

## AsciiDoc attributes

Product-specific attributes are defined in `antora.yml` and available in all pages:

| Attribute | Value | Usage |
|-----------|-------|-------|
| `{TSFName}` | Trusted Software Factory | Full product name |
| `{TSFShortName}` | TSF | Short product name |
| `{TSFCli}` | tsf | CLI tool name |
| `{TSFInstallerImage}` | quay.io/redhat-ads/tsf-cli:unstable | Installer container image |
| `{OCPName}` | OpenShift Container Platform | Full OCP name |
| `{OCPShortName}` | OCP | Short OCP name |
| `{OCPVersion}` | 4.20 | Target OCP version |
| `{OCPCli}` | oc | OCP CLI tool |
| `{KonfluxName}` | Konflux | Build system name |
| `{RHTASName}` | Red Hat Trusted Artifact Signer | Signing service name |
| `{RHTASVersion}` | 1.4 | RHTAS version |
| `{RHTPAName}` | Red Hat Trusted Profile Analyzer | SBOM/vulnerability analyzer name |

Use these attributes instead of hardcoding product names, so values stay consistent and easy to update.

## Adding a new page

1. Create a new `.adoc` file in `docs/modules/ROOT/pages/`.
2. Add an `xref` entry in `docs/modules/ROOT/nav.adoc`:
   ```asciidoc
   * xref:my-new-page.adoc[My new page title]
   ```
3. Build locally to verify the page renders and navigation works.
4. Open a PR against `main`.

## Notes

- `antora-playbook.yml` must live at the repo root (not in `docs/`) because it uses `url: .` which resolves relative to the playbook location, and it needs access to the `.git` directory.
- GitHub Pages requires a `.nojekyll` file in the output directory. The CI workflow creates this automatically. Jekyll ignores `_/`-prefixed directories, which Antora uses for all assets.
- Supplemental UI partials **replace** the default Antora UI partials entirely. If you override a partial (e.g., `head-styles.hbs`), you must include everything it needs (e.g., the `site.css` link).
- `nav.adoc` uses a flat single-level list (no nesting) to avoid duplicate "Trusted Software Factory" entries in the sidebar.
