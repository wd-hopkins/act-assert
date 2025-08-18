package act_assert

import (
	"github.com/docker/docker/api/types/container"
	"github.com/wd-hopkins/act/pkg/runner"
)

type config struct {
	actor                              string                     // the user that triggered the event
	workdir                            string                     // path to working directory
	actionCacheDir                     string                     // path used for caching action contents
	actionOfflineMode                  bool                       // when offline, use caching action contents
	bindWorkdir                        bool                       // bind the workdir to the job container
	eventName                          string                     // name of event to run
	eventPath                          string                     // path to JSON file to use for event.json in containers
	defaultBranch                      string                     // name of the main branch for this repository
	reuseContainers                    bool                       // reuse containers to maintain state
	forcePull                          bool                       // force pulling of the image, even if already present
	forceRebuild                       bool                       // force rebuilding local docker image action
	logOutput                          bool                       // log the output from docker run
	jSONLogger                         bool                       // use json or text logger
	logPrefixJobID                     bool                       // switches from the full job name to the job id
	env                                map[string]string          // env for containers
	inputs                             map[string]string          // manually passed action inputs
	secrets                            map[string]string          // list of secrets
	vars                               map[string]string          // list of vars
	token                              string                     // GitHub token
	insecureSecrets                    bool                       // switch hiding output when printing to terminal
	platforms                          map[string]string          // list of platforms
	privileged                         bool                       // use privileged mode
	usernsMode                         string                     // user namespace to use
	containerArchitecture              string                     // Desired OS/architecture platform for running containers
	containerDaemonSocket              string                     // Path to Docker daemon socket
	containerOptions                   string                     // Options for the job container
	useGitIgnore                       bool                       // controls if paths in .gitignore should not be copied into container, default true
	gitHubInstance                     string                     // GitHub instance to use, default "github.com"
	runAsUser                          string                     // User UID with which to run the job container
	containerCapAdd                    []string                   // list of kernel capabilities to add to the containers
	containerCapDrop                   []string                   // list of kernel capabilities to remove from the containers
	autoRemove                         bool                       // controls if the container is automatically removed upon workflow completion
	artifactServerPath                 string                     // the path where the artifact server stores uploads
	artifactServerAddr                 string                     // the address the artifact server binds to
	artifactServerPort                 string                     // the port the artifact server binds to
	noSkipCheckout                     bool                       // do not skip actions/checkout
	remoteName                         string                     // remote name in local git repo config
	replaceGheActionWithGithubCom      []string                   // Use actions from GitHub Enterprise instance to GitHub
	replaceGheActionTokenWithGithubCom string                     // Token of private action repo on GitHub.
	matrix                             map[string]map[string]bool // Matrix config to run
	containerNetworkMode               container.NetworkMode      // the network mode of job containers (the value of --network)
	actionCache                        runner.ActionCache         // Use a custom ActionCache Implementation
}

func (c config) toRunnerConfig() *runner.Config {
	return &runner.Config{
		Actor:                              c.actor,
		Workdir:                            c.workdir,
		ActionCacheDir:                     c.actionCacheDir,
		ActionOfflineMode:                  c.actionOfflineMode,
		BindWorkdir:                        c.bindWorkdir,
		EventName:                          c.eventName,
		EventPath:                          c.eventPath,
		DefaultBranch:                      c.defaultBranch,
		ReuseContainers:                    c.reuseContainers,
		ForcePull:                          c.forcePull,
		ForceRebuild:                       c.forceRebuild,
		LogOutput:                          c.logOutput,
		JSONLogger:                         c.jSONLogger,
		LogPrefixJobID:                     c.logPrefixJobID,
		Env:                                c.env,
		Inputs:                             c.inputs,
		Secrets:                            c.secrets,
		Vars:                               c.vars,
		Token:                              c.token,
		InsecureSecrets:                    c.insecureSecrets,
		Platforms:                          c.platforms,
		Privileged:                         c.privileged,
		UsernsMode:                         c.usernsMode,
		ContainerArchitecture:              c.containerArchitecture,
		ContainerDaemonSocket:              c.containerDaemonSocket,
		ContainerOptions:                   c.containerOptions,
		UseGitIgnore:                       c.useGitIgnore,
		GitHubInstance:                     c.gitHubInstance,
		RunAsUser:                          c.runAsUser,
		ContainerCapAdd:                    c.containerCapAdd,
		ContainerCapDrop:                   c.containerCapDrop,
		AutoRemove:                         c.autoRemove,
		ArtifactServerPath:                 c.artifactServerPath,
		ArtifactServerAddr:                 c.artifactServerAddr,
		ArtifactServerPort:                 c.artifactServerPort,
		NoSkipCheckout:                     c.noSkipCheckout,
		RemoteName:                         c.remoteName,
		ReplaceGheActionWithGithubCom:      c.replaceGheActionWithGithubCom,
		ReplaceGheActionTokenWithGithubCom: c.replaceGheActionTokenWithGithubCom,
		Matrix:                             c.matrix,
		ContainerNetworkMode:               c.containerNetworkMode,
		ActionCache:                        c.actionCache,
	}
}
