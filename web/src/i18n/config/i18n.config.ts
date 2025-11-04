import i18next, { type ResourceLanguage } from "i18next";
import { initReactI18next } from "react-i18next";

import { DEFAULT_LANGUAGE } from "@i18n/constants";
import { EN_DICT, VI_DICT } from "@i18n/dictionaries";
import { type Language } from "@i18n/types";

export const i18nResources = {
    en: {
        translation: EN_DICT,
    },
    vi: {
        translation: VI_DICT,
    },
} as const satisfies Record<Language, ResourceLanguage>;

export default i18next.use(initReactI18next).init({
    lng: DEFAULT_LANGUAGE,
    fallbackLng: DEFAULT_LANGUAGE,
    defaultNS: "translation",
    resources: i18nResources,
    returnNull: false,
    interpolation: {
        escapeValue: false,
    },
});
