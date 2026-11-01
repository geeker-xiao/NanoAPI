# Changelog

All notable changes to NanoAPI are documented here. The project follows
[Semantic Versioning](https://semver.org/).

## [1.0.1] - 2026-11-01

本次更新带来大量新功能与界面重构：保存的请求、鉴权、请求体类型下拉、多标签页、外观设置，以及浅色主题迁移至 MoteDeck 风格并配套协调的暗色主题。

### ✨ 新增功能
- **保存的请求（Saved Requests）**：可将当前请求按名称保存、打开、删除；保存内容涵盖 URL、请求方法、Query 参数、Headers、鉴权方式与请求体，方便日后一键复用。
- **鉴权支持（Auth）**：新增 `None / Bearer token / Basic auth / API key` 四种模式；手动填写了同名字段时，手动值优先于鉴权自动生成的字段。
- **请求体类型下拉**：raw 模式下可在 `json / xml / html / javascript / text` 之间切换，切换时编辑器内的示例内容会同步更新。
- **form-data 文件字段**：原 multipart 模式重命名为 form-data，每个字段可为 text 或 file；选择 file 时会弹出系统文件选择器。
- **多标签页**：每个打开的请求拥有独立状态；标签栏支持「多行 / 单行」两种布局，可在设置中心切换（单行下 New 仅显示 `+`，方法徽章更紧凑，关闭按钮合并进标签）。
- **外观设置中心**：可切换主题模式（浅色 / 深色 / 跟随系统）、调整字体、设置背景（纯色或图片）。

### 🎨 界面与主题
- 浅色主题迁移至 MoteDeck 风格（浅灰背景 `#F5F5F5`、白色卡片、近黑主色），并配套一套协调的暗色主题。
- 请求栏重构：method 与 URL 融合为一个组合框，Send 为独立按钮；**URL 始终单行显示**，不再因聚焦而跨行展开，与 method、Send 保持在同一行。

### 📄 文档与许可
- 新增 **MIT 许可证**。
- 完善 README：补充 CI / Release 工作流说明、各平台产物命名规则与发布步骤。

### 💾 数据存储变化
NanoAPI 现在会在用户配置目录（`%UserConfigDir%/nanoapi/`，macOS / Linux 为对应配置目录）下持久化：
- `settings.json`：外观设置（主题、字体、背景、标签栏布局等）。
- `requests.json`：已保存的请求，以及上次打开的标签页快照（每 5 秒自动保存，串行写盘）。
- 兼容性：旧版 `multipart` 模式读取时自动映射为 `form-data`；raw 类型为空时回退为自动推断。

## [1.0.0] - 2026-10-29

- 首个公开发布版本。
- 原生 HTTP 客户端：方法选择、URL、发送、超时、跟随重定向、流式响应、下载到文件。
- Query 参数、Headers、请求体（none / raw / form / multipart）编辑。
- 左侧历史记录，可按分组管理。
- GitHub Actions：CI 校验与三平台（Windows / macOS / Linux）自动构建发布。
