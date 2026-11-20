import { defineStore } from 'pinia';

/** system：跟随系统；light：亮色；dark：暗色 */
export type ThemeMode = 'system' | 'light' | 'dark';

const STORAGE_KEY = 'theme-mode';
const mediaQuery = typeof window !== 'undefined' && window.matchMedia
  ? window.matchMedia('(prefers-color-scheme: dark)')
  : undefined;

function readMode(): ThemeMode {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    if (v === 'light' || v === 'dark' || v === 'system') return v;
  } catch {
    // 隐私模式等情况下 localStorage 不可用，使用默认值
  }
  return 'system';
}

function saveMode(mode: ThemeMode) {
  try {
    localStorage.setItem(STORAGE_KEY, mode);
  } catch {
    // 忽略写入失败，只影响刷新后是否记住选择
  }
}

/** Element Plus 的暗色变量挂在 html.dark 上 */
function applyDark(isDark: boolean) {
  const root = document.documentElement;
  root.classList.toggle('dark', isDark);
  root.style.colorScheme = isDark ? 'dark' : 'light';
}

const useThemeStore = defineStore('theme', {
  state: () => ({
    mode: readMode() as ThemeMode,
    systemDark: !!mediaQuery?.matches
  }),
  getters: {
    isDark: (state): boolean => state.mode === 'dark' || (state.mode === 'system' && state.systemDark)
  },
  actions: {
    setMode(mode: ThemeMode) {
      this.mode = mode;
      saveMode(mode);
      applyDark(this.isDark);
    },
    /** 应用启动时调用：应用已保存的选择，并在“跟随系统”时监听系统主题变化 */
    init() {
      applyDark(this.isDark);
      mediaQuery?.addEventListener('change', e => {
        this.systemDark = e.matches;
        applyDark(this.isDark);
      });
    }
  }
});

export default useThemeStore;
