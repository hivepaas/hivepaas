import { z } from "zod";
import {
    EAppHealthCheckGrpcStatus,
    EAppHealthCheckGrpcVersion,
    EAppHealthCheckRestMethod,
    EAppHealthCheckReturnBodyMode,
    EAppHealthCheckType,
} from "~/projects/module-shared/enums";

const NotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const VisibleRestMethodSchema = z.enum([
    EAppHealthCheckRestMethod.GET,
    EAppHealthCheckRestMethod.POST,
    EAppHealthCheckRestMethod.PUT,
]);

export const CreateOrEditAppHealthCheckFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        interval: z.string().trim().min(1, "Interval is required"),
        timeout: z.string().trim().min(1, "Timeout is required"),
        maxRetry: z.number().int().min(0, "Max retry must be greater than or equal to 0"),
        retryDelay: z.string().trim().min(1, "Retry delay is required"),
        healthcheckType: z.nativeEnum(EAppHealthCheckType),
        rest: z.object({
            url: z.string().trim(),
            method: z.nativeEnum(EAppHealthCheckRestMethod),
            contentType: z.string().trim(),
            body: z.string(),
            returnCode: z.string().trim(),
            returnBodyMode: z.nativeEnum(EAppHealthCheckReturnBodyMode),
            textExact: z.string(),
            textRegex: z.string(),
            jsonExact: z.string(),
            jsonContain: z.string(),
        }),
        grpc: z.object({
            version: z.nativeEnum(EAppHealthCheckGrpcVersion),
            addr: z.string().trim(),
            service: z.string().trim(),
            returnStatus: z.nativeEnum(EAppHealthCheckGrpcStatus),
        }),
        notification: z.object({
            successUseDefault: z.boolean(),
            success: NotificationRefSchema.optional(),
            failureUseDefault: z.boolean(),
            failure: NotificationRefSchema.optional(),
            minSendInterval: z.string().trim().min(1, "Min send interval is required"),
        }),
    })
    .superRefine((value, ctx) => {
        if (value.healthcheckType === EAppHealthCheckType.REST) {
            if (!value.rest.url) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "URL is required",
                    path: ["rest", "url"],
                });
            }

            if (!VisibleRestMethodSchema.safeParse(value.rest.method).success) {
                return;
            }

            if (!value.rest.returnCode && value.rest.returnBodyMode === EAppHealthCheckReturnBodyMode.Skipped) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "Return code or return body rule is required",
                    path: ["rest", "returnCode"],
                });
            }

            if (value.rest.returnBodyMode === EAppHealthCheckReturnBodyMode.Text && value.rest.textRegex) {
                try {
                    new RegExp(value.rest.textRegex);
                } catch {
                    ctx.addIssue({
                        code: z.ZodIssueCode.custom,
                        message: "Regular expression is invalid",
                        path: ["rest", "textRegex"],
                    });
                }
            }

            if (value.rest.returnBodyMode === EAppHealthCheckReturnBodyMode.JSON) {
                if (value.rest.jsonExact) {
                    try {
                        JSON.parse(value.rest.jsonExact);
                    } catch {
                        ctx.addIssue({
                            code: z.ZodIssueCode.custom,
                            message: "JSON exact must be valid JSON",
                            path: ["rest", "jsonExact"],
                        });
                    }
                }

                if (value.rest.jsonContain) {
                    try {
                        JSON.parse(value.rest.jsonContain);
                    } catch {
                        ctx.addIssue({
                            code: z.ZodIssueCode.custom,
                            message: "JSON contains must be valid JSON",
                            path: ["rest", "jsonContain"],
                        });
                    }
                }
            }
        }

        if (value.healthcheckType === EAppHealthCheckType.GRPC && !value.grpc.addr) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Address is required",
                path: ["grpc", "addr"],
            });
        }
    });

export type CreateOrEditAppHealthCheckFormInput = z.input<typeof CreateOrEditAppHealthCheckFormSchema>;
export type CreateOrEditAppHealthCheckFormOutput = z.output<typeof CreateOrEditAppHealthCheckFormSchema>;
