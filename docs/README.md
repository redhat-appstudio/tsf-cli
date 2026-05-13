# TSF Documentation

This directory contains the [Antora](https://antora.org/)-based documentation for the Trusted Software Factory (TSF).

**Live site**: https://redhat-appstudio.github.io/tsf-cli/

## Branches

| Branch | Purpose |
|--------|---------|
| `main` | Development branch. All new doc work starts here. PRs to `main` validate the build but do not publish. |
| `release-X.Y` | Release branches (e.g., `release-0.1`). The **latest** release branch is the source for the published site. Docs are built and deployed directly from this branch — there is no separate `docs` branch or tag involved. |

## Contributing documentation changes

### Day-to-day workflow

1. Create a feature branch from `main`:
   ```sh
   git checkout -b docs/my-change upstream/main
   ```
2. Edit files under `docs/modules/ROOT/pages/`.
3. Build locally to verify (see [Local build](#local-build) below).
4. Open a PR against `main`. CI runs the `validate` job to check the Antora build.
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

Once the push lands on the latest release branch, CI automatically builds and deploys the site. The live site updates within a few minutes.

## Cutting a new release

When creating a new release branch (e.g., `release-0.2`):

1. **Create the branch** from `main` as usual.
2. **Update `display_version`** in `docs/antora.yml` on the new release branch:
   ```yaml
   display_version: '0.2'
   ```
   This controls the version number shown in the toolbar next to "Edit this Page".
3. **No CI or GitHub settings changes are needed.** The workflow automatically detects the latest `release-*` branch by semver sort. Once `release-0.2` exists, it becomes the publishing source and pushes to `release-0.1` stop publishing.

## How publishing works

Publishing is fully automated via GitHub Actions (`.github/workflows/docs.yml`). The docs source is built directly from the release branch that triggered the workflow — no intermediate branch or tag is involved.

### Workflow triggers

The workflow runs on two events:
- **Pull request to `main`** (only if files in `docs/` or `antora-playbook.yml` changed) — validates the build.
- **Push to any `release-*` branch** — checks if this is the latest release, and if so, builds and deploys.

### Jobs

There are four jobs. On a push to a release branch, they run in sequence: `check` → `build` → `deploy`. On a PR to `main`, only `validate` runs.

#### 1. `check` (push to `release-*` only)

Determines whether the pushed branch is the latest release branch.

- Runs `git ls-remote` to list all `release-*` branches on the remote.
- Sorts them by semver (e.g., `release-0.2` > `release-0.1`).
- Writes the result to `$GITHUB_OUTPUT` so downstream jobs can read it.
- Sets `is-latest=true` if the pushed branch matches, `false` otherwise.

**How the gating works:** The `check` job declares an `outputs` section that exposes `is-latest`. The `build` job has `needs: check` and an `if` condition that reads `needs.check.outputs.is-latest == 'true'`. If the check says `false`, `build` is skipped, and since `deploy` depends on `build`, it is also skipped. This is standard GitHub Actions job chaining — one job writes to `$GITHUB_OUTPUT`, the next reads it via `needs.<job>.outputs.<name>`.

#### 2. `validate` (PRs to `main` only)

- Checks out the PR branch.
- Builds the site with Antora to verify the docs compile without errors.
- Does **not** upload artifacts or deploy. This is a build-only check.

#### 3. `build` (after `check`, only if `is-latest` is `true`)

- Checks out the release branch (the same branch that triggered the workflow).
- Installs Antora 3.1 and the Lunr search extension.
- Runs `antora --fetch antora-playbook.yml` to build the site from the docs source on this branch.
- Creates a `.nojekyll` file (required for GitHub Pages to serve Antora's `_/`-prefixed asset directories).
- Uploads the build output as a GitHub Pages artifact.

**Key point:** The build uses the docs content from the release branch itself. There is no separate `docs` branch or `docs` tag. Whatever is on the release branch at the time of the push is what gets published.

#### 4. `deploy` (after `build` succeeds)

- Deploys the Pages artifact to the `github-pages` environment.
- The live site updates at https://redhat-appstudio.github.io/tsf-cli/.

### Flow diagram

```
PR to main ──> validate (build-only, no deploy)

Push to release-* ──> check ──> is this the latest? ──yes──> build (from release branch) ──> deploy
                                                      ──no──> skip (all remaining jobs skipped)
```

### GitHub Pages configuration

These settings are in the GitHub repo settings and do not need to change per release:

- **Source** (Settings → Pages): GitHub Actions (not "Deploy from a branch").
- **Deployment branches** (Settings → Environments → github-pages): `release-*` must be an allowed branch pattern.

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
        ├── toolbar.hbs                #   Version display + Edit this Page link
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
| `{TSFInstallerImage}` | quay.io/redhat-ads/tsf-cli:latest | Installer container image |
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
