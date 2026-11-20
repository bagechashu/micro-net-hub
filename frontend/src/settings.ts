interface ISettings {
  title: string;
}

// 默认站点名称；实际显示的名称由后端 config.yml 的 system.site-name 提供
const settings: ISettings = {
  title: 'Micro Net Hub'
};

export default settings;
