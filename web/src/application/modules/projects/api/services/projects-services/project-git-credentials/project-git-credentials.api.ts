import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectGitCredentials_FindManyBranches_Req,
    ProjectGitCredentials_FindManyBranches_Res,
    ProjectGitCredentials_FindManyPaginated_Req,
    ProjectGitCredentials_FindManyPaginated_Res,
    ProjectGitCredentials_FindManyPullRequests_Req,
    ProjectGitCredentials_FindManyPullRequests_Res,
    ProjectGitCredentials_FindManyRepos_Req,
    ProjectGitCredentials_FindManyRepos_Res,
} from "./project-git-credentials.api.contracts";
import type { ProjectGitCredentialsApiValidator } from "./project-git-credentials.api.validator";

function getProjectGitCredentialsBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/git-credentials`;
    }

    return `/projects/${projectID}/git-credentials`;
}

export class ProjectGitCredentialsApi extends BaseApi {
    public constructor(private readonly validator: ProjectGitCredentialsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectGitCredentials_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectGitCredentialsBasePath(projectID, env), {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findManyPaginated(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findManyRepos(
        request: ProjectGitCredentials_FindManyRepos_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyRepos_Res, Error>> {
        const { projectID, env, itemID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectGitCredentialsBasePath(projectID, env)}/${itemID}/repositories`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findManyRepos(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findManyBranches(
        request: ProjectGitCredentials_FindManyBranches_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyBranches_Res, Error>> {
        const { projectID, env, itemID, owner, repo, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectGitCredentialsBasePath(projectID, env)}/${itemID}/repository/branches`, {
                    params: {
                        ...query.build(),
                        ...(owner !== undefined ? { owner } : {}),
                        repo,
                    },
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findManyBranches(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findManyPullRequests(
        request: ProjectGitCredentials_FindManyPullRequests_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyPullRequests_Res, Error>> {
        const { projectID, env, itemID, owner, repo, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectGitCredentialsBasePath(projectID, env)}/${itemID}/repository/pull-requests`, {
                    params: {
                        ...query.build(),
                        ...(owner !== undefined ? { owner } : {}),
                        repo,
                    },
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findManyPullRequests(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
