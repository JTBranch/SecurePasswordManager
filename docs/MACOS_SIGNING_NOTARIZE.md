MacOS signing & notarization for CI

This document explains how to securely prepare Apple Developer signing credentials and App Store Connect API keys, add them to GitHub Actions secrets, and run the repository CI to produce signed & notarized DMG releases.

Overview

- We sign the assembled `.app` with a Developer ID Application certificate (P12) and notarize the final DMG with `xcrun notarytool`.
- The GitHub Actions workflow will only run signing & notarization if the required secrets are present. No secrets are committed to the repo.

Required items

- Apple Developer Program account (Account Holder role required to create signing certificates).
- Developer ID Application certificate exported as `.p12` (including private key).
- App Store Connect API key (`.p8`) with Key ID and Issuer ID for `notarytool`.

Secrets to add to GitHub (repository-level Actions secrets)

- `MACOS_SIGNING_P12_BASE64` — base64 of your exported `.p12` file
- `MACOS_SIGNING_P12_PASSWORD` — password you set when exporting the `.p12`
- `MACOS_SIGNING_IDENTITY` — exact signing identity string (e.g. `Developer ID Application: Your Name (TEAMID)`)
- `APPLE_API_KEY_P8_BASE64` — base64 of the App Store Connect `.p8` file
- `APPLE_API_KEY_ID` — the Key ID (e.g. `ABC123XYZ`)
- `APPLE_API_ISSUER_ID` — the Issuer ID (GUID)

Step-by-step: create and upload credentials (macOS)

1. Create a CSR and obtain a Developer ID Application certificate

   - In Keychain Access: Certificate Assistant -> Request a Certificate From a Certificate Authority...
   - On developer.apple.com -> Certificates, IDs & Profiles -> Certificates -> + -> Developer ID -> Developer ID Application -> upload CSR -> download `.cer` and install it to Keychain.

2. Export the certificate and private key as .p12 (include private key)

   - In Keychain Access: select the certificate -> expand -> select the private key -> right-click -> Export items... -> choose `.p12` -> set a strong export password.

3. Create an App Store Connect API key (.p8)

   - App Store Connect -> Users and Access -> Keys -> Create API Key -> choose Developer role (or appropriate role) -> download `.p8` file. Note Key ID and Issuer ID.

4. Base64-encode artifacts (do this on your local machine)

```bash
base64 /path/to/DeveloperID.p12 > ~/tmp_signing.p12.base64
base64 /path/to/AuthKey_ABC123XYZ.p8 > ~/tmp_authkey.p8.base64
```

5. Upload secrets to GitHub (recommended: gh CLI)

```bash
cat ~/tmp_signing.p12.base64 | gh secret set MACOS_SIGNING_P12_BASE64 --repo JTBranch/SecurePasswordManager --body -
gh secret set MACOS_SIGNING_P12_PASSWORD --repo JTBranch/SecurePasswordManager --body 'your-p12-password'
gh secret set MACOS_SIGNING_IDENTITY --repo JTBranch/SecurePasswordManager --body 'Developer ID Application: Your Name (TEAMID)'
cat ~/tmp_authkey.p8.base64 | gh secret set APPLE_API_KEY_P8_BASE64 --repo JTBranch/SecurePasswordManager --body -
gh secret set APPLE_API_KEY_ID --repo JTBranch/SecurePasswordManager --body 'ABC123XYZ'
gh secret set APPLE_API_ISSUER_ID --repo JTBranch/SecurePasswordManager --body 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX'
```

6. Run the workflow

- Actions -> Build and upload macOS DMG -> Run workflow -> (optionally set `release_tag`) -> Run

Verification

- After the run completes, check the "Verify DMG" step in the workflow logs. You should see:
  - `codesign` indicating a valid signature (no "not signed at all").
  - `spctl` shows the app is accepted instead of rejected.
  - If notarization ran, `xcrun notarytool` output and `xcrun stapler validate` should succeed.

Cleanup & rotation

- If this was a one-time test, consider revoking the App Store Connect key or rotating your Developer ID certificate if you exported it insecurely.
- To remove secrets, go to Settings -> Secrets & variables -> Actions and delete the secrets.

Troubleshooting

- "no matching identity found": ensure `MACOS_SIGNING_IDENTITY` matches exactly the output of `security find-identity -v -p codesigning`.
- Notarization fails: verify Key ID and Issuer ID, and that the API key has the necessary role.
- Gatekeeper still blocks after notarization: ensure you stapled the ticket (`xcrun stapler staple`) and validated it.

Contact

- If you run into errors, paste the Actions run URL and the relevant `Verify DMG` step logs (do not paste secret contents). I'll help debug.
