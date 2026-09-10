import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type { AuditLogScope } from "~/operations/domain";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AuditLogs_FindManyPaginated_Req,
    AuditLogs_FindManyPaginated_Res,
    AuditLogs_FindOneById_Req,
    AuditLogs_FindOneById_Res,
    AuditLogs_FindTypes_Req,
    AuditLogs_FindTypes_Res,
} from "./audit-logs.api.contracts";
import type { AuditLogsApiValidator } from "./audit-logs.api.validator";

export function resolveAuditLogsEndpoint(scope?: AuditLogScope): string {
    if (!scope || scope.type === "global") {
        return "/system/audit-logs";
    }

    switch (scope.type) {
        case "project":
            return `/projects/${scope.projectID}/audit-logs`;
        case "project-env":
            return `/projects/${scope.projectID}/${scope.projectEnvID}/audit-logs`;
        case "app":
            return `/projects/${scope.projectID}/${scope.projectEnvID}/apps/${scope.appID}/audit-logs`;
        default:
            return "/system/audit-logs";
    }
}

export class AuditLogsApi extends BaseApi {
    public constructor(private readonly validator: AuditLogsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AuditLogs_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AuditLogs_FindManyPaginated_Res, Error>> {
        const {
            scope,
            pagination,
            sorting,
            search,
            type,
            source,
            result,
            actorID,
            resourceID,
            fromDate,
            toDate,
            projectID,
            appID,
            scopeOnly,
        } = request.data;
        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination)
            .sorting(sorting)
            .search(search)
            .filterBy({
                type,
                source,
                result,
                actorId: actorID,
                resourceId: resourceID,
                fromDate: fromDate ? [fromDate] : undefined,
                toDate: toDate ? [toDate] : undefined,
                projectId: projectID ? [projectID] : undefined,
                appId: appID ? [appID] : undefined,
                scopeOnly: scopeOnly ? [true] : undefined,
            });

        const url = resolveAuditLogsEndpoint(scope);

        return lastValueFrom(
            from(
                this.client.v1.get(url, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(data => Ok(data)),
                catchError(err => of(Err(parseApiError(err)))),
            ),
        );
    }

    async findOneById(
        request: AuditLogs_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AuditLogs_FindOneById_Res, Error>> {
        const { scope, itemID } = request.data;
        const baseUrl = resolveAuditLogsEndpoint(scope);
        const url = `${baseUrl}/${itemID}`;

        return lastValueFrom(
            from(
                this.client.v1.get(url, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(data => Ok(data)),
                catchError(err => of(Err(parseApiError(err)))),
            ),
        );
    }

    async findTypes(
        request: AuditLogs_FindTypes_Req = { data: {} },
        signal?: AbortSignal,
    ): Promise<Result<AuditLogs_FindTypes_Res, Error>> {
        const { scope } = request.data;
        const baseUrl = resolveAuditLogsEndpoint(scope);
        const url = `${baseUrl}/types`;

        return lastValueFrom(
            from(
                this.client.v1.get(url, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findTypes),
                map(data => Ok(data)),
                catchError(err => of(Err(parseApiError(err)))),
            ),
        );
    }
}
