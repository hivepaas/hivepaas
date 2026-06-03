import type { PaginationState, SortingState } from "@infrastructure/data";
import type { ClusterVolume } from "~/cluster/domain";

import type { ApiRequestBase, ApiResponsePaginated } from "@infrastructure/api";

export type ClusterVolumes_List_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    type?: "cluster" | "volume";
}>;

export type ClusterVolumes_List_Res = ApiResponsePaginated<ClusterVolume>;
