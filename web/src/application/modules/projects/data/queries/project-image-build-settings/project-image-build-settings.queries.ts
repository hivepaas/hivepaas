import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectImageBuildSettingsApi } from "~/projects/api/hooks";
import type {
    ProjectImageBuildSettings_FindOne_Req,
    ProjectImageBuildSettings_FindOne_Res,
    ProjectImageBuildSettings_FindRepoCache_Req,
    ProjectImageBuildSettings_FindRepoCache_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindOneReq = ProjectImageBuildSettings_FindOne_Req["data"];
type FindOneRes = ProjectImageBuildSettings_FindOne_Res;
type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

function useFindOne(request: FindOneReq, options: FindOneOptions = {}) {
    const { queries } = useProjectImageBuildSettingsApi();

    return useQuery({
        queryKey: [QK["projects.image-build-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindRepoCacheReq = ProjectImageBuildSettings_FindRepoCache_Req["data"];
type FindRepoCacheRes = ProjectImageBuildSettings_FindRepoCache_Res;
type FindRepoCacheOptions = Omit<UseQueryOptions<FindRepoCacheRes>, "queryKey" | "queryFn">;

function useFindRepoCache(request: FindRepoCacheReq, options: FindRepoCacheOptions = {}) {
    const { queries } = useProjectImageBuildSettingsApi();

    return useQuery({
        queryKey: [QK["projects.image-build-settings.repo-cache.$.find-one"], request],
        queryFn: ({ signal }) => queries.findRepoCache(request, signal),
        ...options,
    });
}

export const ProjectImageBuildSettingsQueries = Object.freeze({
    useFindOne,
    useFindRepoCache,
});
