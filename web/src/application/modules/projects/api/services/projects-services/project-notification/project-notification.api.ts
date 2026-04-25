import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectNotification_FindManyPaginated_Req,
    ProjectNotification_FindManyPaginated_Res,
    ProjectNotification_FindOneById_Req,
    ProjectNotification_FindOneById_Res,
} from "./project-notification.api.contracts";
import type { ProjectNotificationApiValidator } from "./project-notification.api.validator";

export class ProjectNotificationApi extends BaseApi {
    public constructor(private readonly validator: ProjectNotificationApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectNotification_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNotification_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/notifications`, {
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

    async findOneById(
        request: ProjectNotification_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNotification_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/notifications/${id}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
