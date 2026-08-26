export type EmailAccountTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
          env?: string;
      };
