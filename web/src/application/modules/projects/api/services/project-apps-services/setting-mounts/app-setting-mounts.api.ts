import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppSettingMountPayload,
    AppSettingMountsApiValidator,
    AppSettingMounts_CreateOne_Req,
    AppSettingMounts_CreateOne_Res,
    AppSettingMounts_DeleteOne_Req,
    AppSettingMounts_DeleteOne_Res,
    AppSettingMounts_FindManyPaginated_Req,
    AppSettingMounts_FindManyPaginated_Res,
    AppSettingMounts_FindOneById_Req,
    AppSettingMounts_FindOneById_Res,
    AppSettingMounts_FindSources_Req,
    AppSettingMounts_FindSources_Res,
    AppSettingMounts_UpdateOne_Req,
    AppSettingMounts_UpdateOne_Res,
    AppSettingMounts_UpdateStatus_Req,
    AppSettingMounts_UpdateStatus_Res,
} from "~/projects/api/services/project-apps-services";

import { BaseApi, parseApiError } from "@infrastructure/api";

function basePath(projectID: string, env: string, appID: string): string {
    return `/projects/${projectID}/${env}/apps/${appID}/setting-mounts`;
}

function toBody({ name, inheritable, sourceID, files }: AppSettingMountPayload) {
    return {
        name,
        inheritable,
        source: { id: sourceID },
        files: files.map(file => ({
            part: file.part,
            path: file.path,
            uid: file.uid,
            gid: file.gid,
            mode: file.mode,
        })),
    };
}

export class AppSettingMountsApi extends BaseApi {
    public constructor(private readonly validator: AppSettingMountsApiValidator) {
        super();
    }

    /**
     * Find many app setting mounts paginated
     */
    async findManyPaginated(
        request: AppSettingMounts_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_FindManyPaginated_Res, Error>> {
        const { projectID, env, appID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get(basePath(projectID, env, appID), { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Find one app setting mount by id
     */
    async findOneById(
        request: AppSettingMounts_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_FindOneById_Res, Error>> {
        const { projectID, env, appID, settingMountID } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`${basePath(projectID, env, appID)}/${settingMountID}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Find the setting types an entry may mount from, and their parts
     */
    async findSources(
        request: AppSettingMounts_FindSources_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_FindSources_Res, Error>> {
        const { projectID, env, appID } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`${basePath(projectID, env, appID)}/sources`, { signal })).pipe(
                map(this.validator.findSources),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Create app setting mount
     */
    async createOne(
        request: AppSettingMounts_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_CreateOne_Res, Error>> {
        const { projectID, env, appID, ...payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(basePath(projectID, env, appID), toBody(payload), { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update app setting mount
     */
    async updateOne(
        request: AppSettingMounts_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_UpdateOne_Res, Error>> {
        const { projectID, env, appID, settingMountID, updateVer, ...payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `${basePath(projectID, env, appID)}/${settingMountID}`,
                    { ...toBody(payload), updateVer },
                    { signal },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Enable or disable app setting mount
     */
    async updateStatus(
        request: AppSettingMounts_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_UpdateStatus_Res, Error>> {
        const { projectID, env, appID, settingMountID, updateVer, status } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `${basePath(projectID, env, appID)}/${settingMountID}/status`,
                    { status, updateVer },
                    { signal },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete app setting mount
     */
    async deleteOne(request: AppSettingMounts_DeleteOne_Req): Promise<Result<AppSettingMounts_DeleteOne_Res, Error>> {
        const { projectID, env, appID, settingMountID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${basePath(projectID, env, appID)}/${settingMountID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
