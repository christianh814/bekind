# BeKind Documentation

This directory contains the source files for the BeKind documentation site, built with Jekyll and the just-the-docs theme.

## Local Development

To preview the documentation site locally:

### Prerequisites

- Ruby 3.1 or newer
- Bundler

### Setup

```bash
cd docs
bundle install
```

### Run Locally

```bash
bundle exec jekyll serve
```

Then open [http://localhost:4000/bekind/](http://localhost:4000/bekind/) in your browser.

The site will automatically rebuild when you make changes to the source files.

### Run with Live Reload

```bash
bundle exec jekyll serve --livereload
```

## Documentation Structure

```
docs/
├── _config.yml                    # Jekyll configuration
├── Gemfile                        # Ruby dependencies
├── index.md                       # Homepage
├── installation.md                # Installation guide
├── configuration.md               # Configuration reference
└── features/                      # Feature documentation
    ├── index.md                   # Features overview
    ├── helm-charts.md             # Helm charts feature
    ├── loading-images.md          # Image loading feature
    ├── post-install-manifests.md  # Manifests feature
    └── post-install-actions.md    # Actions feature
```

## Publishing

The documentation site is automatically built and deployed to GitHub Pages when changes are pushed to the `main` branch.

The deployment is handled by the GitHub Actions workflow at `.github/workflows/pages.yml`.

## Theme

This site uses the [just-the-docs](https://github.com/just-the-docs/just-the-docs) Jekyll theme.

For theme customization options, see the [just-the-docs documentation](https://just-the-docs.github.io/just-the-docs/).
