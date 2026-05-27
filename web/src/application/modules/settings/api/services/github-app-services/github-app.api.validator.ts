import { type AxiosResponse } from "axios";
import { z } from "zod";
import { GithubAppSettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    GithubApp_BeginManifestFlow_Res,
    GithubApp_BeginReprovision_Res,
    GithubApp_CreateOne_Res,
    GithubApp_DeleteOne_Res,
    GithubApp_FindManyPaginated_Res,
    GithubApp_FindOneById_Res,
    GithubApp_ListInstallations_Res,
    GithubApp_TestConnection_Res,
    GithubApp_UpdateOne_Res,
    GithubApp_UpdateStatus_Res,
} from "./github-app.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(GithubAppSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: GithubAppSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
        callbackURL: z.string().optional(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

const BeginManifestFlowSchema = z.object({
    data: z.object({
        redirectURL: z.string(),
        settingId: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const BeginReprovisionSchema = z.object({
    data: z.object({
        redirectURL: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const ListInstallationsSchema = z.object({
    data: z.array(
        z.object({
            id: z.coerce.number(),
            nodeId: z.string().default(""),
            appId: z.coerce.number(),
            appSlug: z.string().default(""),
        }),
    ),
    meta: PagingMetaApiSchema,
});

export class GithubAppApiValidator {
    findManyPaginated = (response: AxiosResponse): GithubApp_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: AxiosResponse): GithubApp_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: AxiosResponse): GithubApp_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: AxiosResponse): GithubApp_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: AxiosResponse): GithubApp_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): GithubApp_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    testConnection = (response: AxiosResponse): GithubApp_TestConnection_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    beginManifestFlow = (response: AxiosResponse): GithubApp_BeginManifestFlow_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: BeginManifestFlowSchema,
        });

        return { data, meta };
    };

    beginReprovision = (response: AxiosResponse): GithubApp_BeginReprovision_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: BeginReprovisionSchema,
        });

        return { data, meta };
    };

    listInstallations = (response: AxiosResponse): GithubApp_ListInstallations_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: ListInstallationsSchema,
        });

        return { data, meta };
    };
}
