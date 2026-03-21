import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ProjectGitCredentials_FindManyPaginated_Res,
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
}
