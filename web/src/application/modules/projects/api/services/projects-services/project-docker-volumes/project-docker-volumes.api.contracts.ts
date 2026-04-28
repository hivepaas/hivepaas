import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type DockerVolume } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * List project docker volumes
 */
export type ProjectDockerVolumes_List_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectDockerVolumes_List_Res = ApiResponsePaginated<DockerVolume>;
