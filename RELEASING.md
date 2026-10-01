# Releasing

conftest releases in the first week of each month, after the new version of Open
Policy Agent is released. Patch releases are not generally created while we are
on v0, but we may create one if there is a blocking bug in a newly released
feature.

## Release branches

Tags MUST be created against the branch for the version being released:

| Version          | Branch          |
| ---------------- | --------------- |
| v0.x             | `releases/0.x`  |
| v1.x and later   | `master`        |

`master` carries the in-progress v1 code, so a v0.x tag created there would
publish v1 code as a v0 release. Confirm the branch before you tag.

## New release

1. Check for any open
   [pull requests](https://github.com/open-policy-agent/conftest/pulls) that are
   ready to merge, and merge them.

1. Verify that all
   [post-merge CI tasks](https://github.com/open-policy-agent/conftest/actions/workflows/post_merge.yaml)
   have completed successfully.

1. Check out the release branch for the version you are releasing and ensure you
   have the latest changes.

   ```sh
   git checkout releases/0.x # master for v1.x and later
   git pull
   ```

1. Determine the next version number, and create a tag. You can check the
   [releases](https://github.com/open-policy-agent/conftest/releases) page to
   see the previous version if you do not know it. Verify you are on the right
   branch first, as the tag points at whatever is checked out.

   ```sh
   git branch --show-current
   git tag v<VERSION>
   git push origin v<VERSION>
   ```

1. Monitor the
   [release workflow](https://github.com/open-policy-agent/conftest/actions/workflows/release.yaml)
   and verify it does not error. This usually takes ~45min due to slow speeds of
   the Docker cross-compiles.
