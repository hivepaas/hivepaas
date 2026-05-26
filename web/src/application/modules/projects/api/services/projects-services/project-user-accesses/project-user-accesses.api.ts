import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectUserAccessesApiValidator,
    ProjectUserAccesses_FindOne_Req,
    ProjectUserAccesses_FindOne_Res,
    ProjectUserAccesses_UpdateOne_Req,
    ProjectUserAccesses_UpdateOne_Res,
} from "~/projects/api/services/projects-services";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class ProjectUserAccessesApi extends BaseApi {
    public constructor(private readonly validator: ProjectUserAccessesApiValidator) {
        super();
    }

    /**
     * Find project user accesses
     */
    async findOne(
        request: ProjectUserAccesses_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectUserAccesses_FindOne_Res, Error>> {
        const { projectID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/user-accesses`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update project user accesses
     */
    async updateOne(
        request: ProjectUserAccesses_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectUserAccesses_UpdateOne_Res, Error>> {
        const { projectID, userAccesses } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `/projects/${projectID}/user-accesses`,
                    {
                        userAccesses: JsonTransformer.array({
                            data: userAccesses,
                        }),
                    },
                    {
                        signal,
                    },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
