import { z } from "zod";
import { EAppDeploymentMethod, EBuildTool, ERepoType } from "~/projects/module-shared/enums";

const SettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const NotificationSchema = z.object({
    useDefaultOnSuccess: z.boolean(),
    success: SettingsRefSchema.optional(),
    useDefaultOnFailure: z.boolean(),
    failure: SettingsRefSchema.optional(),
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
        buildTool: z.nativeEnum(EBuildTool),
        repoType: z.nativeEnum(ERepoType),
        repoUrl: z.string().min(1, "Repository URL is required"),
        repoRef: z.string().min(1, "Branch / Commit is required"),
        credentials: SettingsRefSchema.optional(),
        dockerfilePath: z.string().optional(),
        imageName: z.string().optional(),
        imageTags: z.string().optional(),
        pushToRegistry: SettingsRefSchema.optional(),
    }),
});

const ImageMethodSchema = BaseDeploymentSettingsSchema.extend({
    activeMethod: z.literal(EAppDeploymentMethod.Image),
    image: z.string().min(1, "Docker Image is required"),
    registryAuth: SettingsRefSchema.optional(),
});

export const AppConfigDeploymentSettingsFormSchema = z.discriminatedUnion("activeMethod", [
    RepoMethodSchema,
    ImageMethodSchema,
]);

export type AppConfigDeploymentSettingsFormSchemaInput = z.input<typeof AppConfigDeploymentSettingsFormSchema>;
export type AppConfigDeploymentSettingsFormSchemaOutput = z.output<typeof AppConfigDeploymentSettingsFormSchema>;
