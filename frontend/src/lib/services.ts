export type ServiceDef = {
  id: string;
  label: string;
  image: string;
  versions: string[];
  defaultPorts: string[];
  defaultEnv: Record<string, string>;
  initials: string;
  colorClass: string;
};

export const SERVICE_CATALOG: ServiceDef[] = [
  {
    id: 'mysql',
    label: 'MySQL',
    image: 'mysql',
    versions: ['5.5.62', '5.6.51', '5.7.43', '8.0.35', '8.3'],
    defaultPorts: ['3306:3306'],
    defaultEnv: { MYSQL_ROOT_PASSWORD: 'vulnkit', MYSQL_DATABASE: 'testdb' },
    initials: 'My',
    colorClass: 'svc-blue',
  },
  {
    id: 'apache',
    label: 'Apache',
    image: 'httpd',
    versions: ['2.2.34', '2.4.49', '2.4.51', '2.4.57', '2.4.62'],
    defaultPorts: ['8080:80'],
    defaultEnv: {},
    initials: 'Ap',
    colorClass: 'svc-coral',
  },
  {
    id: 'postgres',
    label: 'PostgreSQL',
    image: 'postgres',
    versions: ['9.6', '10.23', '11.22', '12.19', '13.15', '14.12', '15.7', '16.3'],
    defaultPorts: ['5432:5432'],
    defaultEnv: { POSTGRES_PASSWORD: 'vulnkit', POSTGRES_DB: 'testdb' },
    initials: 'Pg',
    colorClass: 'svc-teal',
  },
  {
    id: 'nginx',
    label: 'Nginx',
    image: 'nginx',
    versions: ['1.14.2', '1.18.0', '1.20.2', '1.24.0', '1.26.0'],
    defaultPorts: ['8081:80'],
    defaultEnv: {},
    initials: 'Nx',
    colorClass: 'svc-green',
  },
  {
    id: 'redis',
    label: 'Redis',
    image: 'redis',
    versions: ['4.0.14', '5.0.14', '6.0.20', '6.2.14', '7.0.15', '7.2.4'],
    defaultPorts: ['6379:6379'],
    defaultEnv: {},
    initials: 'Re',
    colorClass: 'svc-amber',
  },
  {
    id: 'mongodb',
    label: 'MongoDB',
    image: 'mongo',
    versions: ['3.6', '4.0', '4.2', '4.4', '5.0', '6.0', '7.0'],
    defaultPorts: ['27017:27017'],
    defaultEnv: { MONGO_INITDB_ROOT_USERNAME: 'root', MONGO_INITDB_ROOT_PASSWORD: 'vulnkit' },
    initials: 'Mo',
    colorClass: 'svc-purple',
  },
];

export const TAG_COLORS: Record<string, string> = {
  SQLi: 'tag-blue',
  XSS:  'tag-pink',
  RCE:  'tag-coral',
  LFI:  'tag-amber',
  SSRF: 'tag-teal',
};
