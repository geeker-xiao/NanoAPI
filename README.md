# NanoAPI

[![CI](https://github.com/geeker-xiao/NanoAPI/actions/workflows/ci.yml/badge.svg)](https://github.com/geeker-xiao/NanoAPI/actions/workflows/ci.yml)
[![Release](https://github.com/geeker-xiao/NanoAPI/actions/workflows/release.yml/badge.svg)](https://github.com/geeker-xiao/NanoAPI/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/geeker-xiao/NanoAPI)](https://github.com/geeker-xiao/NanoAPI/releases/latest)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)

NanoAPI 是基于 [MyGo](https://github.com/egoist/mygo) 原生 UI 的轻量 API 桌面客户端。界面和 HTTP 执行层均为纯 Go，不需要 WebView 前端构建链。

## 下载

从 [Releases](https://github.com/geeker-xiao/NanoAPI/releases) 页面下载对应平台安装包：

- **Windows**：`NanoAPI-<version>-windows-amd64.zip`（内含 `nanoapi-setup-<version>-amd64.exe` 安装程序，以及 `NanoAPI.exe` 便携版）
- **macOS**：`NanoAPI-<version>-darwin-universal.tar.gz`（含 `NanoAPI-<version>-universal.dmg`，解压后挂载拖入应用程序；支持 Apple Silicon 与 Intel）
- **Linux**：`NanoAPI-<version>-linux-amd64.tar.gz`（含 `nanoapi_<version>_amd64.deb` 与 `nanoapi-<version>-linux-amd64.tar.gz`，后者带 `install.sh` 可安装到 `~/.local`）

> macOS 未签名版本首次打开需在“系统设置 → 隐私与安全性”中允许，或使用 `xattr -dr com.apple.quarantine /Applications/NanoAPI.app`。

## 功能

- Method 与 URL 融合在同一个输入框内，Send 为独立按钮；URL 默认单行显示（超出部分不展开），点击后展开为多行编辑
- 左右分栏布局：左侧历史与分组，右侧请求工作区
- 多标签页：每个标签是一个独立请求，各带自己的 Method、URL、Send、参数、Header、Body、选项和响应；可新建、切换、关闭
- 请求可保存：每个标签点 Save 即可自己命名保存，连同 URL、Query 参数、Header、鉴权信息、Body 与选项一起存到本机，左侧「Saved requests」随时重新打开
- 鉴权（Auth）独立配置：None / Bearer token / Basic auth / API key（可选随 Header 或 Query 发送），手写的同名 Header 或参数优先
- 请求标签两种排布，在 Appearance 中切换：默认**多行**（method 在上、名称在下，标签栏自动换行）；**单行**时标签栏横向滚动，method 收窄为小色条加缩写字号，关闭按钮并入标签内，右侧新建只显示 `+`
- 打开的请求标签自动保存在本机，关掉应用再打开会恢复到离开时的标签页
- GET、POST、PUT、PATCH、DELETE、HEAD、OPTIONS、PUSH 请求
- URL Query 参数与请求 Header 的添加、启用和删除
- User-Agent 独立配置，以及 Authorization、Cookie 等任意自定义 Header
- Header Value 支持固定值和多种随机值：UUID、ULID、随机字符串/整数、时间戳、ISO 时间、时间+随机、模板拼接
- GET、POST、PUT、PATCH、DELETE、HEAD、OPTIONS、PUSH 方法颜色区分
- Raw Body 内容类型下拉选择：`json`、`xml`、`html`、`javascript`、`text`（未指定时仍按内容自动推断）
- 切换内容类型时，编辑区示例同步替换为对应类型的样例（已手动编辑过的 Body 不会被覆盖）
- JSON / 文本 Body、URL 编码表单
- form-data（multipart/form-data）：每行字段可选 `text` / `file`，选 `file` 时右侧按钮弹出文件选择，支持文本与文件混合提交
- 响应 Body、Header、状态、耗时、大小查看
- 图片响应预览：PNG、JPEG、GIF、WebP、BMP、SVG
- HTML 响应预览：WebView 渲染，自动移除脚本和危险跳转
- JSON 自动格式化
- SSE / Chunked 流式响应，可随时取消
- 完整响应下载到文件，已完成响应另存
- 请求分组：新建、修改名称和描述、删除分组
- 历史记录按方法、URL、Header、Body、状态和分组搜索
- URL 地址栏支持多行显示，发送时自动忽略编辑换行
- 自定义背景颜色和背景图片，支持预览、铺满窗口和重启恢复
- 纸飞机应用图标，窗口和打包资源共用
- 标签字体颜色默认 `#111111`，可在 Appearance 中修改
- 浅色主题采用样例配色：页面底色 `#f5f5f5`、白色卡片、近黑主色 `#111111`、绿色强调 `#22c55e`、圆角 10
- 深色主题为配套版本：页面底色 `#111111`、卡片 `#1c1c1c`、浅色主按钮（黑字）、绿色强调不变；主题切换时默认背景色自动跟随
- HTTP 方法色随主题切换深浅变体，标签色对比度过低时自动回退为正文色
- 自定义字体、字号，设置自动保存
- 请求历史 JSON 导入、导出与去重合并，包含分组信息

## 环境

- Go 1.27.3+
- Windows、macOS 或 Linux
- MyGo v0.7.1

## 运行

```powershell
go run ./cmd/nanoapi
```

## 构建

```powershell
go tool mygo build -platform windows/amd64 -o build
```

Linux / macOS：

```sh
go tool mygo build -platform linux/amd64 -o build
go tool mygo build -platform darwin/universal -o build
```

## 从源码发布（GitHub Actions）

仓库内置两条工作流：

- `ci.yml`：每次推送与 PR 自动运行 `go test`、`go vet` 与 MyGo 编译检查。
- `release.yml`：推送形如 `v1.2.3` 的 tag 时，自动为 Windows / macOS / Linux 构建安装包，并创建 GitHub Release 附上各平台产物。

发布新版本只需：

```sh
git tag v1.0.1
git push origin v1.0.1
```

版本号取自 `mygo.json` 中的 `version` 字段，发布前请同步修改它。

## 测试与检查

```powershell
go test ./...
go test -race ./...
go vet ./...
go tool mygo vet ./...
```

## 代码结构

- `cmd/nanoapi/main.go`：MyGo 原生 UI、窗口和应用状态
- `cmd/nanoapi/groups.go`：请求分组、分组编辑和历史搜索
- `cmd/nanoapi/response_preview.go`：图片和 HTML 响应预览
- `internal/httpclient/client.go`：HTTP 请求、Multipart 上传、流式读取、下载落盘
- `internal/preview/html.go`：HTML 安全处理和预览包装
- `internal/history/store.go`：请求历史与分组 JSON 格式

## 历史文件

历史记录使用带版本号的 JSON 文件，可通过左侧 Export / Import 按钮迁移。导入会校验文件版本、分组引用和请求字段，并与当前历史去重合并。

外观设置保存在系统用户配置目录的 `nanoapi/settings.json`。旧版 `mygo-api-studio/settings.json` 会在首次启动时自动迁移。

## 许可证

本项目采用 [MIT 许可证](LICENSE)。
