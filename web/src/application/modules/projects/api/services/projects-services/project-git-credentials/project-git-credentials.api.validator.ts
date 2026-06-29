import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ProjectGitCredentials_FindManyBranches_Res,
    ProjectGitCredentials_FindManyPaginated_Res,
    ProjectGitCredentials_FindManyPullRequests_Res,
    ProjectGitCredentials_FindManyRepos_Res,
} from "./project-git-credentials.api.contracts";

const GitCredentialSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(GitCredentialSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const GitRepoSchema = z.object({
    id: z.string(),
    name: z.string(),
    fullName: z.string(),
    defaultBranch: z.string(),
    cloneURL: z.string(),
    gitURL: z.string(),
});

const FindManyReposSchema = z.object({
    data: z.array(GitRepoSchema),
    meta: PagingMetaApiSchema,
});

const GitBranchSchema = z.object({
    name: z.string(),
    sha: z.string(),
    ref: z.string(),
});

const FindManyBranchesSchema = z.object({
    data: z.array(GitBranchSchema),
    meta: PagingMetaApiSchema,
});

const GitPullRequestSchema = z.object({
    id: z.string(),
    number: z.number(),
    title: z.string(),
    state: z.string(),
    branch: z.string(),
    sha: z.string(),
    ref: z.string(),
    author: z.string(),
    htmlURL: z.string(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
});

const FindManyPullRequestsSchema = z.object({
    data: z.array(GitPullRequestSchema),
    meta: PagingMetaApiSchema,
});

export class ProjectGitCredentialsApiValidator {
    findManyPaginated = (response: AxiosResponse): ProjectGitCredentials_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findManyRepos = (response: AxiosResponse): ProjectGitCredentials_FindManyRepos_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyReposSchema,
        });

        return { data, meta };
    };

    findManyBranches = (response: AxiosResponse): ProjectGitCredentials_FindManyBranches_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyBranchesSchema,
        });

        return { data, meta };
    };

    findManyPullRequests = (response: AxiosResponse): ProjectGitCredentials_FindManyPullRequests_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPullRequestsSchema,
        });

        return { data, meta };
    };
}
