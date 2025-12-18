const modules = import.meta.glob('/src/icons/svg/*.svg', { eager: true });
const icons: string[] = Object.keys(modules).map((p: string) => {
  const m = p.match(/([^\/]+)\.svg$/);
  return m ? m[1] : p;
});
export default icons;
