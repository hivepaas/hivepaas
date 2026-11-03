import type { ComboboxOption } from "@application/shared/components";

export type NotificationSettingsRef = Record<"id" | "name", string>;

export type NotificationSettingsSource = {
    options: ComboboxOption<NotificationSettingsRef>[];
    isFetching?: boolean;
    isRefreshing?: boolean;
    onSearch?: (search: string) => void;
    onRefresh?: () => void;
};

export type NotificationSettingsSources = {
    success: NotificationSettingsSource;
    failure: NotificationSettingsSource;
};

export type NotificationSettingsManageLink = {
    to: string;
    label: string;
};
