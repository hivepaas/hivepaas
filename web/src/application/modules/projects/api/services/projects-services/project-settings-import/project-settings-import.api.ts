import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectSettingsImport_Import_Req,
    ProjectSettingsImport_Import_Res,
} from "./project-settings-import.api.contracts";
import type { ProjectSettingsImportApiValidator } from "./project-settings-import.api.validator";

export class ProjectSettingsImportApi extends BaseApi {
    public constructor(private readonly validator: ProjectSettingsImportApiValidator) {
        super();
    }

    async importSettings(
        request: ProjectSettingsImport_Import_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSettingsImport_Import_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/settings-import`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.importSettings),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
