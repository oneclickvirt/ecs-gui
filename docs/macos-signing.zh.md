# macOS 代码签名流程

[English](macos-signing.en.md)

GitHub Actions 默认产物未签名，因此首次打开可能出现安全提示。正式分发时建议使用 Apple Developer ID 证书签名并可选进行 notarization。

## 准备

1. 在 Apple Developer 账户中创建 `Developer ID Application` 证书。
2. 在本地钥匙串或 CI 临时钥匙串中导入 `.p12` 证书。
3. 准备签名身份名称，例如 `Developer ID Application: Example Inc (TEAMID)`。

## 本地签名

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

## CI 建议

- 将 `.p12` 证书以 base64 形式保存到 GitHub Secrets。
- 将证书密码、Apple ID、Team ID、App-specific password 分别保存为 secrets。
- 在 macOS job 中创建临时 keychain，导入证书，签名、notarize、staple 后再打包。
- CI 结束时删除临时 keychain，避免证书残留。
