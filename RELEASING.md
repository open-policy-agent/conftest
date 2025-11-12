# Releasing

conftest releases in the first week of each month, after the new version of Open
Policy Agent is released. Patch releases are not generally created while we are
on v0, but we may create one if there is a blocking bug in a newly released
feature.

## New release

1. Check for any open
   [pull requests](https://github.com/open-policy-agent/conftest/pulls) that are
   ready to merge, and merge them.

1. Verify that all
   [post-merge CI tasks](https://github.com/open-policy-agent/conftest/actions/workflows/post_merge.yaml)
   have completed successfully.

1. Check out to the master branch and ensure you have the latest changes.

   ```sh
   git checkout master
   git pull
   ```

1. Determine the next version number, and create a tag. You can check the
   [releases](https://github.com/open-policy-agent/conftest/releases) page to
   see the previous version if you do not know it.

   ```sh
   git tag v<VERSION>
   git push --tags
   ```

1. Monitor the
   [release workflow](https://github.com/open-policy-agent/conftest/actions/workflows/release.yaml)
   and verify it does not error. This usually takes ~45min due to slow speeds of
   the Docker cross-compiles.
