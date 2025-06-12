package act_assert

type GithubEnv string

const (
	GithubRunAttempt      GithubEnv = "GITHUB_RUN_ATTEMPT"
	GithubRunID           GithubEnv = "GITHUB_RUN_ID"
	GithubRunNumber       GithubEnv = "GITHUB_RUN_NUMBER"
	GithubRepositoryOwner GithubEnv = "GITHUB_REPOSITORY_OWNER"
	GithubRetentionDays   GithubEnv = "GITHUB_RETENTION_DAYS"
	RunnerPerflog         GithubEnv = "RUNNER_PERFLOG"
	RunnerTrackingId      GithubEnv = "RUNNER_TRACKING_ID"
	GithubRepository      GithubEnv = "GITHUB_REPOSITORY"
	GithubRef             GithubEnv = "GITHUB_REF"
	ShaRef                GithubEnv = "SHA_REF"
	GithubRefName         GithubEnv = "GITHUB_REF_NAME"
	GithubRefType         GithubEnv = "GITHUB_REF_TYPE"
	GithubBaseRef         GithubEnv = "GITHUB_BASE_REF"
	GithubHeadRef         GithubEnv = "GITHUB_HEAD_REF"
	GithubWorkspace       GithubEnv = "GITHUB_WORKSPACE"
	GithubServerUrl       GithubEnv = "GITHUB_SERVER_URL"
	GithubApiUrl          GithubEnv = "GITHUB_API_URL"
	GithubGraphqlUrl      GithubEnv = "GITHUB_GRAPHQL_URL"
)
