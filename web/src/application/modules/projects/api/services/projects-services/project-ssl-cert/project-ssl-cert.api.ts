import type { AxiosResponse } from "axios";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectSslCert_CreateOne_Req,
    ProjectSslCert_CreateOne_Res,
    ProjectSslCert_DeleteOne_Req,
    ProjectSslCert_DeleteOne_Res,
    ProjectSslCert_DownloadBundle_Req,
    ProjectSslCert_DownloadBundle_Res,
    ProjectSslCert_FindManyPaginated_Req,
    ProjectSslCert_FindManyPaginated_Res,
    ProjectSslCert_FindOneById_Req,
    ProjectSslCert_FindOneById_Res,
    ProjectSslCert_RenewOne_Req,
    ProjectSslCert_RenewOne_Res,
    ProjectSslCert_UpdateOne_Req,
    ProjectSslCert_UpdateOne_Res,
    ProjectSslCert_UpdateStatus_Req,
    ProjectSslCert_UpdateStatus_Res,
} from "./project-ssl-cert.api.contracts";
import type { ProjectSslCertApiValidator } from "./project-ssl-cert.api.validator";

function parseFilenameFromContentDisposition(contentDisposition?: string): string | undefined {
    if (!contentDisposition) {
        return undefined;
    }

    const encodedFilename = /filename\*=UTF-8''([^;]+)/i.exec(contentDisposition)?.[1];
    if (encodedFilename) {
        return decodeURIComponent(encodedFilename);
    }

    const filename = /filename="?([^";]+)"?/i.exec(contentDisposition)?.[1];
    return filename ? decodeURIComponent(filename) : undefined;
}

function mapDownloadResponse(response: AxiosResponse<Blob>): ProjectSslCert_DownloadBundle_Res {
    const headers = response.headers as Record<string, unknown>;
    const contentDisposition = headers["content-disposition"];

    return {
        data: {
            blob: response.data,
            filename: parseFilenameFromContentDisposition(
                typeof contentDisposition === "string" ? contentDisposition : undefined,
            ),
        },
    };
}

function getProjectSslCertBasePath(projectID: string, env?: string): string {
    if (env && env !== "all") {
        return `/projects/${projectID}/${encodeURIComponent(env)}/ssl-certs`;
    }

    return `/projects/${projectID}/ssl-certs`;
}

export class ProjectSslCertApi extends BaseApi {
    public constructor(private readonly validator: ProjectSslCertApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectSslCert_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting, domain } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = {
            ...query.build(),
            ...(domain !== undefined && domain !== "" ? { domain } : {}),
        };

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectSslCertBasePath(projectID, env), {
                    params,
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findManyPaginated(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: ProjectSslCert_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_FindOneById_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectSslCertBasePath(projectID, env)}/${id}`, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findOneById(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: ProjectSslCert_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_CreateOne_Res, Error>> {
        const { projectID, env, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectSslCertBasePath(projectID, env), json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.createOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectSslCert_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_UpdateOne_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectSslCertBasePath(projectID, env)}/${id}`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectSslCert_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_UpdateStatus_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectSslCertBasePath(projectID, env)}/${id}/status`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateStatus(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectSslCert_DeleteOne_Req): Promise<Result<ProjectSslCert_DeleteOne_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectSslCertBasePath(projectID, env)}/${id}`)).pipe(
                map(response => this.validator.deleteOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async renewOne(request: ProjectSslCert_RenewOne_Req): Promise<Result<ProjectSslCert_RenewOne_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`${getProjectSslCertBasePath(projectID, env)}/${id}/renew`, {})).pipe(
                map(response => this.validator.renewOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async downloadBundle(
        request: ProjectSslCert_DownloadBundle_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslCert_DownloadBundle_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get<Blob>(`${getProjectSslCertBasePath(projectID, env)}/${id}/download`, {
                    responseType: "blob",
                    signal,
                }),
            ).pipe(
                map(mapDownloadResponse),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
