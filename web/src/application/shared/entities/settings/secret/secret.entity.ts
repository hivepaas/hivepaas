import { SettingsBaseEntity, SettingType } from '../settings.base.entity';

export interface SecretEntity extends SettingsBaseEntity {
  type: SettingType.Secret;
  key: string;
}
