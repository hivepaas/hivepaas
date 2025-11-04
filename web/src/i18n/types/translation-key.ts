import { type EN_DICT } from "@i18n/dictionaries";

import { type FlattenObjectKeys } from "@infrastructure/utility-types";

export type TranslationKey = FlattenObjectKeys<typeof EN_DICT>;
