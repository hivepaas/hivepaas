import { type Prettify } from "@infrastructure/utility-types";

export interface Meta {
    code?: string;
    message?: string;
}

export type ApiResponseBase<TData, TMeta extends object | undefined = undefined> = TMeta extends undefined
    ? {
          data: TData;
          meta?: Meta | null;
      }
    : {
          data: TData;
          meta: Prettify<Meta & TMeta>;
      };

export type ApiResponse<TData, TMeta extends object | undefined = undefined> = ApiResponseBase<TData, TMeta>;
