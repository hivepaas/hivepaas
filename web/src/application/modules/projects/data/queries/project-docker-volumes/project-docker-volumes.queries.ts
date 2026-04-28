import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectDockerVolumesApi } from "~/projects/api";
import {
    type ProjectDockerVolumes_List_Req,
    type ProjectDockerVolumes_List_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type ListReq = ProjectDockerVolumes_List_Req["data"];
type ListRes = ProjectDockerVolumes_List_Res;

type ListOptions = Omit<UseQueryOptions<ListRes>, "queryKey" | "queryFn">;

function useList(request: ListReq, options: ListOptions = {}) {
    const { queries } = useProjectDockerVolumesApi();

    return useQuery({
        queryKey: [QK["projects.docker-volumes.$.list"], request],
        queryFn: ({ signal }) => queries.list(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const ProjectDockerVolumesQueries = Object.freeze({
    useList,
});
