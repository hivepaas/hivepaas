import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectSettingsImport_Import_Req,
    ProjectSettingsImport_Import_Res,
} from "./project-settings-import.api.contracts";
import type { ProjectSettingsImportApiValidator } from "./project-settings-import.api.validator";

function getProjectSettingsImportBasePath(projectID: string, env?: string): string {
    if (env && env !== "all") {
        return `/projects/${projectID}/${encodeURIComponent(env)}/settings-import`;
    }

    return `/projects/${projectID}/settings-import`;
}

export class ProjectSettingsImportApi extends BaseApi {
    public constructor(private readonly validator: ProjectSettingsImportApiValidator) {
        super();
    }

    async importSettings(
        request: ProjectSettingsImport_Import_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSettingsImport_Import_Res, Error>> {
        const { projectID, env, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectSettingsImportBasePath(projectID, env), payload, {
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
