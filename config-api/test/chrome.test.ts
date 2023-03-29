import fs from 'fs';
import path from 'path';
import { chromeFolder, chromeProfile } from "../src/chrome";

test("no profile found", () => {
  expect(() => chromeProfile("")).toThrow();
});

test("profile found", () => {
  const localStateExists =
    fs.lstatSync(path.join(chromeFolder(), "Local State")).isFile();

  if (!localStateExists) return

  const chromeState = JSON.parse(fs.readFileSync(path.join(chromeFolder(), "Local State")).toString())

  for (const profile of Object.entries(chromeState.profile.info_cache)) {
    const email = (profile[1] as { user_name: string }).user_name
    expect(chromeProfile(email)).toBe(profile[0]);
  }
});