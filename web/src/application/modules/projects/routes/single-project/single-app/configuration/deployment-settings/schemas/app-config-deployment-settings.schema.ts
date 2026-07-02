import { z } from "zod";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { ESettingType } from "@application/shared/enums";

const SettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
    type: z.nativeEnum(ESettingType).optional(),
});

const OptionalSettingsRefSchema = SettingsRefSchema.nullish();

const NotificationSchema = z.object({
    successUseDefault: z.boolean(),
    success: OptionalSettingsRefSchema,
    failureUseDefault: z.boolean(),
    failure: OptionalSettingsRefSchema,
});

const BaseDeploymentSettingsSchema = z.object({
    command: z.string().optional(),
    workingDir: z.string().optional(),
    preDeploymentCommand: z.string().optional(),
    postDeploymentCommand: z.string().optional(),
    notification: NotificationSchema.optional(),
});

const RepoMethodSchema = BaseDeploymentSettingsSchema.extend({
    activeMethod: z.literal(EAppDeploymentMethod.Repo),
    repoSource: z.object({
        // buildTool: z.nativeEnum(EBuildTool),
        // repoType: z.nativeEnum(ERepoType),
        repoUrl: z.string().min(1, "Repository URL is required"),
        repoRef: z.string().min(1, "Branch is required"),
        commitHash: z.string().optional(),
        repoOptions: z.object({
            gitSubmodulesEnabled: z.boolean(),
            gitLfsEnabled: z.boolean(),
        }),
        credentials: OptionalSettingsRefSchema,
        dockerfilePath: z.string().optional(),
        imageName: z.string().optional(),
        imageTags: z.string().optional(),
        pushToRegistry: OptionalSettingsRefSchema,
    }),
});

const ImageMethodSchema = BaseDeploymentSettingsSchema.extend({
    activeMethod: z.literal(EAppDeploymentMethod.Image),
    imageSource: z.object({
        image: z.string().min(1, "Docker Image is required"),
        registryAuth: OptionalSettingsRefSchema,
    }),
});

export const AppConfigDeploymentSettingsFormSchema = z.discriminatedUnion("activeMethod", [
    RepoMethodSchema,
    ImageMethodSchema,
]);

export type AppConfigDeploymentSettingsFormSchemaInput = z.input<typeof AppConfigDeploymentSettingsFormSchema>;
export type AppConfigDeploymentSettingsFormSchemaOutput = z.output<typeof AppConfigDeploymentSettingsFormSchema>;
