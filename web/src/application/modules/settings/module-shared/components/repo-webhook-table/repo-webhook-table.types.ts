export type RepoWebhookTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };
