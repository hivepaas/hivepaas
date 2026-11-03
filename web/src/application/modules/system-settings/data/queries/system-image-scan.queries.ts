import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useSystemImageScanApi } from "~/system-settings/api/hooks";
import type {
    SystemImageScan_FindApps_Req,
    SystemImageScan_FindApps_Res,
    SystemImageScan_FindComponent_Req,
    SystemImageScan_FindComponent_Res,
    SystemImageScan_FindComponents_Res,
    SystemImageScan_FindOne_Req,
    SystemImageScan_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = SystemImageScan_FindOne_Req["data"];
type FindOneRes = SystemImageScan_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useSystemImageScanApi();

    return useQuery({
        queryKey: [QK["system-settings.image-scan.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

type FindAppsReq = SystemImageScan_FindApps_Req["data"];
type FindAppsRes = SystemImageScan_FindApps_Res;

function useFindApps(
    request: FindAppsReq = {},
    options: Omit<UseQueryOptions<FindAppsRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSystemImageScanApi();

    return useQuery({
        queryKey: [QK["system-settings.image-scan.find-apps"], request],
        queryFn: ({ signal }) => queries.findApps(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

function useFindComponents(
    options: Omit<UseQueryOptions<SystemImageScan_FindComponents_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSystemImageScanApi();

    return useQuery({
        queryKey: [QK["system-settings.image-scan.find-components"]],
        queryFn: ({ signal }) => queries.findComponents(signal),
        ...options,
    });
}

function useFindComponent(
    request: SystemImageScan_FindComponent_Req["data"],
    options: Omit<UseQueryOptions<SystemImageScan_FindComponent_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSystemImageScanApi();

    return useQuery({
        queryKey: [QK["system-settings.image-scan.find-component"], request],
        queryFn: ({ signal }) => queries.findComponent(request, signal),
        ...options,
    });
}

export const SystemImageScanQueries = Object.freeze({
    useFindOne,
    useFindApps,
    useFindComponents,
    useFindComponent,
});
