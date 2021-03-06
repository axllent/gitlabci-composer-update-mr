# Gitlab Composer Updater MR

A series of Alpine [docker images](https://hub.docker.com/r/axllent/gitlabci-composer-update-mr) containing [enomotodev/gitlabci-composer-update-mr](https://github.com/enomotodev/gitlabci-composer-update-mr) for automated scheduled composer update merge requests.

This is not a standalone container, and is designed for Gitlab CI integration only.


## Usage

### Setting GitLab personal access token to GitLabCI

GitLab personal access token is required for sending merge requests to your repository.

1. Go to your account's settings page and generate a personal access token with `api` scope `Account` -> `Preferences` -> `Access Tokens`
2. In your GitLab dashboard, go to your project's `Settings` -> `CI /CD` -> `Environment variables`
3. Add an environment variable `GITLAB_API_PRIVATE_TOKEN` with your GitLab personal access token


### Configure `.gitlab-ci.yml`

Configure your `.gitlab-ci.yml` to run gitlabci-composer-update-mr, for example:

```yaml
stages:
  - ...
  - ...
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
- `<php-version>` is your php version (currently 5.6, 7.3 or 7.4 supported)
- `<username>` is the git username for the merge request commit
- `<email>` is the git email for the merge request commit
- `<branch>` is the branch your wish to work on and create a merge request for

eg: `- gitlabci-composer-update-mr composer-update-mr me@example.com develop`


### Setting schedule

1. On GitLab dashboard, go to your application's `Schedules` -> `New schedule`
2. Create new schedule and save
