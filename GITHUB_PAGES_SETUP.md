# GitHub Pages Setup Instructions

## What Has Been Created

Your documentation site has been set up with the following:

### Documentation Structure
- `docs/` - Main documentation directory
  - `_config.yml` - Jekyll configuration with just-the-docs theme
  - `Gemfile` - Ruby dependencies
  - `index.md` - Documentation homepage
  - `installation.md` - Installation guide
  - `configuration.md` - Configuration reference
  - `features/` - Feature-specific documentation
    - `index.md` - Features overview
    - `helm-charts.md` - Helm charts documentation
    - `loading-images.md` - Image loading documentation
    - `post-install-manifests.md` - Manifests documentation
    - `post-install-actions.md` - Actions documentation
  - `.gitignore` - Ignore Jekyll build artifacts
  - `README.md` - Local development instructions

### GitHub Actions Workflow
- `.github/workflows/pages.yml` - Automated deployment workflow
  - Only runs on pushes to `main` branch
  - Only runs when `docs/` or the workflow file changes
  - Builds Jekyll site and deploys to GitHub Pages

### Updated README
- `README.md` - Simplified to link to the documentation site

## Steps to Enable GitHub Pages

### 1. Commit and Push Your Changes

```bash
cd /home/chriher3/workspace/git/bekind

# Check what's been created
git status

# Add all the new files
git add .

# Commit
git commit -m "Add documentation site with just-the-docs theme"

# Push to your branch
git push
```

### 2. Enable GitHub Pages via gh CLI

```bash
# Navigate to the repository
cd /home/chriher3/workspace/git/bekind

# Enable GitHub Pages to use GitHub Actions as the source
gh api repos/christianh814/bekind/pages \
  --method POST \
  -f source[branch]=main \
  -f build_type=workflow
```

If the above command fails (Pages might already be partially configured), you can update the settings:

```bash
gh api repos/christianh814/bekind/pages \
  --method PUT \
  -f build_type=workflow
```

### 3. Alternative: Enable via GitHub Web UI

If you prefer using the web interface:

1. Go to https://github.com/christianh814/bekind/settings/pages
2. Under "Build and deployment":
   - Source: Select "GitHub Actions"
3. Save the changes

### 4. Merge to Main Branch

Once your changes are reviewed and ready:

```bash
# If you're on a feature branch, merge to main
git checkout main
git merge your-feature-branch
git push origin main
```

The GitHub Actions workflow will automatically:
- Detect the push to `main`
- Build the Jekyll site
- Deploy to GitHub Pages

### 5. Verify Deployment

After the workflow completes (usually 2-5 minutes):

1. Check the Actions tab: https://github.com/christianh814/bekind/actions
2. Visit your documentation site: https://christianh814.github.io/bekind/

## Testing Locally Before Pushing

To preview the documentation site locally:

```bash
cd docs
bundle install
bundle exec jekyll serve
```

Then visit: http://localhost:4000/bekind/

## Troubleshooting

### If the workflow fails:

1. Check the Actions tab for error messages
2. Verify the Gemfile and _config.yml are valid
3. Ensure Pages is enabled in repository settings

### If Pages isn't deploying:

1. Go to Settings → Pages in your repository
2. Ensure "Source" is set to "GitHub Actions"
3. Check that the workflow has "pages: write" permissions

### If links are broken:

The site uses `baseurl: "/bekind"` which matches your repository name. If your repo name is different, update `docs/_config.yml`:

```yaml
baseurl: "/your-repo-name"
```

## Maintenance

### Adding New Documentation Pages

1. Create a new `.md` file in `docs/` or `docs/features/`
2. Add front matter:
   ```yaml
   ---
   layout: default
   title: Your Page Title
   parent: Features  # if it's under features
   nav_order: 5
   ---
   ```
3. Write your content
4. Commit and push to `main`

### Updating Existing Pages

Just edit the `.md` files and push to `main`. The site will automatically rebuild.

### Theme Customization

To customize the just-the-docs theme, see:
- https://just-the-docs.github.io/just-the-docs/docs/customization/

You can override colors, fonts, and layouts by creating custom SCSS files.

## Next Steps

1. ✅ Commit all the new files
2. ✅ Enable GitHub Pages (via `gh` CLI or web UI)
3. ✅ Merge to `main` branch (if not already there)
4. ✅ Wait for GitHub Actions to deploy
5. ✅ Visit https://christianh814.github.io/bekind/
6. 🎉 Share your documentation!

---

**Note**: The workflow is configured to only run when:
- Changes are pushed to the `main` branch
- Changes affect files in `docs/` or the workflow file itself

This prevents unnecessary builds and saves GitHub Actions minutes.
