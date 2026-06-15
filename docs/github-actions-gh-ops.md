# GitHub Docs + gh 直接发布/排障手册（CI + Tag 发布）

目标：把 CI 失败诊断与发布操作固定成可复用流程，结合 GitHub 官方文档与 `gh` 命令快速定位问题并发布。

## 一、前置
- 在仓库根目录执行并确认 GitHub CLI 已登录：
  - `gh auth status`
- 推荐先确认当前仓库与默认分支：
  - `gh repo view --json nameWithOwner,defaultBranchRef`

## 二、日常 CI 诊断（配合官方文档）
官方文档：
- [GitHub Actions 文档](https://docs.github.com/actions)
- [查看与排查工作流日志](https://docs.github.com/en/actions/managing-workflow-runs)
- [gh 查看/管理 workflow runs](https://cli.github.com/manual/gh_run)

1. 查看最近运行（快速判断失败）
```bash
gh run list --workflow ci.yml --limit 20
```
2. 查看某次运行详情
```bash
gh run view <RUN_ID> --json name,status,conclusion,url
```
3. 查看完整日志（定位 `422 already_exists` / 上传失败之类）
```bash
gh run view <RUN_ID> --log
```
4. 按 job 维度查看失败点
```bash
gh run view <RUN_ID> --json jobs | jq '.jobs[] | {name, conclusion, status, steps: .steps[] | {name,conclusion,status}}'
```
5. 观看运行过程（可持续观察）
```bash
gh run watch <RUN_ID>
```

## 三、发布流程（直接发布）
项目当前为「推送 tag 即触发发布」：
- 工作流：`ci.yml`（`Build` job + tag 触发）
- 发布方式：`goreleaser release --clean`

官方文档：
- [创建与管理标签](https://docs.github.com/en/repositories/working-with-files/managing-files-in-your-repository/managing-tags-in-a-repository)
- [管理发布（Release）](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)
- [gh 发布命令手册](https://cli.github.com/manual/gh_release)

1. 发起一个发布 Tag
```bash
git switch main
git pull --ff-only
git tag -a v0.2.2 -m "release: v0.2.2"
git push origin v0.2.2
```
2. 观察对应 CI 运行是否成功
```bash
gh run list --workflow ci.yml --branch main --limit 5
```
3. 成功后确认 Release 与产物
```bash
gh release view v0.2.2 --json tagName,name,assets
```

## 四、重复发布/并发冲突清理（你刚刚遇到的 422）
若重试同名 tag 触发重复上传，常见报错是：
- `422 Validation Failed ... already_exists`

官方文档：
- [GitHub Releases API](https://docs.github.com/rest/releases/releases)

处理流程：
1. 删除对应临时 Release（如需）
```bash
gh release delete <tag> --yes
```
2. 删除本地与远端 tag（如需重放）
```bash
git tag -d <tag>
git push origin --delete <tag>
```
3. 重新打同名 tag 并再次推送（按上方“发布流程”操作）

## 五、补充：常用 gh 快捷
- 重新触发失败 run：
```bash
gh run rerun <RUN_ID>
```
- 仅重跑失败 jobs：
```bash
gh run rerun --failed <RUN_ID>
```
- 列出 release（便于对账）
```bash
gh release list --limit 50
```
- 按 tag 查 release 详情：
```bash
gh api repos/neko233-com/bytemsg233/releases/tags/<tag>
```
