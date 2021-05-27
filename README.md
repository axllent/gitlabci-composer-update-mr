# Gitlab Composer Updater MR

A series of official PHP Alpine [docker images](https://hub.docker.com/r/axllent/gitlabci-composer-update-mr) containing a utility (`gitlabci-composer-update-mr`) for automated scheduled "`composer update`" merge requests in Gitlab.

This is not a standalone container, and is designed for Gitlab CI integration only.

This binary version of the utility (written in Go) is based on [enomotodev/gitlabci-composer-update-mr](https://github.com/enomotodev/gitlabci-composer-update-mr), however due to PHP version restrictions of the original project it has been completely rewritten. Several extra options have been added too.

## Main features

- Run automated scheduled composer updates on your PHP projects.
- Multiple PHP docker containers available (5.6, 7.0, 7.1, 7.2, 7.3, 7.4 & 8.0).
- Supports both composer 1 & 2 (default 2).
- Identical (open) MRs are detected and ignored (ie: no change since the last MR).
- Replace outdated composer update MRs (default true). Old branches/MRs (that match the same user, and containing the same labels) will be deleted.
- MRs contain a full list of added, updated and deleted packages, linking to version comparisons where possible for each package.
- Auto-assign MR prefix (to suit work flow, eg "feature/")
- Auto-assign MR labels.
- Auto-assign MR to assignees & reviewers.


## Usage

### Setting up GitLab personal access token for GitLabCI

GitLab personal access token is required for sending merge requests to your repository. It may be useful to create a new user for auto updates, and assign them sufficient permissions to read/write to the repository. Gitlab Composer Updater MR will never modify your source branch, and will only (by default) remove old branches of outdated merge requests.

1. Go to your account's settings page and generate a personal access token with `api` scope `Account` -> `Preferences` -> `Access Tokens`
2. In your GitLab dashboard, go to your project's `Settings` -> `CI /CD` -> `Environment variables`
3. Create an environment variable `COMPOSER_MR_TOKEN` with your GitLab personal access token. See [Environment options](#Environment-options) for all options.


### Configure `.gitlab-ci.yml`

Configure your `.gitlab-ci.yml` to run gitlabci-composer-update-mr, for example:

```yaml
stages:
  - composer-update-mr

composer-update-mr:
  stage: composer-update-mr
  image: axllent/gitlabci-composer-update-mr:<php-version>
  only:
    - schedules
  script:
    - gitlabci-composer-update-mr <username> <email> <branch>
```
where:
- `<php-version>` is your php version (see [docker images](https://hub.docker.com/r/axllent/gitlabci-composer-update-mr/tags) for supported versions)
- `<username>` is the git username for the merge request commit
- `<email>` is the email for the merge request commit
- `<branch>` is the branch your wish to work on and create a merge request for

eg: `- gitlabci-composer-update-mr composer-update-mr mr@example.com develop`


### Setting schedule

1. On GitLab dashboard, go to your application's `Schedules` -> `New schedule`
2. Create new schedule and save


## Environment options

The tool has several options which can be configured via Gitlab CI variables, either to your project or alternatively inherited via the group variables. The only required variable is `COMPOSER_MR_TOKEN`, the rest are optional.

|CI environment variable|Default|Description|
--- | :---: | ---
|**`COMPOSER_MR_TOKEN`**       |      |User token for merge requests (required)            |
|`COMPOSER_MR_COMPOSER_VERSION`|`2`   |Composer version (1 or 2)                           |
|`COMPOSER_MR_BRANCH_PREFIX`   |      |MR branch prefix                                    |
|`COMPOSER_MR_LABELS`          |      |MR labels (comma-separated)                         |
|`COMPOSER_MR_ASSIGNEES`       |      |MR assignees (comma-separated usernames)            |
|`COMPOSER_MR_REVIEWERS`       |      |MR reviewers (comma-separated usernames)            |
|`COMPOSER_MR_REPLACE_OPEN`    |`true`|Replace outdated open composer-update merge requests|
