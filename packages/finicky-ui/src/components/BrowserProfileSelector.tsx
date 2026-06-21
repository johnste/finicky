import { XIcon } from "./icons/X";
import type { BrowserProfile, BrowserOptions, BrowserProfileCustom } from "../types";
import styles from "./BrowserProfileSelector.module.css";

const CUSTOM = "__custom__";

interface Props {
  value: BrowserProfile;
  custom: BrowserProfileCustom;
  browsers: BrowserOptions;
  disabled?: boolean;
  required?: boolean;
  onChange?: (value: BrowserProfile, custom: BrowserProfileCustom, committed: boolean) => void;
  onRequestProfiles?: (b: string) => void;
}

interface BrowserFieldProps {
  browser: string;
  isCustom: boolean;
  required?: boolean;
  disabled?: boolean;
  placeholder: string;
  installedBrowsers: string[];
  onChange: (val: string) => void;
  onInput: (val: string) => void;
  onClear: () => void;
  onBlur: () => void;
}

function BrowserField({ browser, isCustom, required, disabled, placeholder, installedBrowsers, onChange, onInput, onClear, onBlur }: BrowserFieldProps) {
  if (isCustom) {
    return (
      <>
        <input
          className={styles.browserInput}
          type="text"
          placeholder="e.g. Firefox"
          value={browser}
          onChange={(e) => onInput(e.target.value)}
          onBlur={onBlur}
          disabled={disabled}
        />
        <button className={styles.clearBtn} onClick={onClear} aria-label="Clear browser">
          <XIcon />
        </button>
      </>
    );
  }
  return (
    <div className={`${styles.selectWrapper}${required && !browser ? " " + styles.empty : ""}`}>
      <select className={styles.browserDropdown} value={browser} onChange={(e) => onChange(e.target.value)} disabled={disabled}>
        <option value="">{placeholder}</option>
        {installedBrowsers.map((b) => <option key={b} value={b}>{b}</option>)}
        <option value={CUSTOM}>Custom...</option>
      </select>
    </div>
  );
}

interface ProfileFieldProps {
  profile: string;
  isProfileCustom: boolean;
  disabled?: boolean;
  profileOptions: string[];
  onChange: (val: string) => void;
  onInput: (val: string) => void;
  onClear: () => void;
  onBlur: () => void;
}

function ProfileField({ profile, isProfileCustom, disabled, profileOptions, onChange, onInput, onClear, onBlur }: ProfileFieldProps) {
  if (isProfileCustom) {
    return (
      <>
        <input
          className={styles.browserInput}
          type="text"
          placeholder="Profile name"
          value={profile}
          onChange={(e) => onInput(e.target.value)}
          onBlur={onBlur}
          disabled={disabled}
        />
        <button className={styles.clearBtn} onClick={onClear} aria-label="Clear profile">
          <XIcon />
        </button>
      </>
    );
  }
  return (
    <div className={styles.selectWrapper}>
      <select className={styles.browserDropdown} value={profile} onChange={(e) => onChange(e.target.value)} disabled={disabled}>
        <option value="">No profile</option>
        {profileOptions.map((p) => <option key={p} value={p}>{p}</option>)}
        <option value={CUSTOM}>Custom...</option>
      </select>
    </div>
  );
}

export function BrowserProfileSelector({
  value,
  custom,
  browsers,
  disabled,
  required,
  onChange = () => {},
  onRequestProfiles = () => {},
}: Props) {
  const { browser, profile } = value;
  const profileOptions = browsers.profiles[browser] ?? [];

  function handleBrowserSelect(val: string) {
    if (val === CUSTOM) {
      onChange({ browser: "", profile: "" }, { browser: true, profile: false }, false);
    } else {
      if (val && browsers.profiles[val] === undefined) onRequestProfiles(val);
      onChange({ browser: val, profile: "" }, { browser: false, profile: false }, true);
    }
  }

  function handleProfileSelect(val: string) {
    if (val === CUSTOM) {
      onChange({ browser, profile: "" }, { browser: custom.browser, profile: true }, false);
    } else {
      onChange({ browser, profile: val }, { browser: custom.browser, profile: false }, true);
    }
  }

  return (
    <div className={styles.browserSelectRow}>
      <BrowserField
        browser={browser}
        isCustom={custom.browser}
        required={required}
        disabled={disabled}
        placeholder={required ? "Select browser" : "System default"}
        installedBrowsers={browsers.installed}
        onChange={handleBrowserSelect}
        onInput={(val) => onChange({ browser: val, profile }, { browser: true, profile: custom.profile }, false)}
        onClear={() => onChange({ browser: "", profile: "" }, { browser: false, profile: false }, true)}
        onBlur={() => onChange({ browser, profile }, custom, true)}
      />
      {!custom.browser && browser && (profileOptions.length > 0 || custom.profile) && (
        <ProfileField
          profile={profile}
          isProfileCustom={custom.profile}
          disabled={disabled}
          profileOptions={profileOptions}
          onChange={handleProfileSelect}
          onInput={(val) => onChange({ browser, profile: val }, { browser: custom.browser, profile: true }, false)}
          onClear={() => onChange({ browser, profile: "" }, { browser: custom.browser, profile: false }, true)}
          onBlur={() => onChange({ browser, profile }, custom, true)}
        />
      )}
    </div>
  );
}
