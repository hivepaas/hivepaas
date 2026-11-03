export type ApiRequestBase<TData, TMeta extends object | undefined = undefined> = TMeta extends undefined
    ? {
          data: TData;
      }
    : {
          data: TData;
          meta: TMeta;
      };
