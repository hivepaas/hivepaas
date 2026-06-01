import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { ProjectUserAccesses_FindOne_Res } from "~/projects/api/services/projects-services";

import { EUserRole } from "@application/shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const AccessActionsSchema = z.object({
    read: z.boolean(),
    execute: z.boolean().optional().default(false),
    write: z.boolean(),
    delete: z.boolean(),
});

const UserAccessSchema = z.object({
    id: z.string(),
    username: z.string(),
    email: z.string(),
    fullName: z.string(),
    photo: z.string().nullable(),
    role: z.nativeEnum(EUserRole),
    access: AccessActionsSchema,
});

const FindOneSchema = z.object({
    data: z.object({
        ownerAccess: UserAccessSchema,
        userAccesses: z.array(UserAccessSchema),
        moduleUserAccesses: z.array(UserAccessSchema),
        currentUserActions: z.object({
            canUpdateProjectUserAccesses: z.boolean(),
            canViewModuleUserAccesses: z.boolean(),
        }),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectUserAccessesApiValidator {
    findOne = (response: AxiosResponse): ProjectUserAccesses_FindOne_Res => {
        return parseApiResponse({
            response,
            schema: FindOneSchema,
        });
    };
}
