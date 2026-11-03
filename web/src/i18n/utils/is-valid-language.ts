import { LANGUAGES } from "@i18n/constants";
import { type Language } from "@i18n/types";

export function isValidLanguage(value: Language): boolean {
    return LANGUAGES.includes(value);
}
