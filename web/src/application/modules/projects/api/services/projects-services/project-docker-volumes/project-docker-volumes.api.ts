import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type ProjectDockerVolumes_List_Req,
    type ProjectDockerVolumes_List_Res,
} from "./project-docker-volumes.api.contracts";
import { type ProjectDockerVolumesApiValidator } from "./project-docker-volumes.api.validator";

export class ProjectDockerVolumesApi extends BaseApi {
    public constructor(private readonly validator: ProjectDockerVolumesApiValidator) {
        super();
    }

    async list(
        request: ProjectDockerVolumes_List_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectDockerVolumes_List_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/docker-volumes`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.list),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    get queries(): {
        list: (request: ProjectDockerVolumes_List_Req, signal?: AbortSignal) => Promise<Result<ProjectDockerVolumes_List_Res, Error>>;
    } {
        return {
            list: this.list.bind(this),
        };
    }
}
