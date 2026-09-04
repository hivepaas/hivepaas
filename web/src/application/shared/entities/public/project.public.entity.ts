export type ProjectPublicEnv = {
    id: string;
    name: string;
    color: string;
};

// Declared as a type alias, not an interface: it gets an implicit index
// signature that way, which the Combobox option generic requires.
export type ProjectPublic = {
    id: string;
    name: string;
    envs: ProjectPublicEnv[];
};
