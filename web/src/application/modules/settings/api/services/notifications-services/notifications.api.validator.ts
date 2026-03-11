import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type Notifications_CreateOne_Res,
    type Notifications_DeleteOne_Res,
    type Notifications_FindManyPaginated_Res,
    type Notifications_FindOneById_Res,
    type Notifications_UpdateOne_Res,
} from "./notifications.api.contracts";

const NotificationViaEmailSchema = z.object({
    sender: SettingsBaseEntitySchema.optional(),
    toProjectMembers: z.boolean(),
    toProjectOwners: z.boolean(),
    toAllAdmins: z.boolean(),
    toAddresses: z.array(z.string()),
});

const NotificationViaSlackSchema = z.object({
    webhook: SettingsBaseEntitySchema.optional(),
});

const NotificationViaDiscordSchema = z.object({
    webhook: SettingsBaseEntitySchema.optional(),
});

const NotificationEntitySchema = SettingsBaseEntitySchema.extend({
    type: z.literal(ESettingType.Notification),
    viaEmail: NotificationViaEmailSchema.optional(),
    viaSlack: NotificationViaSlackSchema.optional(),
    viaDiscord: NotificationViaDiscordSchema.optional(),
    minSendInterval: z.string(),
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
});

/**
 * Create one notification API response schema
 */
const CreateOneSchema = z.object({
    data: NotificationEntitySchema,
});

/**
 * Update one notification API response schema
 */
const UpdateOneSchema = z.object({
    data: NotificationEntitySchema,
});

/**
 * Delete one notification API response schema
 */
const DeleteOneSchema = z.object({
    type: z.literal("success"),
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
        const { data } = parseApiResponse({
            response,
            schema: UpdateOneSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform delete one notification API response
     */
    deleteOne = (response: AxiosResponse): Notifications_DeleteOne_Res => {
        const { type } = parseApiResponse({
            response,
            schema: DeleteOneSchema,
        });

        return {
            data: {
                type,
            },
        };
    };
}
