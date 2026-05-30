export const EAppHealthCheckRestMethod = {
    GET: "GET",
    POST: "POST",
    PUT: "PUT",
    PATCH: "PATCH",
    DELETE: "DELETE",
} as const;

export type EAppHealthCheckRestMethod = (typeof EAppHealthCheckRestMethod)[keyof typeof EAppHealthCheckRestMethod];
