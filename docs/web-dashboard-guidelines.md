# Web Dashboard 规范

Commory Web 以 `references/art-design-pro` 为 UI 模板基线，以 `references/memos/web` 作为产品组织参考。参考资料只用于学习结构和体验，不复制源码实现。

## 模板基线

- 页面布局、导航、工作台标签、设置面板、暗色模式和主题变量遵循 Art Design Pro 现有模式。
- 新 dashboard 页面优先复用 `web/src/components/core`、Element Plus、Pinia、Vue Router 和现有 style tokens。
- 不新增与模板冲突的全局样式系统；确需新增时先写入局部组件或明确扩展现有变量。

## 国际化

- 用户可见文案必须进入 `web/src/locales/langs/zh.json` 和 `web/src/locales/langs/en.json`。
- Vue 模板使用 `$t(...)`，composition API 中使用 `useI18n()`。
- 不在 `.vue`、`.ts` 中硬编码中文或英文用户文案。
- 路由 `meta.title` 应使用 i18n key；`formatMenuTitle()` 负责把 key 转成当前语言文案，缺失 key 会在开发环境提示。

## 权限模式

- 当前产品默认保持 `VITE_ACCESS_MODE=frontend`；除非出现多租户、运营后台动态配置菜单或大量角色差异，不切换到 backend 菜单模式。
- 路由 `meta.roles` 控制页面访问；登录用户的 `buttons` 权限码控制按钮级能力。
- 前端模式下按钮显示优先使用 `useAuth().hasAuth()`，它会响应式读取 `userStore.info.buttons`。
- `v-roles` 继续用于角色级 DOM 显示；`v-auth` 仅适用于 backend 菜单返回的 `meta.authList`，不要在当前 frontend 模式下把它作为按钮权限主路径。

## 暗色模式与主题

- 新组件必须在 light、dark、system theme 下可读。
- 颜色使用现有主题变量或 `AppConfig.systemMainColor`，避免单独写死大面积色值。
- 图表必须跟随主题变化，优先使用 `web/src/hooks/core/useChart.ts` 和 `web/src/plugins/echarts.ts` 的模式。

## 图表与数据展示

- ECharts 只按需注册实际使用的 chart/component。
- 图表配色遵循 Art Design Pro 的蓝色主色、成功/警告/危险状态色和暗色背景对比。
- 表格、筛选、分页和空状态优先使用现有 `ArtTable`、Element Plus 和项目 API client。

## 交付检查

- `pnpm lint`
- `pnpm build`
- 手动检查中英文切换、dark mode、刷新路由、图表 resize、空数据和错误状态。
