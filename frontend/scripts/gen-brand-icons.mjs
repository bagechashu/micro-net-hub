/**
 * 从 simple-icons（CC0）中挑选内网/运维常用的品牌 Logo，生成 src/icons/brands.ts。
 * 只打包这里列出的图标，避免把 3000+ 个图标全部打进前端。
 * 增删品牌：修改下面的 SLUGS 后运行 `npm run gen:brands`。slug 见 https://simpleicons.org
 */
import { writeFileSync } from 'node:fs';
import * as icons from 'simple-icons';

const SLUGS = [
  // 代码与协作
  'gitlab', 'github', 'gitea', 'bitbucket', 'git', 'jira', 'confluence', 'atlassian', 'redmine', 'notion',
  'gitbook', 'bookstack', 'wikidotjs', 'outline', 'discourse', 'mattermost', 'rocketdotchat', 'wechat',
  'qq', 'zoom', 'onlyoffice', 'nextcloud', 'owncloud', 'seafile', 
  // CI/CD 与制品
  'jenkins', 'gitlab', 'argo', 'drone', 'teamcity', 'sonarqubeserver', 'sonatype', 'harbor', 'jfrog', 'npm', 'pypi',
  'docker', 'kubernetes', 'helm', 'rancher', 'portainer', 'podman', 'redhatopenshift', 'k3s',
  // 运维与基础设施
  'ansible', 'terraform', 'vault', 'consul', 'nomad', 'packer', 'saltproject', 'puppet', 'chef',
  'nginx', 'apache', 'apachetomcat', 'traefikproxy', 'caddy', 'haproxy', 'kong', 'envoyproxy', 'istio',
  'vmware', 'proxmox', 'openstack', 'truenas', 'synology', 'qnap', 'ceph', 'minio', 'linux', 'ubuntu', 'debian',
  'centos', 'redhat', 'rockylinux', 'almalinux', 'apple',
  // 网络与安全
  'cloudflare', 'wireguard', 'openvpn', 'pfsense', 'opnsense', 'ubiquiti', 'mikrotik', 'cisco', 'huawei',
  'pihole', 'adguard', 'keycloak', 'authelia', 'okta', 'auth0', 'openid', 'letsencrypt', 'bitwarden', 'vaultwarden',
  // 监控与日志
  'grafana', 'prometheus', 'sentry', 'elastic', 'elasticsearch', 'kibana', 'logstash',
  'opentelemetry', 'jaeger', 'datadog', 'newrelic', 'uptimekuma', 'netdata', 'influxdb', 'victoriametrics',
  'graylog', 'splunk',
  // 数据与中间件
  'mysql', 'mariadb', 'postgresql', 'mongodb', 'redis', 'sqlite', 'clickhouse', 'apachekafka', 'rabbitmq',
  'apachepulsar', 'apacherocketmq', 'natsdotio', 'etcd', 'apachehadoop', 'apachespark',
  'apacheflink', 'apacheairflow', 'apachesuperset', 'metabase', 'jupyter', 'apachedoris', 'tidb',
  // 云厂商
  'alibabacloud', 'baidu', 'googlecloud',
  'digitalocean',
  // 开发相关
  'swagger', 'postman', 'apifox', 'spring', 'springboot', 'go', 'python', 'nodedotjs', 'openjdk',
  'php', 'rust', 'vuedotjs', 'react', 'jetbrains', 'ollama', 'huggingface'
];

const bySlug = new Map(Object.values(icons).filter(i => i && i.slug).map(i => [i.slug, i]));
const picked = [];
const missing = [];
for (const slug of [...new Set(SLUGS)]) {
  const icon = bySlug.get(slug);
  if (icon) picked.push({ slug, title: icon.title, hex: icon.hex, path: icon.path });
  else missing.push(slug);
}

const out = `/* 由 scripts/gen-brand-icons.mjs 生成，请勿手改。图标来源 simple-icons（CC0 1.0） */
export interface BrandIcon {
  slug: string;
  title: string;
  /** 品牌色，不带 # */
  hex: string;
  /** 24x24 viewBox 下的 SVG path */
  path: string;
}

export const brandIcons: BrandIcon[] = ${JSON.stringify(picked, null, 2)};

export const brandIconMap: Record<string, BrandIcon> = Object.fromEntries(brandIcons.map(i => [i.slug, i]));
`;
writeFileSync(new URL('../src/icons/brands.ts', import.meta.url), out);
console.log(`brands: ${picked.length} picked, ${missing.length} missing: ${missing.join(' ')}`);
