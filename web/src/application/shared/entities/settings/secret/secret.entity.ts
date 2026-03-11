import { type ESettingType } from "../../../enums";
import { type SettingsBaseEntity } from '../settings.base.entity';

export interface SecretEntity extends SettingsBaseEntity {
  type: ESettingType.Secret;
  key: string;
}
