import { z } from "zod";
import { EJoinNodeMethod } from "~/cluster/module-shared/enums";

export const JoinNewNodeFormSchema = z
    .discriminatedUnion("method", [
        z.object({
            method: z.literal(EJoinNodeMethod.RunCommandManually),
            joinAsManager: z.boolean(),
        }),
        z.object({
            method: z.literal(EJoinNodeMethod.RunCommandViaSSH),
            joinAsManager: z.boolean(),
            host: z
                .string({
                    required_error: "Host is required",
                })
                .min(1, "Host is required"),
            port: z
                .number({
                    required_error: "Port is required",
                })
                .min(1, "Port is required"),
            user: z
                .string({
                    required_error: "User is required",
                })
                .min(1, "User is required"),
            sshKey: z
                .object({
                    id: z.string(),
                    name: z.string(),
                })
                .nullable(),
        }),
    ])
    .superRefine((arg, ctx) => {
        if (arg.method === EJoinNodeMethod.RunCommandViaSSH && !arg.sshKey) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "SSH key is required",
            });
        }
    });

export type JoinNewNodeFormInput = z.input<typeof JoinNewNodeFormSchema>;
export type JoinNewNodeFormOutput = z.output<typeof JoinNewNodeFormSchema>;
