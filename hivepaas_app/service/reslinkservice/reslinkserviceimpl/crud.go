package reslinkserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/reslinkservice"
)

func (s *service) SetLinks(
	ctx context.Context,
	db database.IDB,
	srcType base.ResourceType,
	srcID string,
	dstType base.ResourceType,
	dstIDs []string,
	options ...reslinkservice.SetLinkOption,
) error {
	currLinks, _, err := s.resLinkRepo.List(ctx, db, nil,
		bunex.SelectWhere("res_link.src_type = ?", srcType),
		bunex.SelectWhere("res_link.src_id = ?", srcID),
		bunex.SelectWhere("res_link.dst_type = ?", dstType),
	)
	if err != nil {
		return apperrors.New(err)
	}

	// Calculate links update
	mapCurrDstLinks := make(map[string]*entity.ResLink, len(currLinks))
	for _, link := range currLinks {
		mapCurrDstLinks[link.DstID] = link
	}

	dstIDs = gofn.ToSet(dstIDs)
	upsertingLinks := make([]*entity.ResLink, 0, len(dstIDs))
	timeNow := timeutil.NowUTC()

	for _, newDstID := range dstIDs {
		if _, ok := mapCurrDstLinks[newDstID]; ok {
			delete(mapCurrDstLinks, newDstID)
		} else { // No existing link in the current map, need to add
			upsertingLinks = append(upsertingLinks, &entity.ResLink{
				SrcType:   srcType,
				SrcID:     srcID,
				DstType:   dstType,
				DstID:     newDstID,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			})
		}
	}

	// Remaining links in the map need to delete
	for _, link := range mapCurrDstLinks {
		link.DeletedAt = timeNow
		upsertingLinks = append(upsertingLinks, link)
	}

	for _, option := range options {
		option(upsertingLinks)
	}

	err = s.resLinkRepo.UpsertMulti(ctx, db, upsertingLinks,
		entity.ResLinkUpsertingConflictCols, entity.ResLinkUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) AddLinks(
	ctx context.Context,
	db database.IDB,
	srcType base.ResourceType,
	srcID string,
	dstType base.ResourceType,
	dstIDs []string,
	options ...reslinkservice.SetLinkOption,
) error {
	if len(dstIDs) == 0 {
		return nil
	}

	// Make sure no duplicated ID in dstIDs
	dstIDs = gofn.ToSet(dstIDs)

	// Create and insert new links
	timeNow := timeutil.NowUTC()
	newLinks := make([]*entity.ResLink, 0, len(dstIDs))
	for _, dstID := range dstIDs {
		newLinks = append(newLinks, &entity.ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   dstType,
			DstID:     dstID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}

	for _, option := range options {
		option(newLinks)
	}

	err := s.resLinkRepo.UpsertMulti(ctx, db, newLinks,
		entity.ResLinkUpsertingConflictCols, entity.ResLinkUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) RemoveLinks(
	ctx context.Context,
	db database.IDB,
	srcType base.ResourceType,
	srcID string,
	dstType base.ResourceType,
	dstIDs []string,
) error {
	if len(dstIDs) == 0 {
		return nil
	}

	err := s.resLinkRepo.DeleteAllBySourceIDs(ctx, db, srcType, []string{srcID},
		bunex.DeleteWhere("res_link.dst_type = ?", dstType),
		bunex.DeleteWhereIn("res_link.dst_id IN (?)", dstIDs...),
	)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
