package imagebuildserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

// imageReferences is every name a build tags its image with: a local image's
// own name alone, or the app's names.
func imageReferences(
	req *imagebuildservice.ImageBuildReq,
	inputs *imagebuildservice.BuildInputs,
) ([]string, error) {
	if req.LocalImage != "" {
		return []string{req.LocalImage}, nil
	}
	return buildImageReferences(req.App, req.CommitHash, req.ImageTags, pushRegistry(inputs))
}

// buildImageReferences is every name the built image is tagged with, in the order
// they are applied.
//
// The first is what the service spec runs, so it is the commit's and it is the
// one carrying the registry. The local reference is last and is never pushed:
// push skips anything without a "/", which is how a build on a node keeps a name
// it can run from without sending it anywhere.
func buildImageReferences(
	app *entity.App,
	commitHash string,
	extraTags []string,
	regAuth *entity.RegistryAuth,
) ([]string, error) {
	repoName, err := app.ImageRepoName()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	commitTag, err := app.ImageTag(commitHash)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	prefix, err := app.ImageTagPrefix()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	tags := make([]string, 0, len(extraTags)+1)
	tags = append(tags, commitTag)
	for _, extra := range extraTags {
		tags = append(tags, prefix+"-"+extra)
	}

	refs := make([]string, 0, len(tags)*2) //nolint:mnd // the registry's and the local one
	if regAuth != nil {
		for _, tag := range tags {
			refs = append(refs, entity.ImageReference(regAuth.Address, regAuth.Username, repoName, tag))
		}
	}
	for _, tag := range tags {
		refs = append(refs, repoName+":"+tag)
	}
	return refs, nil
}
