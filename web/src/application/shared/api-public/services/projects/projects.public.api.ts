import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type ProjectsPublicApiValidator,
    type Public_Projects_FindManyPaginated_Req,
    type Public_Projects_FindManyPaginated_Res,
} from "@application/shared/api-public/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class ProjectsPublicApi extends BaseApi {
    public constructor(private readonly validator: ProjectsPublicApiValidator) {
        super();
    }

    /**
     * Find many public projects paginated
     */
    async findManyPaginated(
        request: Public_Projects_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Public_Projects_FindManyPaginated_Res, Error>> {
        const { search } = request.data;

        const query = this.queryBuilder.getInstance();

        query.sorting([{ id: "name", desc: false }]).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/projects/base", {
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
