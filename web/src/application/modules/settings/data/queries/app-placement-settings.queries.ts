import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppPlacementSettingsApi } from "~/settings/api/hooks";
import type { AppPlacementSettings_FindOne_Res } from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindOneRes = AppPlacementSettings_FindOne_Res;
type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

function useFindOne(options: FindOneOptions = {}) {
    const { queries } = useAppPlacementSettingsApi();

    return useQuery({
        queryKey: [QK["settings.app-placement-settings.find-one"]],
        queryFn: ({ signal }) => queries.findOne(signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppPlacementSettingsQueries = Object.freeze({
    useFindOne,
});
