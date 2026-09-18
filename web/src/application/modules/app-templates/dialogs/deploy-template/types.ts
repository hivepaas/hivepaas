export interface DeployTemplateDialogOptions {
    templateName: string;
    initialVersion?: string;
    initialVariant?: string;
    initialEnv?: string;
}

export interface DeployTemplateDialogState {
    state: {
        mode: "open" | "closed";
        projectId: string | null;
        templateName: string | null;
    };
    props: DeployTemplateDialogOptions;
}
