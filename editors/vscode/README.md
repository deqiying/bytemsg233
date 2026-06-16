# ByteMsg233 VSCode Extension

为 `.bmsg` 文件提供基础语言支持：

- 语法高亮
- `// "中文" | "English"` i18n 注释高亮
- `message` / `enum` / `schema` / `package` 关键字高亮
- `@type` / `@alias` 注解高亮
- 基础 snippets

## 本地调试

```bash
cd editors/vscode
npm install
```

在 VSCode 里打开这个目录后，按 `F5` 启动 Extension Development Host。

## 本地打包

```bash
cd editors/vscode
npm run package
```

成功后会生成类似 `bytemsg233-vscode-0.1.0.vsix` 的安装包。

## 发布到 VS Marketplace

1. 在 [Visual Studio Marketplace](https://marketplace.visualstudio.com/manage) 创建 Publisher
2. 确认 [package.json](/D:/Code/neko233-Projects/bytemsg233/editors/vscode/package.json) 里的 `publisher` 与 Marketplace Publisher ID 一致
3. 创建 Personal Access Token，至少勾选 `Marketplace > Manage`
4. 本地发布：

```bash
cd editors/vscode
npx vsce publish -p <VSCE_PAT>
```

5. 或者使用仓库里的 GitHub Actions 工作流，配置 `VSCE_PAT` secret 后手动触发

## 备注

- 当前扩展重点支持 `.bmsg`
- `.bmsg.yaml` 仍建议交给 VSCode 内建 YAML 能力处理
