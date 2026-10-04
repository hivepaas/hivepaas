// Package composeservice reads a Docker Compose file into what spec import
// writes: a project, one env, an app per service. It touches nothing on the
// server - neither its files nor its environment - and decides nothing that
// needs the database: the import plans and checks what it reads.
// See docs/superpowers/specs/2026-10-04-project-from-compose-design.md.
package composeservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

type Service interface {
	// Convert reads a compose file and the choices made of it into a bundle.
	Convert(ctx context.Context, req *ConvertReq) (*ConvertResp, error)
}

// PortAs is what a published port becomes.
type PortAs string

const (
	// PortAsDomain is HTTP through Traefik, on a domain.
	PortAsDomain PortAs = "domain"
	// PortAsNode is the port published on the nodes, as compose publishes it.
	PortAsNode PortAs = "node"
	// PortAsNone is not published: the env's network reaches it anyway.
	PortAsNone PortAs = "none"
)

var AllPortAs = []PortAs{PortAsDomain, PortAsNode, PortAsNone}

// ConvertReq is a compose file, what it reads, and the choices made of it.
type ConvertReq struct {
	// Compose is the compose file; DotEnv the .env beside it.
	Compose string
	DotEnv  string
	// Files are the files the compose file reads, by their path in it.
	Files map[string][]byte
	// Variables are what was typed in the review, by name: over the .env's.
	Variables map[string]*VariableReq
	// The project and its env, checked by the caller; NetworkName the env's
	// network, which every app joins.
	ProjectKey, ProjectName string
	EnvKey, EnvName         string
	NetworkName             string
	// Profiles are the profiles whose services are created, beside those with
	// none.
	Profiles []string
	// Services are the choices made of each service, by its name.
	Services map[string]*ServiceReq
	// Volume is the cluster volume the services' data goes to; nil for the
	// project's default.
	Volume *specmodel.ExternalRef
	// RootDomain is what a domain is suggested under; empty for none.
	RootDomain string
	// MayWriteCluster and MayBindHost are what the caller may grant an app:
	// capabilities, and a host's directory. What it may not is left out of the
	// app, which is created without it.
	MayWriteCluster bool
	MayBindHost     bool
}

// VariableReq is what the review says of a variable.
type VariableReq struct {
	// Value is the variable's, over the .env's; nil keeps that one.
	Value *string
	// Secret keeps the value as an env secret; nil decides by the name.
	Secret *bool
}

// ServiceReq is what the review says of a service.
type ServiceReq struct {
	// Image is the image of a service with only a build.
	Image string
	// Ports are the choices for its published ports; one not listed takes its
	// default.
	Ports []*PortReq
}

// PortReq is the choice for one published port, found by Published, Target
// and Protocol.
type PortReq struct {
	Published uint32
	Target    uint32
	Protocol  string
	As        PortAs
	Domain    string
}

// ConvertResp is the bundle, the issues of reading it, and what the review
// shows.
type ConvertResp struct {
	// Bundle is nil while a required variable has no value: nothing can be read
	// before it has one.
	Bundle *specmodel.ImportBundle
	// Issues are the plan's, by node path.
	Issues map[string][]specmodel.Issue
	// FileName is the project's name in the file, `name:`; empty for none.
	FileName  string
	Services  []*ServiceView
	Variables []*VariableView
	Needs     []*FileNeed
	// Profiles are every profile the file names.
	Profiles []string
}

// ServiceView is a service as the app it becomes.
type ServiceView struct {
	Name string `json:"name"`
	// App is the key of the app it becomes.
	App   string `json:"app"`
	Image string `json:"image"`
	// Build says the file builds the service's image.
	Build bool `json:"build"`
	// Skipped says the service is not created, and Reason why.
	Skipped bool   `json:"skipped"`
	Reason  string `json:"reason"`
	Mode    string `json:"mode"`
	// Replicas of a replicated service.
	Replicas uint64        `json:"replicas"`
	Ports    []*PortView   `json:"ports"`
	Volumes  []*VolumeView `json:"volumes"`
	// Aliases are the names the service is reached by beside its key.
	Aliases []string `json:"aliases"`
	// Dropped are the compose fields not carried to the app.
	Dropped []string `json:"dropped"`
}

// PortView is one published port, and what it becomes.
type PortView struct {
	Published uint32 `json:"published"`
	Target    uint32 `json:"target"`
	Protocol  string `json:"protocol"`
	As        PortAs `json:"as"`
	Default   PortAs `json:"default"`
	Domain    string `json:"domain"`
	// Suggested is the domain the review offers.
	Suggested string `json:"suggested"`
}

// VolumeKind is what a service's mount becomes.
type VolumeKind string

const (
	// VolumeKindVolume is the app's directory on the project's volume.
	VolumeKindVolume VolumeKind = "volume"
	// VolumeKindShared is another app's, a volume both mount.
	VolumeKindShared VolumeKind = "shared"
	// VolumeKindFile is a file of the env's, mounted from a setting.
	VolumeKindFile VolumeKind = "file"
	// VolumeKindHost is a directory of the host's.
	VolumeKindHost VolumeKind = "host"
	// VolumeKindTmpfs is memory.
	VolumeKindTmpfs VolumeKind = "tmpfs"
	// VolumeKindDropped is not mounted.
	VolumeKindDropped VolumeKind = "dropped"
)

// VolumeView is one mount of a service.
type VolumeView struct {
	Target string `json:"target"`
	// Source is what the compose file mounts: a volume's name, a path.
	Source   string     `json:"source"`
	Kind     VolumeKind `json:"kind"`
	ReadOnly bool       `json:"readOnly"`
	// Owner is the app whose directory a shared volume is.
	Owner string `json:"owner"`
}

// VariableView is one variable the file uses.
type VariableView struct {
	Name     string `json:"name"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
	// Given says it has a value, from the .env or the review.
	Given  bool `json:"given"`
	Secret bool `json:"secret"`
}

// FileNeed is a file the compose file reads.
type FileNeed struct {
	Path string `json:"path"`
	// As is what reads it: env_file, config, secret, bind.
	As string `json:"as"`
	// By are the services that read it.
	By []string `json:"by"`
	// Given says the request carries it.
	Given bool `json:"given"`
}
