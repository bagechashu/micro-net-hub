/**
 * 导航站点收藏：只保存在当前浏览器的 localStorage，不需要登录、不经过后端。
 * 存的是站点 ID 列表（按收藏先后排序），展示时从导航数据里查站点，站点被删除后自动忽略。
 * 所有组件共用同一份状态；其他标签页修改时通过 storage 事件同步。
 */
import { ref } from 'vue';

const STORAGE_KEY = 'sitenav-bookmarks';

function read(): number[] {
  try {
    const v = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
    return Array.isArray(v) ? v.filter((id): id is number => typeof id === 'number') : [];
  } catch {
    return [];
  }
}

function write(ids: number[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
  } catch {
    // 隐私模式等存储不可用时只在本次页面内生效
  }
}

const bookmarkIds = ref<number[]>(read());

if (typeof window !== 'undefined') {
  window.addEventListener('storage', e => {
    if (e.key === STORAGE_KEY) bookmarkIds.value = read();
  });
}

export function useBookmarks() {
  function isBookmarked(id?: number) {
    return id != null && bookmarkIds.value.includes(id);
  }

  /** 切换收藏状态，返回切换后是否已收藏 */
  function toggleBookmark(id?: number): boolean {
    if (id == null) return false;
    const exists = bookmarkIds.value.includes(id);
    bookmarkIds.value = exists ? bookmarkIds.value.filter(x => x !== id) : [...bookmarkIds.value, id];
    write(bookmarkIds.value);
    return !exists;
  }

  /** 按收藏顺序取出站点；allSites 里找不到的（已删除）跳过 */
  function pickBookmarked<T extends { ID?: number }>(allSites: T[]): T[] {
    const byId = new Map<number, T>();
    for (const s of allSites) if (s.ID != null) byId.set(s.ID, s);
    return bookmarkIds.value.map(id => byId.get(id)).filter((s): s is T => !!s);
  }

  return { bookmarkIds, isBookmarked, toggleBookmark, pickBookmarked };
}
