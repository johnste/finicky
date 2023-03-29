import fs from "fs"
import path from "path"

export function chromeFolder() {
  const home = process.env.HOME;

  if (!home) throw new Error("HOME env variable not set");

  if (process.platform === "darwin")
    return path.join(
      home,
      "Library",
      "Application Support",
      "Google",
      "Chrome"
    );
  if (process.platform === "linux")
    return path.join(home, ".config", "google-chrome");
  if (process.platform === "win32")
    return path.join(home, "AppData", "Local", "Google", "Chrome");

  throw new Error("Unsupported platform");
}

type ChromeState = {
  profile: {
    info_cache: {
      [key: string]: {
        user_name: string
      }
    }
  }
}

const chromeState: ChromeState = JSON.parse(fs.readFileSync(path.join(chromeFolder(), "Local State")).toString())

/**
 * Utility function to get the profile id for a given email
 */
export function chromeProfile(email: string) {
  const profile = Object.entries(chromeState.profile.info_cache).find(
    ([, profile]) => profile.user_name === email
  );

  if (!profile) throw new Error(`No profile found for ${email}`);

  return profile[0];
}