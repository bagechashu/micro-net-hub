import defaultSettings from '@/settings';

const title = defaultSettings.title || 'Micro Net Hub';

export default function getPageTitle(pageTitle) {
  if (pageTitle) {
    return `${pageTitle} - ${title}`;
  }
  return `${title}`;
}
