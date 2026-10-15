import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppImageScanApi } from "../../../api/hooks/project-apps";
import type {
    AppImageScan_FindOne_Req,
    AppImageScan_FindOne_Res,
    ProjectImageScan_Find_Req,
    ProjectImageScan_Find_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: AppImageScan_FindOne_Req["data"],
    options: Omit<UseQueryOptions<AppImageScan_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppImageScanApi();

    return useQuery({
        queryKey: [QK["projects.apps.image-scan.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

function useFindProject(
    request: ProjectImageScan_Find_Req["data"],
    options: Omit<UseQueryOptions<ProjectImageScan_Find_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppImageScanApi();

    return useQuery({
        queryKey: [QK["projects.image-scan.find-project"], request],
        queryFn: ({ signal }) => queries.findProject(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppImageScanQueries = Object.freeze({
    useFindOne,
    useFindProject,
});
