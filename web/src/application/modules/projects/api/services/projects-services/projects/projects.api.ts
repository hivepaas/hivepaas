import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectsApiValidator,
    Projects_FindManyPaginated_Req,
    Projects_FindManyPaginated_Res,
} from "~/projects/api/services/projects-services/projects";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class ProjectsApi extends BaseApi {
    public constructor(private readonly validator: ProjectsApiValidator) {
        super();
    }

    /**
     * Find many projects paginated
     */
    async findManyPaginated(
        request: Projects_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Projects_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/projects", {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
