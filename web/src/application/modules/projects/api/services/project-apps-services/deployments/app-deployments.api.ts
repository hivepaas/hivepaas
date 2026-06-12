import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppDeploymentsApiValidator,
    AppDeployments_Cancel_Req,
    AppDeployments_Cancel_Res,
    AppDeployments_FindManyPaginated_Req,
    AppDeployments_FindManyPaginated_Res,
    AppDeployments_FindOneById_Req,
    AppDeployments_FindOneById_Res,
} from "~/projects/api/services/project-apps-services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AppDeploymentsApi extends BaseApi {
    public constructor(private readonly validator: AppDeploymentsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AppDeployments_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDeployments_FindManyPaginated_Res, Error>> {
        const { projectID, appID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/deployments`, {
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
        request: AppDeployments_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDeployments_FindOneById_Res, Error>> {
        const { projectID, appID, deploymentID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/deployments/${deploymentID}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async cancel(request: AppDeployments_Cancel_Req): Promise<Result<AppDeployments_Cancel_Res, Error>> {
        const { projectID, appID, deploymentID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps/${appID}/deployments/${deploymentID}/cancel`, {}),
            ).pipe(
                map(this.validator.cancel),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
