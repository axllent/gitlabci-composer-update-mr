# GitLab Composer Updater MR

A series of official PHP [docker images](https://hub.docker.com/r/axllent/gitlabci-composer-update-mr) containing a utility (`gitlabci-composer-update-mr`) for automated scheduled `composer update` merge requests in GitLab.

This docker image is designed for GitLab CI integration only and cannot be used as a stand-alone utility.

This binary version of the utility (written in Go) is based on [enomotodev/gitlabci-composer-update-mr](https://github.com/enomotodev/gitlabci-composer-update-mr), however due to PHP version restrictions of the original project it has been completely rewritten. Several extra options have been added too.


## Main features

- Run automated scheduled composer updates on your PHP projects.
- Multiple PHP docker containers available (5.6, 7.0, 7.1, 7.2, 7.3, 7.4 & 8.0).
- Supports both composer 1 & 2 (default 2) - binaries are named `composer-1` & `composer-2`.
- Identical open MRs are detected and ignored (ie: no change since the last MR).
- Replace outdated composer update MRs (default `true`). Old branches/MRs (that match the same user, and containing the same labels) will be deleted when a updated MR is generated.
- MRs descriptions contain a full list of added, updated and deleted packages, linking to version comparisons where possible for each package.
- Auto-assign MR prefix (to suit work flow, eg "feature/").
- Auto-assign MR labels.
- Auto-assign MR to assignees & reviewers. Note: assigning multiple assignees/reviewers is a GitLab premium feature, see [Environment variable notes](#environment-variable-notes) below.


## Usage

### Setting up GitLab personal access token for GitLabCI

A GitLab personal access token is required for creating merge requests in your repository. It may be useful to create a new user specifically for auto updates,assigning them sufficient permissions to read/write access to the repository. GitLab Composer Updater MR will never modify your source branch, and will only (by default) remove old branches of outdated open merge requests that were previously created by the same user (the owner of the token).

1. Go to your (or the user you created) account's settings page and generate a personal access token with `api` scope `Account` -> `Preferences` -> `Access Tokens`
2. In your GitLab dashboard, go to your project's `Settings` -> `CI /CD` -> `Environment variables`. This can also be set on a group level if you prefer.
3. Create an environment variable `COMPOSER_MR_TOKEN` with the GitLab personal access token from step #1. See [Environment options](#environment-options) for all configuration options.
4. Make sure you have merge requests enabled for your project.


### Configure `.gitlab-ci.yml`

Configure your `.gitlab-ci.yml` to run `gitlabci-composer-update-mr`. You need to specify three arguments after `gitlabci-composer-update-mr`, namely the git commit user & email for the git log, and the source branch:

```yaml
stages:
  - composer-update-mr

composer-update-mr:
  stage: composer-update-mr
  image: axllent/gitlabci-composer-update-mr:<php-version>
  only:
    - schedules
  script:
    - gitlabci-composer-update-mr <commit-user> <commit-email> <source-branch>
```

where:
- `<php-version>` is your php version (see [docker images](https://hub.docker.com/r/axllent/gitlabci-composer-update-mr/tags) for supported versions)
- `<commit-user>` is the git commit username for the merge request commit
- `<git-email>` is the git commit email for the merge request commit
- `<source-branch>` is the branch your wish to work on and create a merge request for

eg: `- gitlabci-composer-update-mr composer-update-mr mr@example.com develop`

Regardless what branch your schedule is set to run on, the latest `<source-branch>` will always be used.


### Setting schedule

1. On GitLab dashboard, go to your application's `Schedules` -> `New schedule`
2. Create new schedule and save


## Environment options

The tool has several options which can be configured via GitLab CI variables, either to your project or alternatively inherited via the group variables. The only required variable is `COMPOSER_MR_TOKEN`, the rest are optional.

|CI environment variable|Default|Description|
--- | :---: | ---
|**`COMPOSER_MR_TOKEN`**       |      |**User token for merge requests (required)**        |
|`COMPOSER_MR_COMPOSER_VERSION`|`2`   |Composer version (1 or 2)                           |
|`COMPOSER_MR_BRANCH_PREFIX`   |      |MR branch prefix, eg "feature/"                     |
|`COMPOSER_MR_LABELS`          |      |MR labels (comma-separated)                         |
|`COMPOSER_MR_ASSIGNEES`       |      |MR assignees (comma-separated usernames)            |
|`COMPOSER_MR_REVIEWERS`       |      |MR reviewers (comma-separated usernames)            |
|`COMPOSER_MR_REPLACE_OPEN`    |`true`|Replace outdated open composer-update merge requests|


## Environment variable notes

### `COMPOSER_MR_COMPOSER_VERSION`

Currently both composer 1 & 2 (default) are provided, and are current at the time of the docker build. If you require the very latest composer version you can always run `composer-1 self-update` or `composer-2 self-update` as part of your CI process prior to the `gitlabci-composer-update-mr` command.


### `COMPOSER_MR_BRANCH_PREFIX`

Merge request branches are named similar to `composer-update-<utc-date>`, eg: `composer-update-20210527083313`. You can add a prefix to these branches, for instance `COMPOSER_MR_BRANCH_PREFIX` => `feature/` which will in future create the branches like `feature/composer-update-20210527083313` (for instance for use within git flow).


### `COMPOSER_MR_LABELS`

You can set as many labels as you like simply by comma-separating the environment value, eg `COMPOSER_MR_LABELS` => `Composer Update, Auto`. Please note that these labels are used to search for previous merge requests, so if you edit or add labels, prior merge requests may get ignored.


### `COMPOSER_MR_ASSIGNEES`/`COMPOSER_MR_REVIEWERS`

Comma-separate usernames to assign to either merge request assignees or reviewers. These users must have access to the project or exist, otherwise they are silently ignored from the merge request.

Please note that multiple assignees/reviewers is a [GitLab premium feature](https://docs.gitlab.com/ee/user/project/issues/multiple_assignees_for_issues.html) and is [not currently supported](https://gitlab.com/gitlab-org/gitlab/-/issues/22171) in the Community Edition of GitLab. If you have assigned multiple users and you are using the Community Edition, then just the first user is assigned.


### `COMPOSER_MR_REPLACE_OPEN`

GitLab Composer Updater MR will always add a checksum of the `composer.lock` to any merge request to allow comparison. Upon update, if an open merge request is found with a matching checksum, then the current update is skipped.

If no matching checksum in a merge request is found, then any previous outdated **open** merge request is closed and their branches removed. If you do not want this behavior then set this environment variable to `false`.

In both instances, merge requests must match the same labels (if set), created by the same user (that owns the `COMPOSER_MR_TOKEN`), and have a title starting with `Composer update: `.


## Building your own docker images

The `docker/Dockerfile-*` recipes builds both `gitlabci-composer-update-mr` and use official PHP docker images. Please refer to those Dockerfiles for details on build instructions.


## Notes

## Composer versions

Both composer 1 & 2 are installed in the docker container, and can be called via `composer-1` and `composer-2` respectively. The composer update process will call the correct one accordingly depending on the `COMPOSER_MR_COMPOSER_VERSION` environment variable.


### Private Packages

If you are using GitLab’s [composer package registry](https://docs.gitlab.com/ee/user/packages/composer_repository/) to host private packages, you need to configure composer to use an API token to retrieve them.

Example:

```yml
composer-update-mr:
  script:
    # Use composer-1 or composer2 based on your requirements.
    # Replace {your-gitlab-server-domain} with the domain name for your server.
    # You can use the same $COMPOSER_MR_TOKEN or supply a different API token with at least the `read_api` scope. $CI_JOB_TOKEN does **not** work.
    - composer-2 config http-basic.{your-gitlab-server-domain} ___token___ "$COMPOSER_MR_TOKEN"
    - gitlabci-composer-update-mr <commit-user> <commit-email> <source-branch>
```