# UI 配色规范

## 基本原则

- **使用纯色背景**，不使用渐变色（`linear-gradient`）
- **科技风配色**，偏好深色系
- **避免灰色系**，不使用沉闷的灰色调

## 主色调

### 科技蓝主题

```css
/* 主色 */
--primary-color: #1a237e;

/* 用于 */
- 登录页背景
- 侧边栏背景
- 按钮背景
- 输入框聚焦态边框
```

## 配色方案

### 登录页面

```css
/* 背景 */
background: #1a237e;

/* 登录按钮 */
background: #1a237e;

/* 输入框聚焦 */
border-color: #1a237e;
box-shadow: 0 0 0 2px rgba(26, 35, 126, 0.1);
```

### 布局页面

```css
/* 侧边栏 */
background: #1a237e;

/* 菜单文字 */
color: rgba(255, 255, 255, 0.7);

/* 菜单激活项 */
color: #fff;
background: rgba(255, 255, 255, 0.15);

/* 菜单悬停 */
background: rgba(255, 255, 255, 0.1);

/* Logo 区域 */
background: rgba(0, 0, 0, 0.2);
```

## 禁止使用的配色

- ❌ 灰色系：`#545c64`、`#304156`、`#263445` 等
- ❌ 渐变色：`linear-gradient(...)`
- ❌ 默认 Element UI 蓝色：`#409EFF`

## 品牌名称

- **正式名称**：Kenaito配置中心
- **使用场景**：登录页标题、侧边栏 Logo、页面标题、欢迎语等
- **不要简写**：不使用"配置中心"

## 参考文件

- 登录页：`web/src/views/Login.vue`
- 布局页：`web/src/views/Layout.vue`
