import type { AxiosResponse } from "axios";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppManaged_AddReplica_Req,
    AppManaged_AddReplica_Res,
    AppManaged_CreateRestorePoint_Req,
    AppManaged_CreateRestorePoint_Res,
    AppManaged_DeleteRestorePoint_Req,
    AppManaged_DeleteRestorePoint_Res,
    AppManaged_FindSettings_Req,
    AppManaged_FindSettings_Res,
    AppManaged_FindStatus_Req,
    AppManaged_FindStatus_Res,
    AppManaged_FindSwitch_Req,
    AppManaged_FindSwitch_Res,
    AppManaged_RefreshCredentials_Req,
    AppManaged_RefreshCredentials_Res,
    AppManaged_RemoveReplica_Req,
    AppManaged_RemoveReplica_Res,
    AppManaged_Restore_Req,
    AppManaged_Restore_Res,
    AppManaged_Switch_Req,
    AppManaged_Switch_Res,
    AppManaged_Unmanage_Req,
    AppManaged_Unmanage_Res,
    AppManaged_UpdateRestorePoint_Req,
    AppManaged_UpdateRestorePoint_Res,
    AppManaged_UpdateSettings_Req,
    AppManaged_UpdateSettings_Res,
} from "./app-managed.api.contracts";
import type { AppManagedApiValidator } from "./app-managed.api.validator";

const SUCCESS = { data: { type: "success" } } as const;

/** A managed app's engine: its settings and state, restores, restore points, and unmanaging. */
export class AppManagedApi extends BaseApi {
    constructor(private readonly validator: AppManagedApiValidator) {
        super();
    }

    private base(scope: { projectID: string; env: string; appID: string }): string {
        return `/projects/${scope.projectID}/${scope.env}/apps/${scope.appID}/managed`;
    }

    private run<T>(request: Promise<AxiosResponse>, read: (response: AxiosResponse) => T): Promise<Result<T, Error>> {
        return lastValueFrom(
            from(request).pipe(
                map(read),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findSettings(
        req: AppManaged_FindSettings_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppManaged_FindSettings_Res, Error>> {
        return this.run(this.client.v1.get(this.base(req.data), { signal }), this.validator.findSettings);
    }

    async findStatus(
        req: AppManaged_FindStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppManaged_FindStatus_Res, Error>> {
        return this.run(this.client.v1.get(`${this.base(req.data)}/status`, { signal }), this.validator.findStatus);
    }

    async updateSettings(req: AppManaged_UpdateSettings_Req): Promise<Result<AppManaged_UpdateSettings_Res, Error>> {
        return this.run(this.client.v1.put(this.base(req.data), req.data.payload), () => SUCCESS);
    }

    async refreshCredentials(
        req: AppManaged_RefreshCredentials_Req,
    ): Promise<Result<AppManaged_RefreshCredentials_Res, Error>> {
        return this.run(
            this.client.v1.post(`${this.base(req.data)}/refresh-credentials`, {}),
            this.validator.refreshCredentials,
        );
    }

    async unmanage(req: AppManaged_Unmanage_Req): Promise<Result<AppManaged_Unmanage_Res, Error>> {
        return this.run(this.client.v1.post(`${this.base(req.data)}/unmanage`, {}), () => SUCCESS);
    }

    async restore(req: AppManaged_Restore_Req): Promise<Result<AppManaged_Restore_Res, Error>> {
        return this.run(
            this.client.v1.post(`${this.base(req.data)}/restore`, req.data.payload),
            this.validator.restore,
        );
    }

    async addReplica(req: AppManaged_AddReplica_Req): Promise<Result<AppManaged_AddReplica_Res, Error>> {
        return this.run(
            this.client.v1.post(`${this.base(req.data)}/replicas`, req.data.payload),
            this.validator.addReplica,
        );
    }

    async removeReplica(req: AppManaged_RemoveReplica_Req): Promise<Result<AppManaged_RemoveReplica_Res, Error>> {
        const path = `${this.base(req.data)}/replicas/${encodeURIComponent(req.data.memberID)}`;
        return this.run(this.client.v1.delete(path), () => SUCCESS);
    }

    async findSwitch(
        req: AppManaged_FindSwitch_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppManaged_FindSwitch_Res, Error>> {
        return this.run(this.client.v1.get(`${this.base(req.data)}/switch`, { signal }), this.validator.findSwitch);
    }

    async switchConsumers(req: AppManaged_Switch_Req): Promise<Result<AppManaged_Switch_Res, Error>> {
        return this.run(
            this.client.v1.post(`${this.base(req.data)}/switch`, req.data.payload),
            this.validator.switchConsumers,
        );
    }

    async createRestorePoint(
        req: AppManaged_CreateRestorePoint_Req,
    ): Promise<Result<AppManaged_CreateRestorePoint_Res, Error>> {
        return this.run(
            this.client.v1.post(`${this.base(req.data)}/restore-points`, req.data.payload),
            this.validator.createRestorePoint,
        );
    }

    async updateRestorePoint(
        req: AppManaged_UpdateRestorePoint_Req,
    ): Promise<Result<AppManaged_UpdateRestorePoint_Res, Error>> {
        const path = `${this.base(req.data)}/restore-points/${encodeURIComponent(req.data.name)}`;
        return this.run(this.client.v1.put(path, req.data.payload), () => SUCCESS);
    }

    async deleteRestorePoint(
        req: AppManaged_DeleteRestorePoint_Req,
    ): Promise<Result<AppManaged_DeleteRestorePoint_Res, Error>> {
        const path = `${this.base(req.data)}/restore-points/${encodeURIComponent(req.data.name)}`;
        return this.run(this.client.v1.delete(path), () => SUCCESS);
    }
}
