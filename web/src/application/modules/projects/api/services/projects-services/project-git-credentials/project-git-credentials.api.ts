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

export class ProjectGitCredentialsApi extends BaseApi {
    public constructor(private readonly validator: ProjectGitCredentialsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectGitCredentials_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/git-credentials`, {
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

    async findManyRepos(
        request: ProjectGitCredentials_FindManyRepos_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyRepos_Res, Error>> {
        const { projectID, itemID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/git-credentials/${itemID}/repositories`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyRepos),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findManyBranches(
        request: ProjectGitCredentials_FindManyBranches_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyBranches_Res, Error>> {
        const { projectID, itemID, owner, repo, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/git-credentials/${itemID}/repository/branches`, {
                    params: {
                        ...query.build(),
                        ...(owner !== undefined ? { owner } : {}),
                        repo,
                    },
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyBranches),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findManyPullRequests(
        request: ProjectGitCredentials_FindManyPullRequests_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGitCredentials_FindManyPullRequests_Res, Error>> {
        const { projectID, itemID, owner, repo, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/git-credentials/${itemID}/repository/pull-requests`, {
                    params: {
                        ...query.build(),
                        ...(owner !== undefined ? { owner } : {}),
                        repo,
                    },
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPullRequests),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
