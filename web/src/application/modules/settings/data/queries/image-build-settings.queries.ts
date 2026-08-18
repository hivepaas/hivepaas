import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useImageBuildSettingsApi } from "~/settings/api/hooks";
import type { ImageBuildSettings_FindOne_Res, ImageBuildSettings_FindRepoCache_Res } from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindOneRes = ImageBuildSettings_FindOne_Res;
type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

function useFindOne(options: FindOneOptions = {}) {
    const { queries } = useImageBuildSettingsApi();

    return useQuery({
        queryKey: [QK["settings.image-build-settings.find-one"]],
        queryFn: ({ signal }) => queries.findOne(signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindRepoCacheRes = ImageBuildSettings_FindRepoCache_Res;
type FindRepoCacheOptions = Omit<UseQueryOptions<FindRepoCacheRes>, "queryKey" | "queryFn">;

function useFindRepoCache(options: FindRepoCacheOptions = {}) {
    const { queries } = useImageBuildSettingsApi();

    return useQuery({
        queryKey: [QK["settings.image-build-settings.repo-cache.find-one"]],
        queryFn: ({ signal }) => queries.findRepoCache(signal),
        ...options,
    });
}

export const ImageBuildSettingsQueries = Object.freeze({
    useFindOne,
    useFindRepoCache,
});
