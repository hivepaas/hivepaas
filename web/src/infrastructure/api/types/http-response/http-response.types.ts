import type { AxiosResponse } from "axios";

/**
 * A response as the API services' validators read it: its body, which a schema checks, and its status and
 * headers. Only those: since 1.20 axios types a response by the request's body and query too, and a validator
 * taking a whole AxiosResponse - `any` there, the request being none of its business - is refused by the linter
 * (no-unsafe-argument) wherever a typed one is passed to it.
 */
export type ApiHttpResponse<T = unknown> = Pick<AxiosResponse<T>, "data" | "status" | "headers">;
