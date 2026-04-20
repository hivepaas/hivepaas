import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { ProjectSslCertApiValidator } from "./project-ssl-cert.api.validator";
import type {
    ProjectSslCert_CreateOne_Req,
    ProjectSslCert_CreateOne_Res,
    ProjectSslCert_DeleteOne_Req,
    ProjectSslCert_DeleteOne_Res,
    ProjectSslCert_FindManyPaginated_Req,
    ProjectSslCert_FindManyPaginated_Res,
    ProjectSslCert_FindOneById_Req,
    ProjectSslCert_FindOneById_Res,
    ProjectSslCert_UpdateOne_Req,
    ProjectSslCert_UpdateOne_Res,
    ProjectSslCert_UpdateStatus_Req,
    ProjectSslCert_UpdateStatus_Res,
} from "./project-ssl-cert.api.contracts";

export class ProjectSslCertApi extends BaseApi {
    public constructor(private readonly validator: ProjectSslCertApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectSslCert_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting, domain } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = {
            ...query.build(),
            ...(domain !== undefined && domain !== "" ? { domain } : {}),
        };

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/ssl-certs`, {
                    params,
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
        request: ProjectSslCert_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/ssl-certs/${id}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: ProjectSslCert_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/ssl-certs`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectSslCert_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/ssl-certs/${id}`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectSslCert_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/ssl-certs/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectSslCert_DeleteOne_Req,
    ): Promise<Result<ProjectSslCert_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/ssl-certs/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
