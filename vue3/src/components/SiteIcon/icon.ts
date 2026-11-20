/**
 * 站点图标（Site Navigation 里每个站点的 icon 字段）。
 *
 * 后端 icon 字段是一个最长 100 的字符串，这里约定几种写法（旧数据的图片路径照常可用）：
 *   si:<slug>            品牌 Logo（simple-icons 精选，见 src/icons/brands.ts）
 *   ti:<name>[:<color>]  Tabler 通用图标 + 颜色（颜色见 ICON_COLORS）
 *   favicon              自动使用站点自己的 favicon（取链接所在域名的 /favicon.ico）
 *   http(s)://…          图片地址
 *   空                   显示站点名称首字
 */
import type { Component } from 'vue';
import {
  IconActivity,
  IconApi,
  IconArchive,
  IconBell,
  IconBolt,
  IconBook,
  IconBooks,
  IconBox,
  IconBrackets,
  IconBrain,
  IconBug,
  IconBuilding,
  IconCalendar,
  IconCamera,
  IconChartBar,
  IconChartLine,
  IconChartPie,
  IconChecklist,
  IconClipboardList,
  IconCloud,
  IconCloudUpload,
  IconCode,
  IconCoin,
  IconContainer,
  IconCpu,
  IconCreditCard,
  IconCube,
  IconDatabase,
  IconDatabaseExport,
  IconDeviceDesktop,
  IconDeviceMobile,
  IconFileCode,
  IconFileText,
  IconFingerprint,
  IconFlag,
  IconFlame,
  IconFolder,
  IconGauge,
  IconGitBranch,
  IconGitMerge,
  IconHeadset,
  IconHeartRateMonitor,
  IconHeartbeat,
  IconHome,
  IconKey,
  IconLayoutDashboard,
  IconLifebuoy,
  IconLink,
  IconLock,
  IconMail,
  IconMap,
  IconMessage,
  IconMicrophone,
  IconMusic,
  IconNetwork,
  IconNotebook,
  IconPackage,
  IconPhone,
  IconPhoto,
  IconPrinter,
  IconReceipt,
  IconReport,
  IconRobot,
  IconRocket,
  IconRouter,
  IconSchool,
  IconSearch,
  IconServer,
  IconServer2,
  IconServerBolt,
  IconServerCog,
  IconSettings,
  IconShield,
  IconShieldLock,
  IconShoppingCart,
  IconSitemap,
  IconSparkles,
  IconStack2,
  IconStar,
  IconTable,
  IconTerminal2,
  IconTicket,
  IconTimeline,
  IconTool,
  IconTruck,
  IconUser,
  IconUserShield,
  IconUsers,
  IconVideo,
  IconWallet,
  IconWifi,
  IconWorld,
  IconWorldWww
} from '@tabler/icons-vue';

export type SiteIconValue =
  | { kind: 'brand'; slug: string }
  | { kind: 'tabler'; name: string; color: string }
  | { kind: 'favicon' }
  | { kind: 'image'; url: string }
  | { kind: 'none' };

/** Tabler 配色 */
export const ICON_COLORS: Record<string, string> = {
  blue: '#066fd1',
  azure: '#4299e1',
  indigo: '#4263eb',
  purple: '#ae3ec9',
  pink: '#d6336c',
  red: '#d63939',
  orange: '#f76707',
  yellow: '#f59f00',
  lime: '#74b816',
  green: '#2fb344',
  teal: '#0ca678',
  cyan: '#17a2b8',
  gray: '#667382'
};
export const DEFAULT_ICON_COLOR = 'blue';

export interface GenericIcon {
  name: string;
  /** 中文关键词，便于搜索 */
  keywords: string;
  component: Component;
}

export const genericIcons: GenericIcon[] = [
  { name: 'server', keywords: '服务器', component: IconServer },
  { name: 'server-2', keywords: '服务器 主机', component: IconServer2 },
  { name: 'server-cog', keywords: '服务器 配置', component: IconServerCog },
  { name: 'server-bolt', keywords: '服务器 电源', component: IconServerBolt },
  { name: 'database', keywords: '数据库', component: IconDatabase },
  { name: 'database-export', keywords: '数据库 导出 备份', component: IconDatabaseExport },
  { name: 'cloud', keywords: '云', component: IconCloud },
  { name: 'cloud-upload', keywords: '云 上传', component: IconCloudUpload },
  { name: 'network', keywords: '网络', component: IconNetwork },
  { name: 'router', keywords: '路由器 网关', component: IconRouter },
  { name: 'wifi', keywords: '无线 wifi', component: IconWifi },
  { name: 'world', keywords: '网站 外网', component: IconWorld },
  { name: 'world-www', keywords: '网站 域名', component: IconWorldWww },
  { name: 'sitemap', keywords: '拓扑 架构', component: IconSitemap },
  { name: 'cpu', keywords: 'CPU 计算', component: IconCpu },
  { name: 'container', keywords: '容器', component: IconContainer },
  { name: 'box', keywords: '盒子 制品', component: IconBox },
  { name: 'package', keywords: '包 制品 仓库', component: IconPackage },
  { name: 'stack-2', keywords: '堆栈 中间件', component: IconStack2 },
  { name: 'cube', keywords: '立方体 模块', component: IconCube },
  { name: 'terminal-2', keywords: '终端 堡垒机 ssh', component: IconTerminal2 },
  { name: 'code', keywords: '代码', component: IconCode },
  { name: 'brackets', keywords: '代码 接口', component: IconBrackets },
  { name: 'file-code', keywords: '代码 文件', component: IconFileCode },
  { name: 'git-branch', keywords: '分支 git', component: IconGitBranch },
  { name: 'git-merge', keywords: '合并 git', component: IconGitMerge },
  { name: 'bug', keywords: '缺陷 bug', component: IconBug },
  { name: 'api', keywords: '接口 api', component: IconApi },
  { name: 'rocket', keywords: '发布 部署', component: IconRocket },
  { name: 'bolt', keywords: '快速 闪电', component: IconBolt },
  { name: 'robot', keywords: '机器人 自动化', component: IconRobot },
  { name: 'brain', keywords: 'AI 智能', component: IconBrain },
  { name: 'sparkles', keywords: 'AI 魔法', component: IconSparkles },
  { name: 'shield', keywords: '安全', component: IconShield },
  { name: 'shield-lock', keywords: '安全 加密', component: IconShieldLock },
  { name: 'lock', keywords: '锁 权限', component: IconLock },
  { name: 'key', keywords: '密钥 密码', component: IconKey },
  { name: 'fingerprint', keywords: '认证 指纹', component: IconFingerprint },
  { name: 'user-shield', keywords: '账号 安全', component: IconUserShield },
  { name: 'activity', keywords: '监控 活动', component: IconActivity },
  { name: 'heartbeat', keywords: '监控 健康', component: IconHeartbeat },
  { name: 'heart-rate-monitor', keywords: '监控 大屏', component: IconHeartRateMonitor },
  { name: 'gauge', keywords: '仪表盘 性能', component: IconGauge },
  { name: 'chart-bar', keywords: '图表 报表', component: IconChartBar },
  { name: 'chart-line', keywords: '图表 趋势', component: IconChartLine },
  { name: 'chart-pie', keywords: '图表 统计', component: IconChartPie },
  { name: 'layout-dashboard', keywords: '仪表盘 看板', component: IconLayoutDashboard },
  { name: 'timeline', keywords: '时间线 日志', component: IconTimeline },
  { name: 'bell', keywords: '告警 通知', component: IconBell },
  { name: 'report', keywords: '报告', component: IconReport },
  { name: 'book', keywords: '文档 手册', component: IconBook },
  { name: 'books', keywords: '知识库 wiki', component: IconBooks },
  { name: 'notebook', keywords: '笔记', component: IconNotebook },
  { name: 'file-text', keywords: '文档 文件', component: IconFileText },
  { name: 'folder', keywords: '文件夹 网盘', component: IconFolder },
  { name: 'archive', keywords: '归档', component: IconArchive },
  { name: 'table', keywords: '表格', component: IconTable },
  { name: 'clipboard-list', keywords: '清单 工单', component: IconClipboardList },
  { name: 'checklist', keywords: '检查 任务', component: IconChecklist },
  { name: 'ticket', keywords: '工单', component: IconTicket },
  { name: 'calendar', keywords: '日历 会议', component: IconCalendar },
  { name: 'mail', keywords: '邮件 邮箱', component: IconMail },
  { name: 'message', keywords: '消息 聊天 IM', component: IconMessage },
  { name: 'headset', keywords: '客服 支持', component: IconHeadset },
  { name: 'lifebuoy', keywords: '帮助 支持', component: IconLifebuoy },
  { name: 'school', keywords: '培训 学习', component: IconSchool },
  { name: 'users', keywords: '用户 团队', component: IconUsers },
  { name: 'user', keywords: '用户 个人', component: IconUser },
  { name: 'home', keywords: '首页 门户', component: IconHome },
  { name: 'building', keywords: '公司 办公', component: IconBuilding },
  { name: 'settings', keywords: '设置 配置', component: IconSettings },
  { name: 'tool', keywords: '工具', component: IconTool },
  { name: 'search', keywords: '搜索', component: IconSearch },
  { name: 'link', keywords: '链接 短链', component: IconLink },
  { name: 'photo', keywords: '图片 相册', component: IconPhoto },
  { name: 'video', keywords: '视频 直播', component: IconVideo },
  { name: 'music', keywords: '音乐 音频', component: IconMusic },
  { name: 'camera', keywords: '摄像头 监控', component: IconCamera },
  { name: 'microphone', keywords: '麦克风 会议', component: IconMicrophone },
  { name: 'phone', keywords: '电话', component: IconPhone },
  { name: 'printer', keywords: '打印机', component: IconPrinter },
  { name: 'map', keywords: '地图', component: IconMap },
  { name: 'truck', keywords: '物流', component: IconTruck },
  { name: 'shopping-cart', keywords: '商城 采购', component: IconShoppingCart },
  { name: 'wallet', keywords: '钱包 财务', component: IconWallet },
  { name: 'coin', keywords: '费用 财务', component: IconCoin },
  { name: 'credit-card', keywords: '支付', component: IconCreditCard },
  { name: 'receipt', keywords: '报销 发票', component: IconReceipt },
  { name: 'device-desktop', keywords: '桌面 电脑', component: IconDeviceDesktop },
  { name: 'device-mobile', keywords: '手机 移动端', component: IconDeviceMobile },
  { name: 'flame', keywords: '热点', component: IconFlame },
  { name: 'star', keywords: '收藏', component: IconStar },
  { name: 'flag', keywords: '标记 里程碑', component: IconFlag }
];

export const genericIconMap: Record<string, GenericIcon> = Object.fromEntries(genericIcons.map(i => [i.name, i]));

export function parseSiteIcon(raw?: string | null): SiteIconValue {
  const v = (raw || '').trim();
  if (!v) return { kind: 'none' };
  if (v === 'favicon') return { kind: 'favicon' };
  if (v.startsWith('si:')) return { kind: 'brand', slug: v.slice(3) };
  if (v.startsWith('ti:')) {
    const [name, color] = v.slice(3).split(':');
    return { kind: 'tabler', name, color: color && ICON_COLORS[color] ? color : DEFAULT_ICON_COLOR };
  }
  return { kind: 'image', url: v };
}

export function formatSiteIcon(v: SiteIconValue): string {
  switch (v.kind) {
    case 'brand':
      return `si:${v.slug}`;
    case 'tabler':
      return v.color && v.color !== DEFAULT_ICON_COLOR ? `ti:${v.name}:${v.color}` : `ti:${v.name}`;
    case 'favicon':
      return 'favicon';
    case 'image':
      return v.url;
    default:
      return '';
  }
}

/** 图片地址只允许 http(s) 绝对地址，挡住 javascript:/data: 等 */
export function isSafeImageUrl(value: string): boolean {
  try {
    const u = new URL((value || '').trim());
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch {
    return false;
  }
}

/** icon 字段的合法性：空、约定前缀或安全的图片地址 */
export function isValidSiteIcon(raw?: string | null): boolean {
  const p = parseSiteIcon(raw);
  if (p.kind === 'image') return isSafeImageUrl(p.url);
  if (p.kind === 'tabler') return !!genericIconMap[p.name];
  if (p.kind === 'brand') return /^[a-z0-9]+$/.test(p.slug);
  return true;
}

/** 站点链接的 favicon 地址（站点自己的 /favicon.ico） */
export function faviconOf(link?: string): string {
  try {
    const u = new URL(link || '');
    if (u.protocol !== 'http:' && u.protocol !== 'https:') return '';
    return `${u.origin}/favicon.ico`;
  } catch {
    return '';
  }
}

/** 名称首字（中文取第一个字，英文取首字母大写）与按名称固定的配色 */
export function initialOf(name?: string): string {
  const n = (name || '').trim();
  return n ? n.slice(0, 1).toUpperCase() : '?';
}

export function colorOfName(name?: string): string {
  const keys = Object.keys(ICON_COLORS).filter(k => k !== 'gray');
  let h = 0;
  for (const ch of name || '') h = (h * 31 + ch.charCodeAt(0)) >>> 0;
  return ICON_COLORS[keys[h % keys.length]];
}

/** 颜色是否偏暗（暗色模式下品牌 Logo 需要反白） */
export function isDarkColor(hex: string): boolean {
  const m = /^#?([0-9a-f]{6})$/i.exec(hex);
  if (!m) return false;
  const n = parseInt(m[1], 16);
  const r = (n >> 16) & 255, g = (n >> 8) & 255, b = n & 255;
  return 0.299 * r + 0.587 * g + 0.114 * b < 60;
}

// 品牌图标数据较大（约 100KB gzip），按需异步加载并缓存
let brandsPromise: Promise<typeof import('@/icons/brands')> | null = null;
export function loadBrands() {
  if (!brandsPromise) brandsPromise = import('@/icons/brands');
  return brandsPromise;
}
