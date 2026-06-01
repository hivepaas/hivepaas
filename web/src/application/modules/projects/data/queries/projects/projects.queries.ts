import { useEffect } from "react";

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectsApi } from "~/projects/api";
import {
    type Projects_FindManyPaginated_Req,
    type Projects_FindManyPaginated_Res,
    type Projects_FindOneById_Req,
    type Projects_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

import { useProfileContext } from "@application/shared/context";
import { useProjectPermissionsStore } from "@application/shared/permissions/store";
import { FULL_ACTIONS } from "@application/shared/permissions/utils";

/**
 * Find many projects paginated query
 */
type FindManyPaginatedReq = Projects_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Projects_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectsApi();

    return useQuery({
        queryKey: [QK["projects.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

/**
 * Find one project by id query
 */
type FindOneByIdReq = Projects_FindOneById_Req["data"];
type FindOneByIdRes = Projects_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectsApi();
    const profile = useProfileContext(state => state.profile);
    const upsertProject = useProjectPermissionsStore(state => state.upsertProject);
    const removeProject = useProjectPermissionsStore(state => state.removeProject);

    const query = useQuery({
        queryKey: [QK["projects.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });

    useEffect(() => {
        const project = query.data?.data;

        if (!project || !profile) {
            return;
        }

        if (project.owner.id === profile.id) {
            upsertProject({
                projectId: project.id,
                actions: FULL_ACTIONS,
            });
            return;
        }

        const currentUserAccess = project.userAccesses.find(user => user.id === profile.id);

        if (!currentUserAccess) {
            removeProject(project.id);
            return;
        }

        upsertProject({
            projectId: project.id,
            actions: currentUserAccess.access,
        });
    }, [profile, query.data?.data, removeProject, upsertProject]);

    return query;
}

export const ProjectsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
