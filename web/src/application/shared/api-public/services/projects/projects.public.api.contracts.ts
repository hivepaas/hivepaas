import type { ProjectPublic } from "@application/shared/entities";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many public projects paginated
 */
export type Public_Projects_FindManyPaginated_Req = ApiRequestBase<{
    search?: string;
}>;

export type Public_Projects_FindManyPaginated_Res = ApiResponsePaginated<ProjectPublic>;
