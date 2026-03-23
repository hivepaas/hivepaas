package appdto

import (
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/docker"
)

type GetAppContainerSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppContainerSettingsReq() *GetAppContainerSettingsReq {
	return &GetAppContainerSettingsReq{}
}

func (req *GetAppContainerSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppContainerSettingsResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *ContainerSettingsResp `json:"data"`
}

type ContainerSettingsResp struct {
	*ContainerSpec

	UpdateVer int `json:"updateVer"`
}

type ContainerSpec struct {
	Labels          map[string]string  `json:"labels"`
	Image           string             `json:"image"`
	Command         string             `json:"command"`
	WorkingDir      string             `json:"workingDir"`
	Hostname        string             `json:"hostname"`
	User            string             `json:"user"`
	Groups          []string           `json:"groups"`
	StopSignal      string             `json:"stopSignal"`
	TTY             bool               `json:"tty"`
	OpenStdin       bool               `json:"openStdin"`
	ReadOnly        bool               `json:"readOnly"`
	StopGracePeriod *timeutil.Duration `json:"stopGracePeriod"`
	Privileges      *Privileges        `json:"privileges"`
	Healthcheck     *Healthcheck       `json:"healthcheck"`
	RestartPolicy   *RestartPolicy     `json:"restartPolicy"`

	UpdateVer int `json:"updateVer"`
}

type RestartPolicy struct {
	Condition   swarm.RestartPolicyCondition `json:"condition,omitempty"`
	Delay       *timeutil.Duration           `json:"delay,omitempty"`
	MaxAttempts *uint64                      `json:"maxAttempts,omitempty"`
	Window      *timeutil.Duration           `json:"window,omitempty"`
}

type Healthcheck struct {
	Enabled bool `json:"enabled,omitempty"`

	// Test is the test to perform to check that the container is healthy.
	// An empty slice means to inherit the default.
	// The options are:
	// {} : inherit healthcheck
	// {"NONE"} : disable healthcheck
	// {"CMD", args...} : exec arguments directly
	// {"CMD-SHELL", command} : run command with system's default shell
	Mode    docker.HealthcheckMode `json:"mode,omitempty"`
	Command string                 `json:"command,omitempty"`

	// Zero means to inherit. Durations are expressed as integer nanoseconds.
	Interval      timeutil.Duration `json:"interval,omitempty"`
	Timeout       timeutil.Duration `json:"timeout,omitempty"`
	StartPeriod   timeutil.Duration `json:"startPeriod,omitempty"`
	StartInterval timeutil.Duration `json:"startInterval,omitempty"`

	// Retries is the number of consecutive failures needed to consider a container as unhealthy.
	// Zero means inherit.
	Retries int `json:"retries,omitempty"`
}

type SELinuxContext struct {
	Disable bool   `json:"disable,omitempty"`
	User    string `json:"user,omitempty"`
	Role    string `json:"role,omitempty"`
	Type    string `json:"type,omitempty"`
	Level   string `json:"level,omitempty"`
}

type SeccompOpts struct {
	Mode    swarm.SeccompMode `json:"mode,omitempty"`
	Profile string            `json:"profile,omitempty"`
}

type AppArmorOpts struct {
	Mode swarm.AppArmorMode `json:"mode,omitempty"`
}

type Privileges struct {
	SELinuxContext  *SELinuxContext `json:"seLinuxContext,omitempty"`
	Seccomp         *SeccompOpts    `json:"seccomp,omitempty"`
	AppArmor        *AppArmorOpts   `json:"appArmor,omitempty"`
	NoNewPrivileges bool            `json:"noNewPrivileges,omitempty"`
}

func TransformContainerSettings(
	service *swarm.Service,
) (resp *ContainerSettingsResp, err error) {
	spec := &service.Spec
	resp = &ContainerSettingsResp{
		UpdateVer: int(service.Version.Index), //nolint:gosec
	}

	resp.ContainerSpec = TransformContainerSpec(&spec.TaskTemplate)

	return resp, nil
}

func TransformContainerSpec(taskSpec *swarm.TaskSpec) *ContainerSpec {
	containerSpec := taskSpec.ContainerSpec
	if containerSpec == nil {
		return nil
	}
	res := &ContainerSpec{
		Labels:        containerSpec.Labels,
		Image:         containerSpec.Image,
		Command:       strings.Join(gofn.Concat(containerSpec.Command, containerSpec.Args), " "),
		WorkingDir:    containerSpec.Dir,
		Hostname:      containerSpec.Hostname,
		User:          containerSpec.User,
		Groups:        containerSpec.Groups,
		StopSignal:    containerSpec.StopSignal,
		TTY:           containerSpec.TTY,
		OpenStdin:     containerSpec.OpenStdin,
		ReadOnly:      containerSpec.ReadOnly,
		Privileges:    TransformContainerPrivileges(containerSpec.Privileges),
		Healthcheck:   TransformContainerHealthcheck(containerSpec.Healthcheck),
		RestartPolicy: TransformContainerRestartPolicy(taskSpec.RestartPolicy),
	}
	if containerSpec.StopGracePeriod != nil {
		res.StopGracePeriod = new(timeutil.Duration(*containerSpec.StopGracePeriod))
	}
	return res
}

func TransformContainerPrivileges(privileges *swarm.Privileges) (res *Privileges) {
	if privileges == nil {
		return nil
	}
	res = &Privileges{
		NoNewPrivileges: privileges.NoNewPrivileges,
	}
	if privileges.SELinuxContext != nil {
		res.SELinuxContext = &SELinuxContext{
			Disable: privileges.SELinuxContext.Disable,
			User:    privileges.SELinuxContext.User,
			Role:    privileges.SELinuxContext.Role,
			Type:    privileges.SELinuxContext.Type,
			Level:   privileges.SELinuxContext.Level,
		}
	}
	if privileges.Seccomp != nil {
		res.Seccomp = &SeccompOpts{
			Mode:    privileges.Seccomp.Mode,
			Profile: reflectutil.UnsafeBytesToStr(privileges.Seccomp.Profile),
		}
	}
	if privileges.AppArmor != nil {
		res.AppArmor = &AppArmorOpts{
			Mode: privileges.AppArmor.Mode,
		}
	}
	return res
}

func TransformContainerHealthcheck(config *container.HealthConfig) *Healthcheck {
	if config == nil {
		return nil
	}
	cmd := config.Test
	var mode docker.HealthcheckMode
	if len(cmd) > 0 {
		mode = docker.HealthcheckMode(cmd[0])
		cmd = cmd[1:]
	}
	res := &Healthcheck{
		Enabled:       mode != "NONE",
		Mode:          mode,
		Command:       strings.Join(cmd, " "),
		Interval:      timeutil.Duration(config.Interval),
		Timeout:       timeutil.Duration(config.Timeout),
		StartPeriod:   timeutil.Duration(config.StartPeriod),
		StartInterval: timeutil.Duration(config.StartInterval),
		Retries:       config.Retries,
	}
	return res
}

func TransformContainerRestartPolicy(policy *swarm.RestartPolicy) *RestartPolicy {
	if policy == nil {
		return nil
	}
	res := &RestartPolicy{
		Condition:   policy.Condition,
		MaxAttempts: policy.MaxAttempts,
	}
	if policy.Delay != nil {
		res.Delay = new(timeutil.Duration(*policy.Delay))
	}
	if policy.Window != nil {
		res.Window = new(timeutil.Duration(*policy.Window))
	}
	return res
}
