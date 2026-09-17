package apptemplateserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/services/registry"
)

const (
	// maxScannedTags bounds what is read from a registry; maxOfferedTags bounds
	// what a person is asked to choose from.
	//
	// The scan has to reach the end of the repository, not just the start: a
	// registry answers in lexical order, so 18.6 sorts after every 17.x, and a cap
	// that cuts the list short hides exactly the releases somebody opened this to
	// find. library/postgres publishes 1421 tags today, and 5000 leaves room for a
	// repository several times that before the answer is marked truncated.
	maxScannedTags = 5000
	maxOfferedTags = 50
)

func (s *service) ImageTags(
	ctx context.Context,
	req *apptemplateservice.ImageTagsReq,
) (*apptemplateservice.ImageTagsResp, error) {
	line, image, err := s.templateImage(ctx, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	ref := registry.ParseRepository(image)
	tags, found := s.tagCache.get(ref.String())
	if !found {
		if tags, err = s.registryClient.ListTags(ctx, ref, maxScannedTags); err != nil {
			return nil, hperrors.Wrap(err)
		}
		s.tagCache.put(ref.String(), tags)
	}

	candidates := templatemodel.SelectTagCandidates(line, image, tags.Tags, maxOfferedTags)
	resp := &apptemplateservice.ImageTagsResp{
		Repository: ref.String(),
		CurrentTag: imageref.Parse(image).Tag,
		Truncated:  tags.Truncated,
		Tags:       make([]*apptemplateservice.ImageTag, 0, len(candidates)),
	}
	for _, candidate := range candidates {
		resp.Tags = append(resp.Tags, &apptemplateservice.ImageTag{
			Tag: candidate.Tag, Class: candidate.Class, Newer: candidate.Newer,
		})
	}
	return resp, nil
}

// templateImage is the image the chosen version and variant pin, and the version's
// name - the line its tags are classified against. A deprecated version is allowed
// here: somebody running one is exactly who needs to see what else the repository
// publishes.
func (s *service) templateImage(
	ctx context.Context,
	req *apptemplateservice.ImageTagsReq,
) (line, image string, err error) {
	loaded, err := s.Template(ctx, req.Name)
	if err != nil {
		return "", "", hperrors.Wrap(err)
	}
	tmpl := loaded.Template

	version := tmpl.DefaultVersion()
	if req.Version != "" {
		version = tmpl.FindVersion(req.Version)
	}
	if version == nil {
		return "", "", hperrors.Wrap(hperrors.ErrAppTemplateVersionNotFound).
			WithParam("Template", req.Name).WithParam("Version", req.Version)
	}

	variantName := req.Variant
	if variantName == "" && len(tmpl.Variants) > 0 {
		if variant := tmpl.DefaultVariant(); variant != nil {
			variantName = variant.Name
		}
	}
	image = version.ImageFor(variantName)
	if image == "" {
		return "", "", hperrors.Wrap(hperrors.ErrAppTemplateVariantUnavailable).
			WithParam("Template", req.Name).WithParam("Version", version.Name).WithParam("Variant", variantName)
	}
	return version.Name, image, nil
}
