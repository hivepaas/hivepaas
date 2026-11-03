import { z } from "zod";

export const SETTING_MOUNT_KEY_PATTERN = /^[a-z0-9]([a-z0-9-]{0,18}[a-z0-9])?$/;
const FILE_MODE_PATTERN = /^0?[0-7]{3}$/;

const FileRowSchema = z.object({
    part: z.string(),
    secret: z.boolean(),
    gated: z.boolean(),
    enabled: z.boolean(),
    path: z.string(),
    mode: z.string(),
    uid: z.string(),
    gid: z.string(),
});

export const AppSettingMountFormSchema = z
    .object({
        name: z.string().trim().regex(SETTING_MOUNT_KEY_PATTERN, "Lowercase letters, digits and hyphens, at most 20"),
        inheritable: z.boolean(),
        sourceType: z.string().min(1, "Choose what to mount from"),
        source: z.object({ id: z.string(), name: z.string() }).nullable(),
        files: z.array(FileRowSchema),
    })
    .superRefine((value, ctx) => {
        if (!value.source?.id) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Choose a setting", path: ["source"] });
        }
        const enabled = value.files.filter(file => file.enabled);
        if (enabled.length === 0) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Mount at least one part", path: ["files"] });
        }
        const seen = new Set<string>();
        value.files.forEach((file, index) => {
            if (!file.enabled) {
                return;
            }
            const path = file.path.trim();
            const message = !path.startsWith("/")
                ? "Use an absolute path"
                : seen.has(path)
                  ? "Another part of this entry has this path"
                  : "";
            if (message) {
                ctx.addIssue({ code: z.ZodIssueCode.custom, message, path: ["files", index, "path"] });
            }
            seen.add(path);
            if (!FILE_MODE_PATTERN.test(file.mode.trim())) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "An octal mode, such as 0400",
                    path: ["files", index, "mode"],
                });
            }
        });
    });

export type AppSettingMountFormInput = z.input<typeof AppSettingMountFormSchema>;
export type AppSettingMountFormOutput = z.output<typeof AppSettingMountFormSchema>;
