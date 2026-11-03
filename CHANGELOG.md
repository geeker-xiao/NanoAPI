# Changelog

All notable changes to NanoAPI are documented here. The project follows
[Semantic Versioning](https://semver.org/).

## [Unreleased] - 开发中

本次更新接入界面语言国际化，支持中英文实时切换。

### ✨ 新增功能
- **语言国际化（i18n）**：基于 MyGo 原生 i18n 能力，新增 `locales/en.json` 与 `locales/zh-Hans.json` 两套消息目录（ICU MessageFormat），由 `go tool mygo generate` 生成类型化访问函数。界面文字不再硬编码。当前共 197 条消息，界面文案已全量接入。
- **中英文切换**：Appearance 新增「语言」下拉，可在「跟随系统 / English / 简体中文」之间选择，**切换即时生效，无需重启**。
- **外观面板已完整接入翻译**：面板标题与说明、主题（浅色／深色）、字体、字号、标签颜色、请求标签布局（多行／单行）、背景方式（纯色／图片）、背景颜色、背景图片及其选择与清除按钮、关闭窗口选项，全部随语言切换。
- **请求区与响应区文案已接入翻译**：保存/删除请求两个对话框、查询参数与表单与 form-data 行编辑器（添加/启用/删除行、选择文件、未选择文件）、请求体编辑器（含「将以 X 发送」说明与 JSON/XML/HTML/JavaScript/纯文本 标签）、文件上传、选项标签、响应预览的全部不可用提示、历史导入导出按钮、超时增减按钮、标签关闭按钮、历史搜索框。
- **鉴权与分组文案已接入翻译**：Auth 标签页的类型下拉（无 / Bearer Token / Basic 认证 / API 密钥）、用户名、密码、键名、键值、发送到，以及各模式的行为说明；分组的新建/编辑/删除按钮与对话框、校验提示、「全部分组」「未分组」。鉴权模式的选择器存的是 `"Bearer token"` 等英文原值，只有显示名跟随语言。
- **框架文案自动跟随**：选中简体中文后，MyGo 自身的控件、文件对话框与原生菜单同步为官方中文译文。
- **系统托盘菜单随语言重建**：托盘的「显示 / 退出」条目会跟着语言改变，不再是固定中文。

### 🐛 修复
- **`MYGO_LOCALE` 被启动逻辑清除**：此前应用启动时若语言为「跟随系统」，会无条件调用 `SetLocale("")`，把通过 `MYGO_LOCALE`（含 `mygo dev -lang`、`en-XA` 伪语种排查）指定的本次运行语言覆盖掉。现在只在没有该环境变量干预时才回放。
- **外观选项可能被写入译文**：主题、标签布局、背景方式三个下拉原先直接把显示文字当作存储值，一旦翻译就会污染 `settings.json`。改为「存储值 / 显示名」分离，`"System"`、`"Multi-line"`、`"Solid color"` 等字面量保持不变。
- **首帧多余写盘**：外观各选项的变更检测初值为空，启动首帧会误判为「用户改过」而保存一次设置，现已在加载配置后播种。

### 🔧 说明
- 界面文案已全量接入消息目录（197 条）。剩余 23 处英文字面量为**刻意保留**：示例值占位符（如 `key`、`X-API-Key`、`eyJhbGciOi...`）、语法示例（`prefix-{{timestamp}}-{{random:12}}`）、协议与技术标识（`"GET"`、`application/x-www-form-urlencoded`、`User-Agent`、`Token`、窗口标题）、纯数据格式化。
- 中文译名约定：`Header`、`Content-Type`、`Authorization`、`Bearer`、`Basic`、`Query`、`Token`、`Cookie`、`MIME`、`JSON`/`XML`/`HTML`/`JavaScript`、`base64` 与图片格式名保持英文不译。
- 保存对话框预填的 `"Untitled request"` **确定保留英文**：它会写入 `requests.json` 成为持久化请求名，翻译会把当前语言烙进用户数据，并让重名检查把中英两种写法当成两个不同请求。标签页标题的「未命名」是纯显示、不落盘，仍随语言切换。
- `requestTabs` / `responseTabs` 由包级变量改为函数：语言相关的列表必须每帧重新求值，否则会被冻结在进程启动时的语言上。
- 存储格式未变：`tabBarDisplay`、`backgroundMode`、鉴权模式、`"默认分组"` 等**参与持久化与字面量比较的值一律不翻译**，仅翻译显示层。

### 💾 数据存储变化
- `settings.json` 新增 `language` 字段，取值为 `"System"`（默认）、`"en"` 或 `"zh-Hans"`。旧配置缺少该字段时自动回落为 `"System"`，无需迁移。

### 🛠️ 构建
- `mygo.json` 新增 `i18n` 段（`dir: locales`、`source: en`）。
- 新增校验命令：`go tool mygo i18n check`。

## [1.0.2] - 2026-11-01

本次更新修复了窗口高度偏低时部分界面内容被裁剪、无法查看的问题，统一引入滚动容器。

### 🐛 修复：矮窗口下的显示不全
- **侧边栏（历史与保存分组）**：整条侧边栏改为单一滚动容器，保存的分组与历史列表随窗口高度自适应滚动，矮窗口下不再被裁掉。
- **响应区（Response）**：请求体编辑器（Params / Headers / Auth / Body 等）限制最大高度并自带滚动，不再抢占响应文本区域空间；响应内容（Preview / Body / Headers / Metadata）改为带最小高度的滚动容器，任意窗口高度下都能完整显示并可滚动查看。

### 📄 文档
- README 发布示例 tag 更新为 `v1.0.2`。

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
