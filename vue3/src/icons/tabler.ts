/**
 * Tabler 图标（https://tabler.io/icons）。
 *
 * 1. 按钮等处原先使用 @element-plus/icons-vue 的图标，这里导出同名组件，内部渲染对应的 Tabler 图标，
 *    页面只需把 import 来源换成 '@/icons/tabler'。
 * 2. menuIcon()：后端菜单里存的是旧图标名（user、component、tree…），映射到 Tabler 图标。
 */
import { defineComponent, h } from 'vue';
import type { Component } from 'vue';
import {
  IconApi,
  IconArrowLeft,
  IconArrowsExchange,
  IconChevronDown,
  IconCopy,
  IconDeviceDesktop,
  IconDownload,
  IconExternalLink,
  IconFileAnalytics,
  IconFileText,
  IconHistory,
  IconHome,
  IconLanguage,
  IconListDetails,
  IconLock,
  IconLogout,
  IconMenu2,
  IconMoon,
  IconPencil,
  IconPlus,
  IconRefresh,
  IconSearch,
  IconSettings,
  IconShieldLock,
  IconSitemap,
  IconSpeakerphone,
  IconStar,
  IconSun,
  IconTrash,
  IconUpload,
  IconUser,
  IconUserCircle,
  IconUsers,
  IconUsersGroup,
  IconWorldWww
} from '@tabler/icons-vue';

/** 统一尺寸（随字号缩放）与线宽，便于放进 el-icon / el-button 的 icon 属性 */
function tabler(icon: Component, name: string) {
  return defineComponent({
    name,
    inheritAttrs: false,
    setup(_, { attrs }) {
      return () => h(icon as any, { size: '1em', stroke: 1.75, ...attrs });
    }
  });
}

// ---- 与 @element-plus/icons-vue 同名的导出 ----
export const ArrowDown = tabler(IconChevronDown, 'ArrowDown');
export const ArrowLeft = tabler(IconArrowLeft, 'ArrowLeft');
export const CopyDocument = tabler(IconCopy, 'CopyDocument');
export const Delete = tabler(IconTrash, 'Delete');
export const Document = tabler(IconFileText, 'Document');
export const Download = tabler(IconDownload, 'Download');
export const Edit = tabler(IconPencil, 'Edit');
export const Lock = tabler(IconLock, 'Lock');
export const Monitor = tabler(IconDeviceDesktop, 'Monitor');
export const Moon = tabler(IconMoon, 'Moon');
export const Plus = tabler(IconPlus, 'Plus');
export const Position = tabler(IconExternalLink, 'Position');
export const Refresh = tabler(IconRefresh, 'Refresh');
export const Search = tabler(IconSearch, 'Search');
export const Setting = tabler(IconSettings, 'Setting');
export const StarFilled = tabler(IconStar, 'StarFilled');
export const Sunny = tabler(IconSun, 'Sunny');
export const SwitchButton = tabler(IconLogout, 'SwitchButton');
export const Upload = tabler(IconUpload, 'Upload');
export const User = tabler(IconUser, 'User');

// ---- 其他直接使用的 Tabler 图标 ----
export { IconEye, IconEyeOff, IconLanguage, IconLock, IconUser } from '@tabler/icons-vue';

// ---- 菜单图标：旧图标名 -> Tabler ----
const menuIcons: Record<string, Component> = {
  home: IconHome, // 网站导航（首页）
  user: IconUsers, // 人员管理
  people: IconUser, // 用户管理
  peoples: IconUsersGroup, // 分组管理
  relation: IconArrowsExchange, // 字段关系管理
  'el-icon-s-tools': IconArrowsExchange, // 字段关系管理（旧数据里的 Element UI 图标名）
  component: IconSettings, // 系统管理
  'eye-open': IconShieldLock, // 角色管理
  'tree-table': IconMenu2, // 菜单管理
  tree: IconApi, // 接口管理
  list: IconListDetails, // 导航配置
  guide: IconWorldWww, // DNS 配置
  message: IconSpeakerphone, // 公告管理
  example: IconFileAnalytics, // 日志管理
  documentation: IconHistory, // 操作日志
  profile: IconUserCircle, // 个人中心
  language: IconLanguage,
  sitemap: IconSitemap
};

/** 菜单图标：有对应的 Tabler 图标时返回组件，否则返回 undefined（调用方回退到 svg-icon） */
export function menuIcon(name?: string): Component | undefined {
  return name ? menuIcons[name] : undefined;
}
