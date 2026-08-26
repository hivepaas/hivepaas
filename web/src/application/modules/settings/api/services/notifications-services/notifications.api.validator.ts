import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type Notifications_CreateOne_Res,
    type Notifications_DeleteOne_Res,
    type Notifications_FindManyPaginated_Res,
    type Notifications_FindOneById_Res,
    type Notifications_UpdateOne_Res,
    type Notifications_UpdateStatus_Res,
} from "./notifications.api.contracts";

const SettingsRefSchema = SettingsBaseEntitySchema.nullish().transform(value => value ?? undefined);

const NotificationViaEmailSchema = z.object({
    enabled: z.boolean(),
    useDefault: z.boolean(),
    sender: SettingsRefSchema,
    toProjectMembers: z
        .boolean()
        .nullish()
        .transform(v => v ?? false),
    toProjectOwners: z
        .boolean()
        .nullish()
        .transform(v => v ?? false),
    toAllAdmins: z
        .boolean()
        .nullish()
        .transform(v => v ?? false),
    toAddresses: z
        .array(z.string())
        .nullish()
        .transform(addresses => addresses ?? []),
});

const NotificationViaSlackSchema = z.object({
    enabled: z.boolean(),
    useDefault: z.boolean(),
    webhook: SettingsRefSchema,
});

const NotificationViaDiscordSchema = z.object({
    enabled: z.boolean(),
    useDefault: z.boolean(),
    webhook: SettingsRefSchema,
});

const NotificationViaTelegramSchema = z.object({
    enabled: z.boolean(),
    useDefault: z.boolean(),
    setting: SettingsRefSchema,
});

const NotificationViaLarkSchema = z.object({
    enabled: z.boolean(),
    useDefault: z.boolean(),
    webhook: SettingsRefSchema,
});

const NotificationEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    viaEmail: NotificationViaEmailSchema.nullish().transform(value => value ?? undefined),
    viaSlack: NotificationViaSlackSchema.nullish().transform(value => value ?? undefined),
    viaDiscord: NotificationViaDiscordSchema.nullish().transform(value => value ?? undefined),
    viaTelegram: NotificationViaTelegramSchema.nullish().transform(value => value ?? undefined),
    viaLark: NotificationViaLarkSchema.nullish().transform(value => value ?? undefined),
    minSendInterval: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    inherited: z.boolean().optional(),
});

/**
 * Find many notifications paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(NotificationEntitySchema),
    meta: PagingMetaApiSchema,
});

/**
 * Find one notification by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: NotificationEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

/**
 * Create one notification API response schema
 */
const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class NotificationsApiValidator {
    /**
     * Validate and transform find many notifications paginated API response
     */
    findManyPaginated = (response: AxiosResponse): Notifications_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data,
            meta,
        };
    };

    /**
     * Validate and transform find one notification by id API response
     */
    findOneById = (response: AxiosResponse): Notifications_FindOneById_Res => {
        const { data } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform create one notification API response
     */
    createOne = (response: AxiosResponse): Notifications_CreateOne_Res => {
        const { data } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform update one notification API response
     */
    updateOne = (response: AxiosResponse): Notifications_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform update notification status API response
     */
    updateStatus = (response: AxiosResponse): Notifications_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform delete one notification API response
     */
    deleteOne = (response: AxiosResponse): Notifications_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return {
            data: {
                type: "success",
            },
        };
    };
}
