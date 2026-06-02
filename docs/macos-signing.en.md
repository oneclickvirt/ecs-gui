# macOS Code Signing

[中文](macos-signing.zh.md)

GitHub Actions artifacts are unsigned by default, so macOS may show a security warning on first launch. For public distribution, sign the app with an Apple Developer ID certificate and optionally notarize it.

## Prepare

1. Create a `Developer ID Application` certificate in the Apple Developer portal.
2. Import the `.p12` certificate into your local keychain or a temporary CI keychain.
3. Prepare the signing identity, for example `Developer ID Application: Example Inc (TEAMID)`.

## Local Signing

```bash
codesign --force --deep --options runtime \
  --sign "Developer ID Application: Example Inc (TEAMID)" \
  goecs.app

codesign --verify --deep --strict --verbose=2 goecs.app
spctl --assess --type execute --verbose=2 goecs.app
```

## Notarization

```bash
ditto -c -k --keepParent goecs.app goecs.zip

xcrun notarytool submit goecs.zip \
  --apple-id "$APPLE_ID" \
  --team-id "$APPLE_TEAM_ID" \
  --password "$APPLE_APP_PASSWORD" \
  --wait

xcrun stapler staple goecs.app
spctl --assess --type execute --verbose=2 goecs.app
```

## CI Notes

- Store the `.p12` certificate as base64 in GitHub Secrets.
- Store the certificate password, Apple ID, Team ID, and app-specific password as separate secrets.
- In the macOS job, create a temporary keychain, import the certificate, sign, notarize, staple, and then package the app.
- Delete the temporary keychain at the end of the job to avoid leaving credentials on the runner.
