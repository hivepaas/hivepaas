import { type EN_DICT } from "@i18n/dictionaries";

type PlainObject = Record<string, unknown>;

type NestedDict<T extends PlainObject> = {
    [Key in keyof T]: T[Key] extends PlainObject ? NestedDict<T[Key]> : string;
};

// The dictionary has no entries yet, so this is {} until it has.
// eslint-disable-next-line @typescript-eslint/no-generated-empty-object-type
export type TranslationDict = NestedDict<typeof EN_DICT>;
